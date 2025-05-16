package pigo8

import (
	"testing"
)

// TestMusicFunctions tests the public music API functions
func TestMusicFunctions(t *testing.T) {
	// These tests just verify that the functions don't panic
	// They don't test actual audio playback since that would require
	// a more complex setup with audio hardware

	t.Run("Music function", func(t *testing.T) {
		// Call Music function with various parameters
		// Just testing that these don't panic
		Music(1)           // Play music 1
		Music(2, true)     // Play music 2 exclusively
		Music(-1)          // Stop all music
	})

	t.Run("StopMusic function", func(t *testing.T) {
		// Call StopMusic function with various parameters
		// Just testing that these don't panic
		StopMusic(1)       // Stop music 1
		StopMusic(-1)      // Stop all music
	})
}
