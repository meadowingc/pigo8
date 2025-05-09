// Package main demonstrates loading a palette from a file
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"fmt"

	"github.com/drpaneas/pigo8"
)

// Game implements the pigo8.Cartridge interface
type Game struct {
	currentColor int
}

// Init loads things once
func (g *Game) Init() {}

// Update is called every frame and is responsible for updating the game state
func (g *Game) Update() {
	if pigo8.Btn(pigo8.O) {
		g.currentColor = (g.currentColor + 1) % pigo8.GetPaletteSize()
	}
}

// Draw is called every frame and is responsible for rendering the game
func (g *Game) Draw() {
	pigo8.Cls(0)

	// Draw a grid of all palette colors
	paletteSize := pigo8.GetPaletteSize()

	// Draw title
	pigo8.Print("load palette from file", 20, 4, 7)
	pigo8.Print(fmt.Sprintf("total colors: %d", paletteSize), 20, 12, 7)
	pigo8.Print("press 'z' to change color", 20, 20, 5)

	// Calculate grid dimensions
	cellSize := 8
	cols := 8
	startX := 16
	startY := 32

	// Draw the color grid
	for i := range paletteSize {
		x := startX + (i%cols)*cellSize
		y := startY + (i/cols)*cellSize

		// Draw color square
		for dy := range cellSize - 1 {
			for dx := range cellSize - 1 {
				pigo8.Pset(x+dx, y+dy, i)
			}
		}

		// Highlight the current color
		if i == g.currentColor {
			pigo8.Rect(x-1, y-1, x+cellSize-1, y+cellSize-1, 7)
		}
	}

	// Display information about the current color
	currentColorY := startY + ((paletteSize+cols-1)/cols)*cellSize + 8

	// Get the actual color value
	clr := pigo8.GetPaletteColor(g.currentColor)
	r, gb, b, a := clr.RGBA()
	r, gb, b, a = r>>8, gb>>8, b>>8, a>>8 // Convert from 0-65535 to 0-255 range

	pigo8.Print("current color:", 40, currentColorY, 5)
	pigo8.Print(fmt.Sprintf("%d", g.currentColor), 100, currentColorY, 14)

	pigo8.Print("rgba:", 40, currentColorY+8, 5)
	pigo8.Print(fmt.Sprintf("%d,%d,%d,%d", r, gb, b, a), 70, currentColorY+8, 14)

	// Draw a larger sample of the current color
	pigo8.Rectfill(90, 50, 120, 80, g.currentColor)
}

func main() {
	settings := pigo8.NewSettings()
	settings.WindowTitle = "PIGO-8 Custom Palette Example"
	pigo8.InsertGame(&Game{})
	pigo8.PlayGameWith(settings)
}
