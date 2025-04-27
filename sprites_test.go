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
