// Package main basic sprite editor
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type myGame struct{}

func (m *myGame) Init() {
	initSquareColors()
}

func (m *myGame) Update() {
	row, col, insideGrid := getMouseGridPosition()
	if insideGrid {
		if p8.Btn(p8.MOUSE_LEFT) { // Left mouse button
			setSquareColor(row, col, 3) // Set color for square at row, col
		} else if p8.Btn(p8.MOUSE_RIGHT) { // Right mouse button
			setSquareColor(row, col, 7) // Reset color for square at row, col
		}
	}

	setSquareColor(2, 3, 3) // Set color for square at row 2, column 3
	setSquareColor(0, 0, 4) // Set color for square at row 4, column 5
	setSquareColor(7, 7, 12)
}

func (m *myGame) Draw() {
	p8.Cls(5)
	// Draw an 8x8 grid of 10x10 squares
	for row := range 8 {
		for col := 0; col < 8; col++ {
			color := getSquareColor(row, col) // Get the color of the square
			if color == -1 {
				color = 7 // Default color if invalid
			}
			x := 10 + col*12 // col determines the x-coordinate
			y := 10 + row*12 // row determines the y-coordinate
			p8.Rectfill(x, y, x+10, y+10, color)
		}
	}
}

var (
	width  = 60
	height = 29
	unit   = 8
)

var squareColors [8][8]int // 8x8 grid to store square colors

func initSquareColors() {
	for row := range 8 {
		for col := 0; col < 8; col++ {
			squareColors[row][col] = 7 // Default color
		}
	}
}

func getSquareColor(row, col int) int {
	if row < 0 || row >= 8 || col < 0 || col >= 8 {
		return -1
	}
	return squareColors[row][col]
}

func setSquareColor(row, col, color int) {
	// Ensure the coordinates are within the 8x8 grid
	if row < 0 || row >= 8 || col < 0 || col >= 8 {
		return
	}
	// Update the color in the squareColors array
	squareColors[row][col] = color
}

// -- Mouse --
func getMouseGridPosition() (int, int, bool) {
	mouseX, mouseY := p8.Mouse()

	// Calculate the grid row and column based on the mouse position
	col := (mouseX - 10) / 12 // 10 is the grid offset, 12 is the square size + spacing
	row := (mouseY - 10) / 12

	// Check if the mouse is within the grid bounds
	if row >= 0 && row < 8 && col >= 0 && col < 8 {
		return row, col, true
	}

	// Return false if the mouse is outside the grid
	return -1, -1, false
}

func main() {
	settings := p8.NewSettings()
	settings.ScreenWidth = width * unit
	settings.ScreenHeight = height * unit
	settings.ScaleFactor = 5
	p8.InsertGame(&myGame{})
	p8.PlayGameWith(settings)
}
