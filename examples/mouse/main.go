// Package main provides a mouse input example for the PIGO-8 fantasy console.
package main

import (
	"fmt"

	"github.com/drpaneas/pigo8"
)

// Game implements the pigo8.Cartridge interface
type Game struct {
	// Canvas for drawing
	canvas [][]int
	// Current drawing color
	drawColor int
	// Circle radius for drawing
	brushSize int
	// Scroll position for the color palette
	paletteScroll int
	// Track if mouse was previously down (for drawing lines)
	prevMouseDown bool
	// Previous mouse position
	prevX, prevY int
}

// Init initializes the game
func (g *Game) Init() {
	// Initialize canvas (128x128 grid)
	g.canvas = make([][]int, pigo8.GetScreenHeight())
	for i := range g.canvas {
		g.canvas[i] = make([]int, pigo8.GetScreenWidth())
		// Fill with color 0 (black)
		for j := range g.canvas[i] {
			g.canvas[i][j] = 0
		}
	}

	// Set initial drawing color
	g.drawColor = 7 // White
	g.brushSize = 1 // Start with 1 pixel brush
	g.paletteScroll = 0
}

// Update handles game logic
func (g *Game) Update() {
	// Get mouse position
	mouseX, mouseY := pigo8.GetMouseXY()

	// Handle mouse wheel for brush size
	if pigo8.Btn(pigo8.ButtonMouseWheelUp) {
		g.brushSize = minInt(g.brushSize+1, 10)
	}
	if pigo8.Btn(pigo8.ButtonMouseWheelDown) {
		g.brushSize = maxInt(g.brushSize-1, 1)
	}

	// Track if we're in drawing or erasing mode
	drawing := pigo8.Btn(pigo8.ButtonMouseLeft)
	erasing := pigo8.Btn(pigo8.ButtonMouseMiddle)

	// Handle drawing or erasing on canvas
	if drawing || erasing {
		// Check if mouse is in the drawing area
		if mouseY < pigo8.GetScreenHeight()-30 {
			// Determine the color to use (draw color or black for erasing)
			color := g.drawColor
			if erasing {
				color = 0 // Use black for erasing
			}

			// Draw at current position
			g.drawCircle(mouseX, mouseY, g.brushSize, color)

			// If mouse was down in previous frame, draw a line to connect points
			if g.prevMouseDown {
				g.drawLine(g.prevX, g.prevY, mouseX, mouseY, g.brushSize, color)
			}

			// Update previous position
			g.prevX, g.prevY = mouseX, mouseY
			g.prevMouseDown = true
		}
	} else {
		g.prevMouseDown = false
	}

	// Handle color selection with right click
	if pigo8.Btnp(pigo8.ButtonMouseRight) {
		// Check if mouse is in the color palette area
		if mouseY >= pigo8.GetScreenHeight()-25 && mouseY < pigo8.GetScreenHeight()-15 {
			// Get the number of colors in the palette
			paletteSize := pigo8.GetPaletteSize()
			if paletteSize > 0 {
				// Calculate color cell width based on screen width and palette size
				cellWidth := pigo8.GetScreenWidth() / paletteSize
				if cellWidth < 4 {
					cellWidth = 4 // Minimum cell width for usability
				}

				// Calculate which color was clicked
				colorIdx := mouseX / cellWidth
				if colorIdx < paletteSize {
					g.drawColor = colorIdx
				}
			}
		}
	}

	// Clear canvas with C key
	if pigo8.Btnp(pigo8.X) {
		for i := range g.canvas {
			for j := range g.canvas[i] {
				g.canvas[i][j] = 0
			}
		}
	}
}

// Draw renders the game
func (g *Game) Draw() {
	// Clear screen
	pigo8.Cls(0)

	// Draw the canvas
	for y := range g.canvas {
		for x := range g.canvas[y] {
			if g.canvas[y][x] != 0 {
				pigo8.Pset(x, y, g.canvas[y][x])
			}
		}
	}

	// Draw color palette
	paletteSize := pigo8.GetPaletteSize()
	if paletteSize > 0 {
		// Calculate color cell width based on screen width and palette size
		cellWidth := pigo8.GetScreenWidth() / paletteSize
		if cellWidth < 4 {
			cellWidth = 4 // Minimum cell width for usability
		}

		// Draw each color in the palette
		for i := 0; i < paletteSize; i++ {
			x1 := i * cellWidth
			x2 := x1 + cellWidth - 1

			// Draw color rectangle
			pigo8.Rectfill(x1, pigo8.GetScreenHeight()-25, x2, pigo8.GetScreenHeight()-15, i)

			// Highlight selected color
			if i == g.drawColor {
				pigo8.Rect(x1, pigo8.GetScreenHeight()-25, x2, pigo8.GetScreenHeight()-15, 7)
			}
		}
	}

	// Draw UI
	mouseX, mouseY := pigo8.GetMouseXY()

	// Draw cursor as a circle outline
	pigo8.Circ(mouseX, mouseY, float64(g.brushSize), 7)

	// Draw info text
	pigo8.Print(fmt.Sprintf("mouse: %d,%d", mouseX, mouseY), 2, pigo8.GetScreenHeight()-10, 7)
	pigo8.Print(fmt.Sprintf("brush: %d", g.brushSize), 70, pigo8.GetScreenHeight()-10, 7)

	// Draw instructions
	if mouseY >= pigo8.GetScreenHeight()-30 {
		pigo8.Print("left: draw | middle: erase | right: select color | wheel: brush size | x: clear", 2, 2, 7)
	}
}

// drawCircle draws a filled circle on the canvas
func (g *Game) drawCircle(x, y, radius, color int) {
	// Calculate bounds for the circle
	minX := x - radius
	minY := y - radius
	maxX := x + radius
	maxY := y + radius

	// Ensure we're within canvas bounds
	minX = maxInt(minX, 0)
	minY = maxInt(minY, 0)
	maxX = minInt(maxX, pigo8.GetScreenWidth()-1)
	maxY = minInt(maxY, pigo8.GetScreenHeight()-31)

	// Fill the circle on our canvas
	for cy := minY; cy <= maxY; cy++ {
		for cx := minX; cx <= maxX; cx++ {
			// Check if point is inside circle (using distance formula)
			dx := cx - x
			dy := cy - y
			if dx*dx+dy*dy <= radius*radius {
				g.canvas[cy][cx] = color
			}
		}
	}
}

// drawLine draws a line between two points using Bresenham's algorithm
func (g *Game) drawLine(x0, y0, x1, y1, thickness, color int) {
	dx := abs(x1 - x0)
	dy := abs(y1 - y0)
	sx, sy := 1, 1

	if x0 >= x1 {
		sx = -1
	}
	if y0 >= y1 {
		sy = -1
	}

	err := dx - dy

	for {
		// Draw a circle at current point
		g.drawCircle(x0, y0, thickness, color)

		if x0 == x1 && y0 == y1 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x0 += sx
		}
		if e2 < dx {
			err += dx
			y0 += sy
		}
	}
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// minInt returns the smaller of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// maxInt returns the larger of two integers
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	// Create game instance
	game := &Game{}

	// Insert the game into the PIGO-8 console
	pigo8.InsertGame(game)

	// Configure settings
	settings := pigo8.NewSettings()
	settings.WindowTitle = "PIGO-8 Mouse Example"

	// Start the game
	pigo8.PlayGameWith(settings)
}
