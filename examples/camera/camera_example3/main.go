// Package main demonstrates the third camera example from the PICO-8 guide
// This example shows using two cameras to lock UI elements in place
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
	// Example 3 from the PICO-8 guide:
	// Using two cameras to create locked UI overlay
	p8.Cls()

	// First camera call - locks the following elements in place
	p8.Camera()                    // Set first camera (0,0)
	p8.Rectfill(0, 0, 127, 127, 2) // Dark purple background
	p8.Rect(0, 0, 127, 127, 8)     // Red outline
	p8.Print("camera(0,0)", 2, 2)  // Label text

	// Second camera call - these elements will be offset but the previous ones stay locked
	p8.Camera(63, 63)                   // Set second camera (63,63)
	p8.Rect(63, 63, 190, 190, 11)       // Green outline (adjusted coordinates)
	p8.Print("camera(63,63)", 136, 182) // Label text for second camera

	// The first set of draw operations (dark purple background, red outline,
	// and "camera(0,0)" text) are now locked in position by the first camera call.
	// They will NOT be affected by the second camera offset.
	// This is useful for creating UI overlays (health, score, etc.) that don't
	// move with the game world camera.
}

func main() {
	p8.InsertGame(&Game{})
	p8.Play()
}
