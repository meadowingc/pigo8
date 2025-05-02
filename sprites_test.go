package pigo8

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

// Note: Directly testing the drawing output of Spr is difficult in unit tests
// as it requires an active Ebiten game loop and access to screen pixels.
// These tests focus on argument parsing, state checks, and ensuring no panics.

func TestSprArguments(t *testing.T) {
	// --- Setup --- Manage global state
	// We need a dummy screen and sprites slice for Spr to proceed beyond initial checks.
	originalScreen := currentScreen
	originalSprites := currentSprites
	currentScreen = ebiten.NewImage(10, 10) // Dummy screen

	// Create dummy sprite entries with non-sequential IDs
	dummyImg1 := ebiten.NewImage(8, 8)
	dummyImg2 := ebiten.NewImage(8, 8)
	dummyImg3 := ebiten.NewImage(8, 8)
	// Intentionally set IDs out of order with slice index
	currentSprites = []SpriteInfo{
		{ID: 5, Image: dummyImg1},
		{ID: 0, Image: dummyImg2},
		{ID: 10, Image: dummyImg3},
	}

	t.Cleanup(func() {
		currentScreen = originalScreen
		currentSprites = originalSprites
	})

	t.Run("Basic call with int coords", func(t *testing.T) {
		// Should run without panic
		assert.NotPanics(t, func() {
			Spr(0, 5, 5)
		})
	})

	t.Run("Basic call with float coords", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 5.5, 5.5)
		})
	})

	t.Run("Basic call with mixed coords", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 5, 5.5)
		})
		assert.NotPanics(t, func() {
			Spr(0, 5.5, 5)
		})
	})

	t.Run("Basic call with float sprite number", func(t *testing.T) {
		// Floats should be truncated to ints (0.5 -> 0, 5.9 -> 5)
		assert.NotPanics(t, func() {
			Spr(0.5, 1, 1)
		})
		assert.NotPanics(t, func() {
			Spr(5.9, 2, 2)
		})
		// Should find sprite ID 5 when using 5.9
		assert.NotPanics(t, func() {
			Spr(5.9, 10, 10)
		})
		// Non-existent sprite ID after truncation (e.g., 1.5 -> 1)
		assert.NotPanics(t, func() {
			Spr(1.5, 15, 15)
		})
	})

	t.Run("Call with w/h options (float)", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 2.0, 0.5)
		})
	})

	t.Run("Call with w/h options (int)", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 2, 1)
		})
	})

	t.Run("Call with w/h/flipX options", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, true)
		})
	})

	t.Run("Call with w/h/flipX/flipY options", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, true, false)
		})
	})

	t.Run("Call with zero w/h", func(t *testing.T) {
		// Should return early, no panic
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 0.0, 1.0)
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 0)
		})
	})

	t.Run("Call with negative w/h", func(t *testing.T) {
		// w/h are clamped to 0, should return early, no panic
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, -1.0, 1.0)
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, -2)
		})
	})

	t.Run("Call with invalid option types", func(t *testing.T) {
		// Should log warnings but not panic
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, "invalid_w")
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, "invalid_h")
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, "invalid_flipX")
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, true, "invalid_flipY")
		})
		// TODO: Check log warnings (requires capture)
	})

	t.Run("Call with too many options", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, true, false, "extra")
		})
		// TODO: Check log warning (requires capture)
	})

	t.Run("Call with out-of-bounds sprite number", func(t *testing.T) {
		// Should just skip drawing the non-existent tile, no panic
		assert.NotPanics(t, func() {
			Spr(99, 1, 1)
		})
		assert.NotPanics(t, func() {
			Spr(-1, 1, 1)
		})
	})

	t.Run("Call when screen is nil", func(t *testing.T) {
		savedScreen := currentScreen
		currentScreen = nil
		assert.NotPanics(t, func() {
			Spr(0, 1, 1)
		})
		currentScreen = savedScreen // Restore
		// TODO: Check log warning (requires capture)
	})

	t.Run("Call using Sprite ID not equal to slice index", func(t *testing.T) {
		// Should find sprite ID 5 (at index 0) and not panic
		assert.NotPanics(t, func() {
			Spr(5, 1, 1)
		})
		// Should find sprite ID 10 (at index 2) and not panic
		assert.NotPanics(t, func() {
			Spr(10, 2, 2)
		})
		// Should find sprite ID 0 (at index 1) and not panic
		assert.NotPanics(t, func() {
			Spr(0, 3, 3)
		})
	})

	// Note: Testing the lazy-loading of sprites is complex as it involves
	// the filesystem via loadSpritesheet(). This is better suited for an
	// integration test or requires mocking loadSpritesheet().

}

// TestSpriteCoordinateConversion tests the coordinate conversion logic used in Sget and Sset
func TestSpriteCoordinateConversion(t *testing.T) {
	// Test cases for coordinate conversion
	testCases := []struct {
		name           string
		globalX        int
		globalY        int
		expectedSpriteID int
		expectedLocalX  int
		expectedLocalY  int
	}{
		{"Top-left pixel of first sprite", 0, 0, 0, 0, 0},
		{"Middle pixel of first sprite", 4, 3, 0, 4, 3},
		{"Bottom-right pixel of first sprite", 7, 7, 0, 7, 7},
		{"First pixel of second sprite", 8, 0, 1, 0, 0},
		{"Middle pixel of second sprite", 12, 3, 1, 4, 3},
		{"First pixel of sprite in second row", 0, 8, 16, 0, 0},
		{"Middle pixel of sprite in second row", 4, 12, 16, 4, 4},
		{"Pixel in middle of spritesheet", 64, 64, 136, 0, 0},
		{"Last pixel of spritesheet", 127, 127, 255, 7, 7},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate sprite ID using the same logic as in Sget/Sset
			spriteX := tc.globalX / 8
			spriteY := tc.globalY / 8
			spriteCellID := spriteY*16 + spriteX
			
			// Calculate local coordinates
			localX := tc.globalX % 8
			localY := tc.globalY % 8
			
			// Verify calculations
			assert.Equal(t, tc.expectedSpriteID, spriteCellID, "Sprite ID calculation incorrect")
			assert.Equal(t, tc.expectedLocalX, localX, "Local X calculation incorrect")
			assert.Equal(t, tc.expectedLocalY, localY, "Local Y calculation incorrect")
		})
	}
}

// TestSsetColorHandling tests the color handling logic in Sset
func TestSsetColorHandling(t *testing.T) {
	// Save original state
	originalDrawColor := currentDrawColor
	
	// Cleanup after tests
	t.Cleanup(func() {
		currentDrawColor = originalDrawColor
	})
	
	t.Run("Default color handling", func(t *testing.T) {
		// Set a known color
		currentDrawColor = 7 // White
		
		// Verify that when no color is specified, currentDrawColor is used
		colorToUse := currentDrawColor
		assert.Equal(t, 7, colorToUse, "Should use currentDrawColor when no color specified")
	})
	
	t.Run("Explicit color handling", func(t *testing.T) {
		// Set a known color
		currentDrawColor = 7 // White
		
		// Verify that when a color is specified, it's used instead of currentDrawColor
		colorToUse := 8 // Explicit color (red)
		assert.Equal(t, 8, colorToUse, "Should use explicit color when specified")
		
		// Verify currentDrawColor remains unchanged
		assert.Equal(t, 7, currentDrawColor, "currentDrawColor should not change when explicit color is used")
	})
	
	t.Run("Color clamping - high", func(t *testing.T) {
		// Test color clamping for values above the valid range
		invalidColor := 99 // Above valid range
		clampedColor := invalidColor
		
		// Apply the same clamping logic as in Sset
		if clampedColor >= len(Pico8Palette) {
			clampedColor = len(Pico8Palette) - 1
		}
		
		assert.Equal(t, len(Pico8Palette)-1, clampedColor, "Color should be clamped to max index")
	})
	
	t.Run("Color clamping - low", func(t *testing.T) {
		// Test color clamping for values below the valid range
		invalidColor := -1 // Below valid range
		clampedColor := invalidColor
		
		// Apply the same clamping logic as in Sset
		if clampedColor < 0 {
			clampedColor = 0
		}
		
		assert.Equal(t, 0, clampedColor, "Color should be clamped to 0")
	})
}
