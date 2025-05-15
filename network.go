package pigo8

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

// --- Network Types ---

// MessageType defines the type of network message being sent
type MessageType int

const (
	// MsgConnect is sent when a player connects to a game
	MsgConnect MessageType = iota
	// MsgDisconnect is sent when a player disconnects from a game
	MsgDisconnect
	// MsgGameState is sent to sync game state between players
	MsgGameState
	// MsgPlayerInput is sent when a player performs an input action
	MsgPlayerInput
	// MsgPing is sent to check connection status
	MsgPing
	// MsgPong is sent in response to a ping
	MsgPong
)

// NetworkMessage represents a message sent over the network
type NetworkMessage struct {
	Type     MessageType `json:"type"`
	PlayerID string      `json:"player_id"`
	Data     []byte      `json:"data"`
}

// NetworkRole defines whether this instance is a server or client
type NetworkRole int

const (
	// RoleServer indicates this instance is hosting the game
	RoleServer NetworkRole = iota
	// RoleClient indicates this instance is connecting to a host
	RoleClient
)

// NetworkConfig holds configuration for network functionality
type NetworkConfig struct {
	Role       NetworkRole // Whether this instance is a server or client
	Address    string      // Address to connect to (for client) or listen on (for server)
	Port       int         // Port to use for connection
	PlayerID   string      // Unique identifier for this player
	BufferSize int         // Size of message buffer
	GameName   string      // Name of the game (for display purposes)
}

// DefaultNetworkConfig returns a default network configuration
func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		Role:       RoleServer,
		Address:    "localhost",
		Port:       8080,
		PlayerID:   fmt.Sprintf("player-%d", time.Now().UnixNano()%10000),
		BufferSize: 100,
		GameName:   "PIGO8 Game",
	}
}

// --- Network Manager ---

// NetworkManager handles all networking functionality
type NetworkManager struct {
	config            *NetworkConfig
	conn              net.Conn
	listener          net.Listener
	clients           map[string]net.Conn
	incomingMsgs      chan NetworkMessage
	outgoingMsgs      chan NetworkMessage
	onConnect         func(playerID string)
	onDisconnect      func(playerID string)
	onGameState       func(playerID string, data []byte)
	onPlayerInput     func(playerID string, data []byte)
	isRunning         bool
	mutex             sync.Mutex
	connectionLost    bool
	waitingForPlayers bool
	networkError      string
}

var (
	// Global network manager instance
	networkManager *NetworkManager
	networkMutex   sync.Mutex
)

// --- Network Initialization ---

// InitNetwork initializes the networking system with the given configuration
func InitNetwork(config *NetworkConfig) error {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	// Clean up any existing network manager
	if networkManager != nil {
		ShutdownNetwork() // Use the global shutdown function
	}

	// If this is a server and address is localhost, use 0.0.0.0 to listen on all interfaces
	if config.Role == RoleServer && (config.Address == "localhost" || config.Address == "127.0.0.1") {
		log.Printf("Server will listen on all interfaces (0.0.0.0) instead of %s", config.Address)
		config.Address = "0.0.0.0"
	}

	// Create a new network manager
	networkManager = &NetworkManager{
		config:            config,
		incomingMsgs:      make(chan NetworkMessage, config.BufferSize),
		outgoingMsgs:      make(chan NetworkMessage, config.BufferSize),
		clients:           make(map[string]net.Conn),
		isRunning:         true,
		waitingForPlayers: config.Role == RoleServer, // Server starts waiting for players
	}

	// Start network processing in background
	go networkManager.processMessages()

	// Start server or client based on role
	var err error
	if config.Role == RoleServer {
		err = networkManager.startServer()
	} else {
		err = networkManager.connectToServer()
	}

	return err
}

// ShutdownNetwork closes all network connections and stops processing
func ShutdownNetwork() {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil {
		return
	}

	networkManager.mutex.Lock()
	networkManager.isRunning = false
	networkManager.mutex.Unlock()

	// Close connections
	if networkManager.listener != nil {
		if err := networkManager.listener.Close(); err != nil {
			log.Printf("Error closing listener: %v", err)
		}
	}

	if networkManager.conn != nil {
		if err := networkManager.conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}

	for _, conn := range networkManager.clients {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing client connection: %v", err)
		}
	}

	// Clear the manager
	networkManager = nil
}

// --- Server Functions ---

// startServer initializes a server that listens for client connections
func (nm *NetworkManager) startServer() error {
	addr := net.JoinHostPort(nm.config.Address, fmt.Sprintf("%d", nm.config.Port))
	log.Printf("Starting server on %s...", addr)

	// Try to listen on all interfaces if address is localhost
	if nm.config.Address == "localhost" || nm.config.Address == "127.0.0.1" {
		log.Printf("Using 0.0.0.0 instead of localhost to listen on all interfaces")
		addr = net.JoinHostPort("0.0.0.0", fmt.Sprintf("%d", nm.config.Port))
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		nm.networkError = fmt.Sprintf("Server start failed: %v", err)
		log.Printf("Failed to start server on %s: %v", addr, err)
		return fmt.Errorf("failed to start server: %v", err)
	}

	nm.listener = listener
	log.Printf("Server successfully started on %s (local IP: %s)", addr, GetLocalIP())

	// Accept connections in background
	go nm.acceptConnections()
	return nil
}

// acceptConnections handles incoming client connections
func (nm *NetworkManager) acceptConnections() {
	log.Printf("Server is now accepting connections...")
	for nm.isRunning {
		conn, err := nm.listener.Accept()
		if err != nil {
			if !nm.isRunning {
				log.Printf("Server shutting down, stopping accept loop")
				return // Server is shutting down
			}
			log.Printf("Error accepting connection: %v", err)
			nm.networkError = fmt.Sprintf("Accept error: %v", err)
			continue
		}

		// Handle the new connection in a goroutine
		go nm.handleConnection(conn)
	}
}

// handleConnection processes messages from a connected client
func (nm *NetworkManager) handleConnection(conn net.Conn) {
	// Get client address for logging
	clientAddr := conn.RemoteAddr().String()
	log.Printf("New client connection from %s", clientAddr)

	// We're now connected with at least one client
	nm.mutex.Lock()
	nm.waitingForPlayers = false
	nm.mutex.Unlock()

	// Create a decoder for incoming messages
	decoder := json.NewDecoder(conn)

	// Read the first message which should be a connect message
	var msg NetworkMessage
	if err := decoder.Decode(&msg); err != nil {
		log.Printf("Error reading connect message: %v", err)
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Error closing connection: %v", closeErr)
		}
		return
	}

	if msg.Type != MsgConnect {
		log.Printf("First message was not a connect message")
		if err := conn.Close(); err != nil {
			log.Printf("Error closing connection: %v", err)
		}
		return
	}

	// Store the connection
	playerID := msg.PlayerID
	nm.mutex.Lock()
	nm.clients[playerID] = conn
	nm.waitingForPlayers = false
	nm.mutex.Unlock()

	log.Printf("Client connected: %s", playerID)

	// Notify about the connection
	if nm.onConnect != nil {
		nm.onConnect(playerID)
	}

	// Process messages from this client
	for nm.isRunning {
		var msg NetworkMessage
		if err := decoder.Decode(&msg); err != nil {
			log.Printf("Error reading from client %s: %v", playerID, err)
			break
		}

		// Handle the message
		nm.incomingMsgs <- msg
	}

	// Client disconnected
	nm.mutex.Lock()
	delete(nm.clients, playerID)
	if len(nm.clients) == 0 {
		nm.waitingForPlayers = true
	}
	nm.mutex.Unlock()
	if err := conn.Close(); err != nil {
		log.Printf("Error closing connection after client disconnect: %v", err)
	}

	// Notify about the disconnection
	if nm.onDisconnect != nil {
		nm.onDisconnect(playerID)
	}

	log.Printf("Client disconnected: %s", playerID)
}

// --- Client Functions ---

// connectToServer connects to a game server
func (nm *NetworkManager) connectToServer() error {
	addr := net.JoinHostPort(nm.config.Address, fmt.Sprintf("%d", nm.config.Port))
	log.Printf("Attempting to connect to server at %s...", addr)

	// Set a timeout for the connection attempt
	dialer := net.Dialer{Timeout: 5 * time.Second}
	conn, err := dialer.Dial("tcp", addr)
	if err != nil {
		nm.mutex.Lock()
		nm.connectionLost = true
		nm.networkError = fmt.Sprintf("Connection failed: %v", err)
		nm.mutex.Unlock()

		log.Printf("Failed to connect to server at %s: %v", addr, err)
		return fmt.Errorf("failed to connect to server: %v", err)
	}

	nm.conn = conn
	log.Printf("Successfully connected to server at %s", addr)

	// Send connect message
	connectMsg := NetworkMessage{
		Type:     MsgConnect,
		PlayerID: nm.config.PlayerID,
		Data:     []byte{},
	}

	if err := json.NewEncoder(conn).Encode(connectMsg); err != nil {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Error closing connection after encoding error: %v", closeErr)
		}
		return fmt.Errorf("failed to send connect message: %v", err)
	}

	// Start receiving messages
	go nm.receiveMessages()
	return nil
}

// receiveMessages handles incoming messages from the server
func (nm *NetworkManager) receiveMessages() {
	decoder := json.NewDecoder(nm.conn)

	for nm.isRunning {
		var msg NetworkMessage
		if err := decoder.Decode(&msg); err != nil {
			if nm.isRunning {
				log.Printf("Error reading from server: %v", err)
				nm.mutex.Lock()
				nm.connectionLost = true
				nm.networkError = "connection to server lost"
				nm.mutex.Unlock()
			}
			break
		}

		// Handle the message
		nm.incomingMsgs <- msg
	}
}

// --- Message Processing ---

// processMessages handles all incoming and outgoing messages
func (nm *NetworkManager) processMessages() {
	for nm.isRunning {
		select {
		case msg := <-nm.incomingMsgs:
			nm.handleIncomingMessage(msg)
		case msg := <-nm.outgoingMsgs:
			nm.sendMessage(msg)
		}
	}
}

// handleIncomingMessage processes an incoming network message
func (nm *NetworkManager) handleIncomingMessage(msg NetworkMessage) {
	switch msg.Type {
	case MsgGameState:
		if nm.onGameState != nil {
			nm.onGameState(msg.PlayerID, msg.Data)
		}
	case MsgPlayerInput:
		if nm.onPlayerInput != nil {
			nm.onPlayerInput(msg.PlayerID, msg.Data)
		}
	case MsgPing:
		// Respond with pong
		nm.outgoingMsgs <- NetworkMessage{
			Type:     MsgPong,
			PlayerID: nm.config.PlayerID,
			Data:     msg.Data, // Echo the timestamp
		}
	}
}

// sendMessage sends a message to the appropriate destination
func (nm *NetworkManager) sendMessage(msg NetworkMessage) {
	if nm.config.Role == RoleServer {
		// Server sends to specific client or broadcasts
		if msg.PlayerID == "all" {
			// Broadcast to all clients
			nm.mutex.Lock()
			for _, conn := range nm.clients {
				if err := json.NewEncoder(conn).Encode(msg); err != nil {
					log.Printf("Error encoding message to client: %v", err)
				}
			}
			nm.mutex.Unlock()
		} else {
			// Send to specific client
			nm.mutex.Lock()
			if conn, ok := nm.clients[msg.PlayerID]; ok {
				if err := json.NewEncoder(conn).Encode(msg); err != nil {
					log.Printf("Error encoding message to client %s: %v", msg.PlayerID, err)
				}
			}
			nm.mutex.Unlock()
		}
	} else {
		// Client always sends to server
		if nm.conn != nil {
			if err := json.NewEncoder(nm.conn).Encode(msg); err != nil {
				log.Printf("Error encoding message to server: %v", err)
			}
		}
	}
}

// --- Public API ---

// IsNetworkInitialized returns whether the network system has been initialized
func IsNetworkInitialized() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	return networkManager != nil
}

// IsServer returns whether this instance is running as a server
func IsServer() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	return networkManager != nil && networkManager.config.Role == RoleServer
}

// IsClient returns whether this instance is running as a client
func IsClient() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	return networkManager != nil && networkManager.config.Role == RoleClient
}

// IsConnectionLost returns whether the connection to the server has been lost
func IsConnectionLost() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return false
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()
	return networkManager.connectionLost
}

// GetNetworkError returns any network error message
func GetNetworkError() string {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return ""
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()
	return networkManager.networkError
}

// IsWaitingForPlayers returns whether the server is waiting for players to connect
func IsWaitingForPlayers() bool {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return false
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()
	return networkManager.waitingForPlayers
}

// GetGameName returns the name of the multiplayer game
func GetGameName() string {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return ""
	}

	return networkManager.config.GameName
}

// GetPlayerID returns the ID of the local player
func GetPlayerID() string {
	networkMutex.Lock()
	defer networkMutex.Unlock()
	if networkManager == nil {
		return ""
	}
	return networkManager.config.PlayerID
}

// GetConnectedPlayers returns a list of connected player IDs (server only)
func GetConnectedPlayers() []string {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil || networkManager.config.Role != RoleServer {
		return []string{}
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()

	players := make([]string, 0, len(networkManager.clients))
	for id := range networkManager.clients {
		players = append(players, id)
	}
	return players
}

// --- Callback Registration ---

// SetOnConnectCallback sets the function to call when a player connects
func SetOnConnectCallback(callback func(playerID string)) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager != nil {
		networkManager.onConnect = callback
	}
}

// SetOnDisconnectCallback sets the function to call when a player disconnects
func SetOnDisconnectCallback(callback func(playerID string)) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager != nil {
		networkManager.onDisconnect = callback
	}
}

// SetOnGameStateCallback sets the function to call when game state is received
func SetOnGameStateCallback(callback func(playerID string, data []byte)) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager != nil {
		networkManager.onGameState = callback
	}
}

// SetOnPlayerInputCallback sets the function to call when player input is received
func SetOnPlayerInputCallback(callback func(playerID string, data []byte)) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager != nil {
		networkManager.onPlayerInput = callback
	}
}

// --- Message Sending ---

// SendGameState sends the current game state to all players or a specific player
func SendGameState(data []byte, targetPlayerID string) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil {
		return
	}

	networkManager.outgoingMsgs <- NetworkMessage{
		Type:     MsgGameState,
		PlayerID: targetPlayerID, // "all" for broadcast
		Data:     data,
	}
}

// SendPlayerInput sends player input to the server
func SendPlayerInput(data []byte) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil || networkManager.config.Role != RoleClient {
		return
	}

	networkManager.outgoingMsgs <- NetworkMessage{
		Type:     MsgPlayerInput,
		PlayerID: networkManager.config.PlayerID,
		Data:     data,
	}
}

// PingServer sends a ping to the server to check connection status
func PingServer() {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil || networkManager.config.Role != RoleClient {
		return
	}

	// Send current timestamp as data
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

	networkManager.outgoingMsgs <- NetworkMessage{
		Type:     MsgPing,
		PlayerID: "server",
		Data:     []byte(timestamp),
	}
}

// ParseNetworkArgs parses command line arguments for network configuration
// This is a helper function to standardize network command line arguments
func ParseNetworkArgs() *NetworkConfig {
	config := DefaultNetworkConfig()

	// Define command line flags
	var role string
	flag.StringVar(&role, "role", "server", "Role: server or client")
	flag.StringVar(&config.Address, "connect", "localhost", "Server address to connect to (client only)")
	flag.IntVar(&config.Port, "port", 8080, "Port to use for connection")
	flag.StringVar(&config.GameName, "name", config.GameName, "Name of the multiplayer game")

	// Parse flags
	flag.Parse()

	// Set role based on flag
	if role == "client" {
		config.Role = RoleClient
	} else {
		config.Role = RoleServer
	}

	return config
}

// DrawNetworkStatus draws the current network status on the screen
// This is a helper function to standardize network status display
func DrawNetworkStatus(x, y, color int) {
	networkMutex.Lock()
	defer networkMutex.Unlock()

	if networkManager == nil {
		return
	}

	networkManager.mutex.Lock()
	defer networkManager.mutex.Unlock()

	// Display network error if any
	if networkManager.networkError != "" {
		Print(networkManager.networkError, x, y, color)
		return
	}

	// Display waiting message if waiting for players
	if networkManager.waitingForPlayers {
		if networkManager.config.Role == RoleServer {
			Print("waiting for player to join...", x, y, color)
			Print("your ip: "+GetLocalIP(), x, y+10, color)
		} else {
			Print("connecting to server...", x, y, color)
		}
		return
	}

	// Display role information
	if networkManager.config.Role == RoleServer {
		Print("server mode", x, y, color)
	} else {
		Print("client mode", x, y, color)
	}
}

// InitNetworkFromMultiplayerSettings initializes the network using the deprecated MultiplayerSettings
// This is a bridge function to help transition from the old API to the new one
// Deprecated: Use InitNetwork with NetworkConfig directly instead
func InitNetworkFromMultiplayerSettings(settings *MultiplayerSettings) error {
	if settings == nil {
		return fmt.Errorf("multiplayer settings cannot be nil")
	}

	// Skip if multiplayer is not enabled
	if !settings.Enabled {
		return nil
	}

	config := &NetworkConfig{
		Role:     RoleServer,
		Port:     settings.Port,
		GameName: settings.GameName,
	}

	// Set address based on server/client role
	if settings.IsServer {
		// Server listens on all interfaces
		config.Address = ""
	} else {
		// Client connects to the specified server IP
		config.Role = RoleClient
		config.Address = settings.ServerIP
	}

	return InitNetwork(config)
}

// GetLocalIP returns the local IP address that can be used for network connections
// It finds the first non-loopback IPv4 address on an active network interface
func GetLocalIP() string {
	// Get all network interfaces
	interfaces, err := net.Interfaces()
	if err != nil {
		return "error getting IP"
	}

	// Look for a non-loopback, IPv4 address
	for _, iface := range interfaces {
		// Skip loopback and interfaces that are down
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Look for IPv4 addresses
		for _, addr := range addrs {
			var ip net.IP

			// Extract IP from address
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			// Skip non-IPv4 and loopback addresses
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// Convert to IPv4 if needed
			ip = ip.To4()
			if ip == nil {
				continue // Not an IPv4 address
			}

			// Found a valid IP address
			return ip.String()
		}
	}

	return "IP not found"
}
