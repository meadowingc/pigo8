// controller_demo demonstrates input handling by drawing an NES-style controller
// and highlighting buttons when they are pressed.
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

// Game represents the game state
type Game struct {
}

// Init initializes the game
func (g *Game) Init() {
}

// Update updates the game state
func (g *Game) Update() {
}

// Draw draws the controller and highlights pressed buttons
func (g *Game) Draw() {
	p8.Cls(1)

	// print in the center screen
	p8.Print("controller", 40, 20, 10)

	p8.Rectfill(10, 50, 117, 90, 7)
	p8.Rect(10, 50, 117, 90, 0)

	// inner rect grey
	p8.Rectfill(12, 52, 115, 88, 5)

	// even more inner rect
	p8.Rectfill(14, 54, 113, 86, 7)

	// and last even even more inner rect
	p8.Rectfill(16, 56, 111, 84, 0)

	// dpad cross style, meaning 4 different rectangles (left, right, up, down)

	p8.Rectfill(20, 65, 27, 73, 0) // left
	p8.Rect(20, 65, 27, 73, 7)     // outer left perimeter
	p8.Print("0", 22, 66, 12)      // print inside it the number 0
	if p8.Btn(0) {                 // p8.ButtonLeft
		p8.Print("0", 22, 66, 4)
	}

	p8.Rectfill(27, 65, 34, 73, 0) // right
	p8.Rect(27, 65, 34, 73, 7)     // outer right perimeter

	p8.Rectfill(34, 65, 41, 73, 0)
	p8.Rect(34, 65, 41, 73, 7)
	p8.Print("1", 36, 66, 12)
	if p8.Btn(1) { // p8.ButtonRight
		p8.Print("1", 36, 66, 4)
	}

	// up
	p8.Rectfill(27, 65-8, 34, 73-8, 0)
	p8.Rect(27, 65-8, 34, 73-8, 7)
	p8.Print("2", 29, 66-8, 12)
	if p8.Btn(2) { // p8.ButtonUp
		p8.Print("2", 29, 66-8, 4)
	}

	// down
	p8.Rectfill(27, 73, 34, 81, 0)
	p8.Rect(27, 73, 34, 81, 7)
	p8.Print("3", 29, 74, 12)
	if p8.Btn(3) { // p8.ButtonDown
		p8.Print("3", 29, 74, 4)
	}

	// rectangle grey in the middle
	p8.Rectfill(50, 75, 76, 75+4, 7)
	// inside of it another reactangle smaller one
	p8.Rectfill(52, 77, 58, 75+2, 0)
	p8.Rectfill(68, 77, 74, 75+2, 0)

	// draw one button to the right side
	p8.Rectfill(83, 62, 83+11, 73, 7)
	// draw inside a red circle
	p8.Circfill(83+6, 68, 5, 8)
	p8.Print("4", 83+4, 65, 12)
	if p8.Btn(4) { // p8.O
		p8.Print("4", 83+4, 65, 4)
	}

	// draw another button next to it, on the right side
	p8.Rectfill(98, 62, 98+11, 73, 7)
	// draw inside a red circle
	p8.Circfill(98+6, 68, 5, 8)
	p8.Print("5", 98+4, 65, 12)
	if p8.Btn(5) { // p8.X
		p8.Print("5", 98+4, 65, 4)
	}

}

func main() {
	// Create a new game instance
	game := &Game{}

	// Insert the game into the engine
	p8.InsertGame(game)

	// Start the game
	p8.Play()
}
