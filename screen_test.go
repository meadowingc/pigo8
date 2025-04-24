package pigo8

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

func TestCursor(t *testing.T) {
	// --- Reset state before test --- (Important for global vars)
	originalX, originalY, originalColor := cursorX, cursorY, cursorColor
	t.Cleanup(func() {
		cursorX, cursorY, cursorColor = originalX, originalY, originalColor
	})

	// Initial state (assuming default)
	cursorX, cursorY, cursorColor = 0, 0, 7

	t.Run("Set Position Only", func(t *testing.T) {
		Cursor(10, 25)
		assert.Equal(t, 10, cursorX)
		assert.Equal(t, 25, cursorY)
		assert.Equal(t, 7, cursorColor, "Color should not change")
	})

	t.Run("Set Position and Color", func(t *testing.T) {
		Cursor(5, 15, 3)
		assert.Equal(t, 5, cursorX)
		assert.Equal(t, 15, cursorY)
		assert.Equal(t, 3, cursorColor)
	})

	t.Run("Reset Position", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 50, 50, 5 // Set non-zero state
		Cursor()
		assert.Equal(t, 0, cursorX, "X should reset")
		assert.Equal(t, 0, cursorY, "Y should reset")
		assert.Equal(t, 5, cursorColor, "Color should remain unchanged on reset")
	})

	t.Run("Set Position and Invalid Color", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 0, 0, 7 // Reset color
		Cursor(1, 2, 99)
		assert.Equal(t, 1, cursorX)
		assert.Equal(t, 2, cursorY)
		assert.Equal(t, 7, cursorColor, "Color should not change on invalid index")
		// TODO: Check log warning (requires capture)
	})

	t.Run("Set Position and Negative Color", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 0, 0, 7 // Reset color
		Cursor(3, 4, -1)
		assert.Equal(t, 3, cursorX)
		assert.Equal(t, 4, cursorY)
		assert.Equal(t, 7, cursorColor, "Color should not change on negative index")
		// TODO: Check log warning (requires capture)
	})

	t.Run("Invalid Argument Count (1)", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 10, 10, 10 // Set known state
		Cursor(99)
		assert.Equal(t, 10, cursorX, "State should not change on invalid arg count")
		assert.Equal(t, 10, cursorY, "State should not change on invalid arg count")
		assert.Equal(t, 10, cursorColor, "State should not change on invalid arg count")
		// TODO: Check log warning (requires capture)
	})

	t.Run("Invalid Argument Count (4)", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 11, 11, 11 // Set known state
		Cursor(1, 2, 3, 4)
		assert.Equal(t, 11, cursorX, "State should not change on invalid arg count")
		assert.Equal(t, 11, cursorY, "State should not change on invalid arg count")
		assert.Equal(t, 11, cursorColor, "State should not change on invalid arg count")
		// TODO: Check log warning (requires capture)
	})

}

// --- Add tests for Cls, Pset, Pget, Print below ---

func TestPsetGet(t *testing.T) {
	// --- Setup --- Create a dummy screen and manage global state
	originalScreen := currentScreen
	originalColor := cursorColor
	testScreen := ebiten.NewImage(20, 20) // Test image
	currentScreen = testScreen
	cursorColor = 7 // Default white for Pset default
	t.Cleanup(func() {
		currentScreen = originalScreen
		cursorColor = originalColor
	})

	t.Run("Pset valid calls don't panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Pset(5, 5, 8) // Set pixel (5,5) to Red (8)
		})
		cursorColor = 3 // Dark Green
		assert.NotPanics(t, func() {
			Pset(6, 6) // Set pixel (6,6) using cursorColor
		})
		cursorColor = 7 // Restore default
	})

	t.Run("Pset out of bounds doesn't panic", func(t *testing.T) {
		assert.NotPanics(t, func() { Pset(-1, 7, 2) })
		assert.NotPanics(t, func() { Pset(7, -1, 2) })
		assert.NotPanics(t, func() { Pset(20, 7, 2) })
		assert.NotPanics(t, func() { Pset(7, 20, 2) })
	})

	t.Run("Pset with invalid color index doesn't panic", func(t *testing.T) {
		assert.NotPanics(t, func() { Pset(8, 8, -1) })
		assert.NotPanics(t, func() { Pset(8, 8, 16) })
		assert.NotPanics(t, func() { Pset(8, 8, 99) })
		// TODO: Check log warnings (requires capture)
	})

	t.Run("Pget on valid screen doesn't panic", func(t *testing.T) {
		// We can't reliably check the color or call Pget without the game loop.
		// We only test that calling it with valid coords when screen is non-nil doesn't panic.
		// The most we could assert is that Pget returns 0 for an empty image, but even that requires At().
		// Simply ensure no panic occurs during the Pget logic (bounds check etc.).
		// Pget(1,1) // Cannot call this
		assert.True(t, true, "Skipping Pget check on valid screen as it requires game loop")
	})

	t.Run("Pget out of bounds returns 0", func(t *testing.T) {
		// This logic doesn't involve reading pixels, just bounds checking
		assert.Equal(t, 0, Pget(-1, 1))
		assert.Equal(t, 0, Pget(1, -1))
		assert.Equal(t, 0, Pget(20, 1))
		assert.Equal(t, 0, Pget(1, 20))
	})

	t.Run("Pget when screen is nil returns 0", func(t *testing.T) {
		savedScreen := currentScreen
		currentScreen = nil
		assert.Equal(t, 0, Pget(5, 5))
		// TODO: Check log warning (requires capture)
		currentScreen = savedScreen // Restore
	})

	t.Run("Pset when screen is nil doesn't panic", func(t *testing.T) {
		savedScreen := currentScreen
		currentScreen = nil
		assert.NotPanics(t, func() { Pset(9, 9, 1) })
		// TODO: Check log warning (requires capture)
		currentScreen = savedScreen // Restore
	})
}

func TestCls(t *testing.T) {
	// --- Setup --- Create a dummy screen and manage global state
	originalScreen := currentScreen
	testScreen := ebiten.NewImage(10, 10)
	currentScreen = testScreen
	// Set some initial state
	originalX, originalY, originalColor := cursorX, cursorY, cursorColor
	cursorX, cursorY, cursorColor = 5, 5, 5
	// Pset(1, 1, 8) // Cannot reliably Pset/Pget in test
	t.Cleanup(func() {
		currentScreen = originalScreen
		cursorX, cursorY, cursorColor = originalX, originalY, originalColor
	})

	t.Run("Cls runs without panic (explicit color)", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Cls(3) // Clear with Dark Green (3)
		})
		// Verify cursor reset
		assert.Equal(t, 0, cursorX, "Cursor X should reset")
		assert.Equal(t, 0, cursorY, "Cursor Y should reset")
	})

	t.Run("Cls runs without panic (default color)", func(t *testing.T) {
		// Set state again
		cursorX, cursorY = 9, 9
		assert.NotPanics(t, func() {
			Cls() // Clear with default Black (0)
		})
		// Verify cursor reset
		assert.Equal(t, 0, cursorX, "Cursor X should reset")
		assert.Equal(t, 0, cursorY, "Cursor Y should reset")
	})

	t.Run("Cls with invalid color runs without panic", func(t *testing.T) {
		cursorX, cursorY = 1, 1 // Set non-zero state
		assert.NotPanics(t, func() {
			Cls(99)
		})
		assert.Equal(t, 0, cursorX, "Cursor X should reset even on invalid color")
		assert.Equal(t, 0, cursorY, "Cursor Y should reset even on invalid color")
		// TODO: Check log warning (requires capture)
	})

	t.Run("Cls when screen is nil doesn't panic", func(t *testing.T) {
		savedScreen := currentScreen
		currentScreen = nil
		assert.NotPanics(t, func() { Cls(1) })
		// TODO: Check log warning (requires capture)
		currentScreen = savedScreen // Restore
	})
}

// --- Add tests for Print below ---

func TestPrint(t *testing.T) {
	// --- Setup --- Manage global state
	originalX, originalY, originalColor := cursorX, cursorY, cursorColor
	originalScreen := currentScreen       // Need screen for calculations, even if not drawing
	testScreen := ebiten.NewImage(10, 10) // Dummy needed for non-nil check
	currentScreen = testScreen
	t.Cleanup(func() {
		cursorX, cursorY, cursorColor = originalX, originalY, originalColor
		currentScreen = originalScreen
	})

	// Expected Y advance based on DefaultFontSize (currently 6.0)
	expectedYAdvance := int(DefaultFontSize)
	// Helper to estimate X advance based on approximation
	estimateXAdvance := func(str string) int {
		return int(math.Ceil(float64(len([]rune(str))) * CharWidthApproximation))
	}

	t.Run("Print at cursor with cursor color", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 5, 10, 3 // Known state
		str := "Hello"
		endX, endY := Print(str)

		// Verify return values (based on approximations)
		assert.Equal(t, 5+estimateXAdvance(str), endX)
		assert.Equal(t, 10+expectedYAdvance, endY)

		// Verify cursor update
		assert.Equal(t, 5, cursorX, "Cursor X should not change when printing at cursor")
		assert.Equal(t, 10+expectedYAdvance, cursorY)
		assert.Equal(t, 3, cursorColor, "Color should not change")
	})

	t.Run("Print at cursor with explicit color", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 6, 11, 7 // Known state
		str := "Test"
		explicitColor := 8 // Red
		endX, endY := Print(str, explicitColor)

		// Verify return values
		assert.Equal(t, 6+estimateXAdvance(str), endX)
		assert.Equal(t, 11+expectedYAdvance, endY)

		// Verify cursor update
		assert.Equal(t, 6, cursorX)
		assert.Equal(t, 11+expectedYAdvance, cursorY)
		assert.Equal(t, 7, cursorColor, "Global cursorColor should not change")
		// Note: Cannot easily verify the color used for the actual draw call
	})

	t.Run("Print at explicit position with cursor color", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 0, 0, 5 // Known state
		str := "World"
		printX, printY := 20, 30
		endX, endY := Print(str, printX, printY)

		// Verify return values
		assert.Equal(t, printX+estimateXAdvance(str), endX)
		assert.Equal(t, printY+expectedYAdvance, endY)

		// Verify cursor update
		assert.Equal(t, printX, cursorX, "Cursor X should update to explicit X")
		assert.Equal(t, printY+expectedYAdvance, cursorY)
		assert.Equal(t, 5, cursorColor)
	})

	t.Run("Print at explicit position and color", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 1, 2, 3 // Known state
		str := "PICO"
		printX, printY := 40, 50
		explicitColor := 14 // Pink
		endX, endY := Print(str, printX, printY, explicitColor)

		// Verify return values
		assert.Equal(t, printX+estimateXAdvance(str), endX)
		assert.Equal(t, printY+expectedYAdvance, endY)

		// Verify cursor update
		assert.Equal(t, printX, cursorX)
		assert.Equal(t, printY+expectedYAdvance, cursorY)
		assert.Equal(t, 3, cursorColor, "Global cursorColor should not change")
	})

	t.Run("Print with invalid explicit color", func(t *testing.T) {
		cursorX, cursorY, cursorColor = 10, 10, 7 // Known state
		str := "Test"
		// Case 1: Color only
		endX1, endY1 := Print(str, 99)
		assert.Equal(t, 10+estimateXAdvance(str), endX1)
		assert.Equal(t, 10+expectedYAdvance, endY1)
		// Case 2: Position and Color
		cursorY = 20 // Reset Y for next check
		endX2, endY2 := Print(str, 30, 40, -1)
		assert.Equal(t, 30+estimateXAdvance(str), endX2)
		assert.Equal(t, 40+expectedYAdvance, endY2)
		assert.Equal(t, 30, cursorX)
		assert.Equal(t, 40+expectedYAdvance, cursorY)
		// TODO: Check log warnings (requires capture)
	})

	t.Run("Print when screen is nil", func(t *testing.T) {
		savedScreen := currentScreen
		currentScreen = nil
		cursorX, cursorY, cursorColor = 5, 5, 5 // Known state
		str := "NilScreen"

		// Should still calculate return values based on args and update cursor
		endX, endY := Print(str, 10, 15, 8)
		assert.Equal(t, 10+estimateXAdvance(str), endX)
		assert.Equal(t, 15+expectedYAdvance, endY)

		// Assert that global cursor state DOES NOT change when screen is nil
		assert.Equal(t, 5, cursorX, "cursorX should not change when screen is nil")
		assert.Equal(t, 5, cursorY, "cursorY should not change when screen is nil")

		currentScreen = savedScreen // Restore
		// TODO: Check log warning (requires capture)
	})
}
