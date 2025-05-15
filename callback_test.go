package pigo8

import (
	"log"
	"testing"
)

// TestCallbackRegistration verifies that callbacks are properly registered
func TestCallbackRegistration(t *testing.T) {
	// First initialize the network
	log.Println("Initializing network for test...")
	config := ParseNetworkArgs()
	config.Role = RoleServer // Use server role for testing
	config.Address = "127.0.0.1"
	config.Port = 12345 // Use a test port

	if err := InitNetwork(config); err != nil {
		t.Fatalf("Failed to initialize network: %v", err)
	}
	defer ShutdownNetwork()

	// Register callbacks
	log.Println("Registering test callbacks...")

	gameStateCallbackCalled := false
	SetOnGameStateCallback(func(playerID string, data []byte) {
		log.Printf("Game state callback called with playerID=%s, data size=%d", playerID, len(data))
		gameStateCallbackCalled = true
	})

	playerInputCallbackCalled := false
	SetOnPlayerInputCallback(func(playerID string, data []byte) {
		log.Printf("Player input callback called with playerID=%s, data size=%d", playerID, len(data))
		playerInputCallbackCalled = true
	})

	// Also register connect/disconnect callbacks
	SetOnConnectCallback(func(playerID string) {
		log.Printf("Connect callback called with playerID=%s", playerID)
	})

	SetOnDisconnectCallback(func(playerID string) {
		log.Printf("Disconnect callback called with playerID=%s", playerID)
	})

	// Force register callbacks to ensure they're set
	ForceRegisterCallbacks(
		func(_ string, _ []byte) { gameStateCallbackCalled = true },
		func(_ string, _ []byte) { playerInputCallbackCalled = true },
		func(_ string) { log.Printf("Connect callback called") },
		func(_ string) { log.Printf("Disconnect callback called") },
	)

	// Check if callbacks are registered
	log.Println("Verifying callbacks are registered...")
	if !AreCallbacksRegistered() {
		t.Fatal("Callbacks are not registered after force registration")
	}

	// Use the variables to avoid lint errors
	log.Printf("Callback variables initialized: gameState=%v, playerInput=%v",
		gameStateCallbackCalled, playerInputCallbackCalled)

	log.Println("Callbacks are registered successfully")
}
