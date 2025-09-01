// Package main demonstrates the first camera example from the PICO-8 guide
// This example shows basic drawing with camera(0,0) - no offset
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
	// Example 1 from the PICO-8 guide:
	// Clear screen, draw background and outline with no camera offset
	p8.Cls()                       // Clear screen
	p8.Rectfill(0, 0, 127, 127, 2) // Dark purple background
	p8.Rect(0, 0, 127, 127, 8)     // Red outline
	p8.Print("camera(0,0)", 2, 2)  // Label text
}

func main() {
	p8.InsertGame(&Game{})
	p8.Play()
}
