package pigo8

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Note: Mouse button constants are defined in controls.go

// mouseX and mouseY store the current mouse position
var (
	mouseX     int
	mouseY     int
	mouseWheel struct {
		x float64
		y float64
	}
)

// UpdateMouseState updates the internal mouse state.
// This should be called once per frame in the game's Update method.
func UpdateMouseState() {
	// Update mouse position
	mouseX, mouseY = ebiten.CursorPosition()

	// Update mouse wheel values
	wheelX, wheelY := ebiten.Wheel()
	mouseWheel.x = wheelX
	mouseWheel.y = wheelY
}

// Mouse returns the current mouse X and Y coordinates.
// This mimics PICO-8's mouse() function.
//
// Usage:
//
//	x, y := Mouse()
//
// Example:
//
//	// Get mouse position and draw a circle at that position
//	mouseX, mouseY := Mouse()
//	Circ(mouseX, mouseY, 4, 8) // Draw a circle at mouse position with radius 4 and color 8
func Mouse() (int, int) {
	return mouseX, mouseY
}

// MouseBtn checks if a specific mouse button is currently held down.
// This is a convenience wrapper around Btn for mouse-specific buttons.
//
// buttonIndex: The mouse button index (MOUSE_LEFT, MOUSE_RIGHT).
//
// Usage:
//
//	MouseBtn(buttonIndex)
//
// Example:
//
//	// Check if the left mouse button is held
//	if MouseBtn(MOUSE_LEFT) {
//		// Do something when left mouse button is held
//	}
func MouseBtn(buttonIndex int) bool {
	// Simply use the Btn function with the mouse button index
	return Btn(buttonIndex)
}

// MouseBtnJustPressed checks if a specific mouse button was just pressed.
// This is a convenience wrapper around Btnp for mouse-specific buttons.
//
// buttonIndex: The mouse button index (MOUSE_LEFT, MOUSE_RIGHT).
//
// Usage:
//
//	MouseBtnJustPressed(buttonIndex)
//
// Example:
//
//	// Check if the right mouse button was just pressed
//	if MouseBtnJustPressed(MOUSE_RIGHT) {
//		// Do something when right mouse button is just pressed
//	}
func MouseBtnJustPressed(buttonIndex int) bool {
	// Simply use the Btnp function with the mouse button index
	return Btnp(buttonIndex)
}

// This function is no longer needed as we're using Btnp directly
