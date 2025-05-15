package pigo8

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

// TestUDPCommunication is a simple test to verify UDP communication
func TestUDPCommunication(t *testing.T) {
	// Start a UDP server
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9999")
	if err != nil {
		t.Fatalf("Failed to resolve server address: %v", err)
	}

	serverConn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Fatalf("Failed to start UDP server: %v", err)
	}
	defer func() {
		if err := serverConn.Close(); err != nil {
			log.Printf("Error closing server connection: %v", err)
		}
	}()

	// Start a UDP client
	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to resolve client address: %v", err)
	}

	clientConn, err := net.DialUDP("udp", clientAddr, serverAddr)
	if err != nil {
		t.Fatalf("Failed to start UDP client: %v", err)
	}
	defer func() {
		if err := clientConn.Close(); err != nil {
			log.Printf("Error closing client connection: %v", err)
		}
	}()

	// Server goroutine to receive messages
	serverReceived := make(chan string, 1)
	go func() {
		buffer := make([]byte, 1024)
		n, addr, err := serverConn.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("Error reading from UDP: %v", err)
			return
		}
		log.Printf("Server received %d bytes from %s: %s", n, addr.String(), string(buffer[:n]))
		serverReceived <- string(buffer[:n])

		// Send a response back
		response := "Hello from server"
		_, err = serverConn.WriteToUDP([]byte(response), addr)
		if err != nil {
			log.Printf("Error sending response: %v", err)
		}
	}()

	// Client sends a message
	message := "Hello from client"
	_, err = clientConn.Write([]byte(message))
	if err != nil {
		t.Fatalf("Failed to send message from client: %v", err)
	}
	log.Printf("Client sent message: %s", message)

	// Client receives response
	buffer := make([]byte, 1024)
	if err := clientConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		log.Printf("Error setting read deadline: %v", err)
		t.Fatalf("Failed to set read deadline: %v", err)
	}
	n, _, err := clientConn.ReadFromUDP(buffer)
	if err != nil {
		t.Fatalf("Failed to receive response: %v", err)
	}
	log.Printf("Client received response: %s", string(buffer[:n]))

	// Wait for server to receive message
	select {
	case received := <-serverReceived:
		if received != message {
			t.Errorf("Server received incorrect message: got %s, want %s", received, message)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for server to receive message")
	}

	fmt.Println("UDP communication test passed!")
}
