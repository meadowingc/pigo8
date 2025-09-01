// Package main demonstrates the second camera example from the PICO-8 guide
// This example shows how camera offset affects previously drawn elements
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

// Game holds the state of our game
type Game struct{}

// Init is called once at the beginning of the game
func (g *Game) Init() {}

// Update is called once per frame
func (g *Game) Update() {}

// Draw is called once per frame
func (g *Game) Draw() {
	// Example 2 from the PICO-8 guide:
	// Shows how camera offset affects previously drawn elements
	p8.Cls()                       // Clear screen
	p8.Rectfill(0, 0, 127, 127, 2) // Dark purple background
	p8.Rect(0, 0, 127, 127, 8)     // Red outline
	p8.Print("camera(0,0)", 2, 2)  // Label text

	// Now set camera offset - this will affect the previously drawn elements too!
	p8.Camera(63, 63)                   // Start camera offset
	p8.Rect(63, 63, 127+63, 127+63, 11) // New camera outline (green)
	p8.Print("camera(63,63)", 136, 182) // Label text for new camera position

	// Note: The previously drawn elements (dark purple background, red outline,
	// and "camera(0,0)" text) will be shifted up and to the left by the camera offset,
	// even though they were drawn before calling Camera(63, 63)
}

func main() {
	p8.InsertGame(&Game{})
	p8.Play()
}
