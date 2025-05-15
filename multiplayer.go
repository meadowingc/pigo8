package pigo8

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"sync"
	"time"
)

// --- Multiplayer Settings ---

// MultiplayerSettings defines configuration for multiplayer games
// Deprecated: Use NetworkConfig from network.go instead.
type MultiplayerSettings struct {
	Enabled      bool   // Whether multiplayer is enabled
	IsServer     bool   // Whether this instance is a server (host) or client
	ServerIP     string // IP address of the server (for clients)
	Port         int    // Port to use for connection
	GameName     string // Name of the game for display
	SyncInterval int    // Milliseconds between state synchronizations
}

// DefaultMultiplayerSettings returns default multiplayer settings
// Deprecated: Use DefaultNetworkConfig from network.go instead.
func DefaultMultiplayerSettings() *MultiplayerSettings {
	return &MultiplayerSettings{
		Enabled:      false,
		IsServer:     true,
		ServerIP:     "localhost",
		Port:         8080,
		GameName:     "PIGO8 Game",
		SyncInterval: 16, // ~60 updates per second
	}
}

// --- State Registration and Synchronization ---

// StateField represents a field in the game state that should be synchronized
type StateField struct {
	Name      string
	FieldType reflect.Type
	Value     interface{}
	Path      []int // Index path to the field in the struct
}

// RegisteredState holds information about a registered game state
type RegisteredState struct {
	Object        interface{}
	Fields        []StateField
	LastSync      time.Time
	SyncInterval  time.Duration
	mutex         sync.RWMutex
	serverUpdates chan []byte
	clientUpdates chan []byte
}

var (
	registeredState *RegisteredState
	stateMutex      sync.RWMutex
	multiSettings   *MultiplayerSettings
)

// ParseMultiplayerArgs parses command line arguments for multiplayer settings
// This allows games to enable multiplayer through command line flags without
// having to implement their own argument parsing
// Deprecated: Use ParseNetworkArgs from network.go instead.
func ParseMultiplayerArgs() *MultiplayerSettings {
	settings := DefaultMultiplayerSettings()
	
	// Check for command line arguments to enable multiplayer
	args := os.Args
	
	// Simple command line parsing
	for i, arg := range args {
		if arg == "--multiplayer" || arg == "-m" {
			settings.Enabled = true
		}
		if arg == "--client" || arg == "-c" {
			settings.IsServer = false
		}
		if arg == "--connect" && i+1 < len(args) {
			settings.ServerIP = args[i+1]
		}
		if arg == "--port" && i+1 < len(args) {
			if port, err := strconv.Atoi(args[i+1]); err == nil {
				settings.Port = port
			}
		}
		if arg == "--name" && i+1 < len(args) {
			settings.GameName = args[i+1]
		}
	}
	
	return settings
}

// InitMultiplayer initializes multiplayer functionality with the given settings
// Deprecated: Use InitNetwork from network.go instead.
func InitMultiplayer(settings *MultiplayerSettings) error {
	if settings == nil {
		settings = DefaultMultiplayerSettings()
	}
	
	multiSettings = settings
	
	if !settings.Enabled {
		return nil // Multiplayer disabled, nothing to do
	}
	
	// Configure network settings based on multiplayer settings
	netConfig := DefaultNetworkConfig()
	if settings.IsServer {
		netConfig.Role = RoleServer
	} else {
		netConfig.Role = RoleClient
		netConfig.Address = settings.ServerIP
	}
	netConfig.Port = settings.Port
	netConfig.GameName = settings.GameName
	
	// Initialize networking
	if err := InitNetwork(netConfig); err != nil {
		return fmt.Errorf("failed to initialize networking: %v", err)
	}
	
	// Set up callbacks for state synchronization
	SetOnGameStateCallback(handleStateSync)
	
	return nil
}

// RegisterGameState registers a game state object for automatic synchronization
// The object should be a pointer to a struct containing the game state
// Deprecated: Use the network.go API with explicit state synchronization instead.
func RegisterGameState(stateObj interface{}) error {
	stateMutex.Lock()
	defer stateMutex.Unlock()
	
	if registeredState != nil {
		return fmt.Errorf("a game state is already registered")
	}
	
	// Verify that stateObj is a pointer to a struct
	objType := reflect.TypeOf(stateObj)
	if objType.Kind() != reflect.Ptr {
		return fmt.Errorf("state object must be a pointer")
	}
	
	elemType := objType.Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("state object must be a pointer to a struct")
	}
	
	// Create the registered state
	state := &RegisteredState{
		Object:       stateObj,
		Fields:       discoverStateFields(elemType, nil),
		LastSync:     time.Now(),
		SyncInterval: time.Duration(multiSettings.SyncInterval) * time.Millisecond,
		serverUpdates: make(chan []byte, 10),
		clientUpdates: make(chan []byte, 10),
	}
	
	registeredState = state
	
	// Start the synchronization goroutine
	go syncGameState()
	
	return nil
}

// discoverStateFields recursively discovers all fields in a struct type
// that should be synchronized in multiplayer
func discoverStateFields(t reflect.Type, path []int) []StateField {
	var fields []StateField
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if field.PkgPath != "" {
			continue
		}
		
		// Check for "nosync" tag to exclude fields from synchronization
		if field.Tag.Get("nosync") == "true" {
			continue
		}
		
		// Build the path to this field
		fieldPath := append(append([]int{}, path...), i)
		
		// Handle nested structs
		if field.Type.Kind() == reflect.Struct {
			nestedFields := discoverStateFields(field.Type, fieldPath)
			fields = append(fields, nestedFields...)
			continue
		}
		
		// Add this field to the list
		fields = append(fields, StateField{
			Name:      field.Name,
			FieldType: field.Type,
			Path:      fieldPath,
		})
	}
	
	return fields
}

// syncGameState runs in a goroutine and handles synchronizing the game state
func syncGameState() {
	if registeredState == nil || multiSettings == nil || !multiSettings.Enabled {
		return
	}
	
	ticker := time.NewTicker(registeredState.SyncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if IsServer() {
				sendStateUpdates()
			} else if IsClient() {
				sendClientInput()
			}
		case data := <-registeredState.serverUpdates:
			if IsClient() {
				applyStateUpdate(data)
			}
		case data := <-registeredState.clientUpdates:
			if IsServer() {
				applyClientInput(data)
			}
		}
	}
}

// sendStateUpdates sends the current game state to all clients
func sendStateUpdates() {
	if registeredState == nil || !IsServer() {
		return
	}
	
	registeredState.mutex.RLock()
	stateObj := registeredState.Object
	registeredState.mutex.RUnlock()
	
	// Serialize the state
	data, err := json.Marshal(stateObj)
	if err != nil {
		log.Printf("Error serializing game state: %v", err)
		return
	}
	
	// Send to all clients
	SendGameState(data, "all")
}

// sendClientInput sends the client's input to the server
func sendClientInput() {
	if registeredState == nil || !IsClient() {
		return
	}
	
	registeredState.mutex.RLock()
	stateObj := registeredState.Object
	registeredState.mutex.RUnlock()
	
	// Serialize the client's input state
	// For simplicity, we're sending the entire state, but in a real
	// implementation you might want to send only input-related fields
	data, err := json.Marshal(stateObj)
	if err != nil {
		log.Printf("Error serializing client input: %v", err)
		return
	}
	
	// Send to server
	SendPlayerInput(data)
}

// handleStateSync processes game state updates from the network
func handleStateSync(playerID string, data []byte) {
	if registeredState == nil {
		return
	}
	
	if IsServer() {
		// Server received input from a client
		registeredState.clientUpdates <- data
	} else {
		// Client received state update from server
		registeredState.serverUpdates <- data
	}
}

// applyStateUpdate applies a state update received from the server
func applyStateUpdate(data []byte) {
	if registeredState == nil || !IsClient() {
		return
	}
	
	registeredState.mutex.Lock()
	defer registeredState.mutex.Unlock()
	
	// Deserialize the state update into the registered object
	if err := json.Unmarshal(data, registeredState.Object); err != nil {
		log.Printf("Error applying state update: %v", err)
	}
}

// applyClientInput applies input received from a client
func applyClientInput(data []byte) {
	if registeredState == nil || !IsServer() {
		return
	}
	
	// In a real implementation, you would:
	// 1. Extract only the input-related fields from the client data
	// 2. Apply those inputs to the game logic
	// 3. Update the game state based on those inputs
	
	// For simplicity in this example, we're just logging that we received input
	log.Printf("Received client input (%d bytes)", len(data))
	
	// The actual implementation would depend on your specific game's input handling
}

// IsMultiplayerEnabled returns whether multiplayer is enabled
// Deprecated: Use IsNetworkInitialized from network.go instead.
func IsMultiplayerEnabled() bool {
	return multiSettings != nil && multiSettings.Enabled
}

// ShutdownMultiplayer cleans up multiplayer resources
// Deprecated: Use ShutdownNetwork from network.go instead.
func ShutdownMultiplayer() {
	if !IsMultiplayerEnabled() {
		return
	}
	
	ShutdownNetwork()
	
	stateMutex.Lock()
	registeredState = nil
	stateMutex.Unlock()
}

// Note: The deprecated multiplayer settings methods have been removed.
// Use the NetworkConfig struct and InitNetwork from network.go instead.
