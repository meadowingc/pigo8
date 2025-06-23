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

// updateMouseState updates the internal mouse state.
// This should be called once per frame in the game's Update method.
func updateMouseState() {
	// Update mouse position
	mouseX, mouseY = ebiten.CursorPosition()

	// Update mouse wheel values
	wheelX, wheelY := ebiten.Wheel()
	mouseWheel.x = wheelX
	mouseWheel.y = wheelY
}

// GetMouseXY returns the current mouse X and Y coordinates.
// This mimics PICO-8's mouse() function.
//
// Usage:
//
//	x, y := GetMouseXY()
//
// Example:
//
//	// Get mouse position and draw a circle at that position
//	mouseX, mouseY := GetMouseXY()
//	Circ(mouseX, mouseY, 4, 8) // Draw a circle at mouse position with radius 4 and color 8
func GetMouseXY() (int, int) {
	return mouseX, mouseY
}
