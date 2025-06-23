// main package for palette demo
package main

import (
	"image/color"
	"log"

	"github.com/drpaneas/pigo8"
)

// Game represents our game state
type Game struct {
	greenPalette []color.Color
}

// NewGame creates a new game instance
func NewGame() *Game {
	// Create a 4-color green palette from the specified hex colors
	return &Game{
		greenPalette: []color.Color{
			color.RGBA{38, 52, 35, 255},    // #263423 - Dark green
			color.RGBA{83, 112, 47, 255},   // #53702f - Medium green
			color.RGBA{166, 182, 61, 255},  // #a6b63d - Light green
			color.RGBA{241, 243, 192, 255}, // #f1f3c0 - Pale yellow
		},
	}
}

// Init initializes the game
func (g *Game) Init() {
	// Set our custom green palette
	pigo8.SetPalette(g.greenPalette)
	log.Println("Set custom green palette with 4 colors")
}

// Update updates the game state
func (g *Game) Update() {}

// Draw draws the game
func (g *Game) Draw() {
	// Clear screen with pale yellow (#f1f3c0 - color 3)
	pigo8.Cls(3)

	// Draw title
	pigo8.Print("custom palette demo", 25, 10, 0) // Use dark green for text

	// Center of the screen
	centerX := pigo8.GetScreenWidth() / 2
	centerY := pigo8.GetScreenHeight() / 2

	// Draw three concentric circles with the green colors
	pigo8.Circfill(centerX, centerY, 40, 0) // Dark green circle
	pigo8.Circfill(centerX, centerY, 30, 1) // Medium green circle
	pigo8.Circfill(centerX, centerY, 20, 2) // Light green circle
}

func main() {
	game := NewGame()

	// Insert the game into PIGO8
	pigo8.InsertGame(game)

	// Configure settings
	settings := pigo8.NewSettings()
	settings.WindowTitle = "PIGO8 Dynamic Palette Demo"
	settings.ScaleFactor = 4

	// Run the game with the configured settings
	pigo8.PlayGameWith(settings)
}
