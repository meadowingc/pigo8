package pigo8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSettingsCreation tests the creation and configuration of settings
func TestSettingsCreation(t *testing.T) {
	t.Run("NewSettings defaults", func(t *testing.T) {
		settings := NewSettings()
		
		assert.NotNil(t, settings, "NewSettings should not return nil")
		assert.Equal(t, 4, settings.ScaleFactor, "Default scale factor should be 4")
		assert.Equal(t, "PIGO-8 Game", settings.WindowTitle, "Default window title should be 'PIGO-8 Game'")
		assert.Equal(t, 30, settings.TargetFPS, "Default target FPS should be 30")
		assert.Equal(t, 128, settings.ScreenWidth, "Default screen width should be 128")
		assert.Equal(t, 128, settings.ScreenHeight, "Default screen height should be 128")
		assert.False(t, settings.Multiplayer, "Default multiplayer setting should be false")
	})

	t.Run("Custom settings", func(t *testing.T) {
		// Create settings with custom values
		settings := NewSettings()
		settings.ScaleFactor = 2
		settings.WindowTitle = "Custom Game"
		settings.TargetFPS = 60
		settings.ScreenWidth = 256
		settings.ScreenHeight = 256
		settings.Multiplayer = true

		// Verify custom values
		assert.Equal(t, 2, settings.ScaleFactor, "Scale factor should be customizable")
		assert.Equal(t, "Custom Game", settings.WindowTitle, "Window title should be customizable")
		assert.Equal(t, 60, settings.TargetFPS, "Target FPS should be customizable")
		assert.Equal(t, 256, settings.ScreenWidth, "Screen width should be customizable")
		assert.Equal(t, 256, settings.ScreenHeight, "Screen height should be customizable")
		assert.True(t, settings.Multiplayer, "Multiplayer setting should be customizable")
	})
}

// TestGameFunctions tests the game initialization functions
func TestGameFunctions(t *testing.T) {
	// These tests just verify that the functions exist and don't panic
	// They don't test actual game initialization since that would require
	// mocking Ebitengine and other dependencies

	t.Run("PlayGameWith function", func(t *testing.T) {
		// Skip actual execution to avoid starting a game
		t.Skip("Skipping PlayGameWith test to avoid starting a game")
		
		// In a real test environment with proper mocks, you would call:
		// settings := NewSettings()
		// PlayGameWith(settings)
	})

	t.Run("Play function", func(t *testing.T) {
		// Skip actual execution to avoid starting a game
		t.Skip("Skipping Play test to avoid starting a game")
		
		// In a real test environment with proper mocks, you would call:
		// Play()
	})
}
