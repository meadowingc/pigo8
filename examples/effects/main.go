// main package for transparency effects demo
package main

import (
	"image/color"
	"log"
	"math"

	"github.com/drpaneas/pigo8"
	"github.com/hajimehoshi/ebiten/v2"
)

// Game represents our game state
type Game struct {
	// Animation counters
	tick         int
	fadeValue    uint8
	fadeIn       bool
	ghostX       float64
	waterOffset  float64
	playerX      int
	playerY      int
	showControls bool
}

// NewGame creates a new game instance
func NewGame() *Game {
	return &Game{
		tick:         0,
		fadeValue:    0,
		fadeIn:       true,
		ghostX:       20,
		waterOffset:  0,
		playerX:      64,
		playerY:      80,
		showControls: true,
	}
}

// Init initializes the game
func (g *Game) Init() {
	log.Println("Initializing transparency effects demo")
}

// Update updates the game state
func (g *Game) Update() {
	// Update tick counter
	g.tick++

	// Handle player movement
	if pigo8.Btn(pigo8.LEFT) && g.playerX > 10 {
		g.playerX--
	}
	if pigo8.Btn(pigo8.RIGHT) && g.playerX < pigo8.GetScreenWidth()-10 {
		g.playerX++
	}
	if pigo8.Btn(pigo8.UP) && g.playerY > 10 {
		g.playerY--
	}
	if pigo8.Btn(pigo8.DOWN) && g.playerY < pigo8.GetScreenHeight()-10 {
		g.playerY++
	}

	// Toggle controls display with X key
	if pigo8.Btnp(pigo8.X) {
		g.showControls = !g.showControls
	}

	// Update fade effect
	if g.fadeIn {
		if g.fadeValue < 255 {
			g.fadeValue += 2
		} else {
			g.fadeIn = false
		}
	} else {
		if g.fadeValue > 0 {
			g.fadeValue -= 2
		} else {
			g.fadeIn = true
		}
	}

	// Update ghost position
	g.ghostX += math.Sin(float64(g.tick)/20) * 0.5
	if g.ghostX < 10 {
		g.ghostX = 10
	}
	if g.ghostX > float64(pigo8.GetScreenWidth()-20) {
		g.ghostX = float64(pigo8.GetScreenWidth() - 20)
	}

	// Update water animation
	g.waterOffset += 0.2
	if g.waterOffset > 8 {
		g.waterOffset = 0
	}
}

// Draw draws the game
func (g *Game) Draw() {
	// Clear the screen
	pigo8.Cls(1) // Dark blue background

	// Get the screen for direct drawing
	screen := pigo8.CurrentScreen()

	// ---- 1. Draw the background scene ----
	// Draw ground
	pigo8.Rectfill(0, 90, pigo8.GetScreenWidth(), pigo8.GetScreenHeight(), 4) // Brown

	// Draw some trees
	drawTree(20, 85)
	drawTree(40, 87)
	drawTree(100, 86)

	// ---- 2. Draw player character ----
	pigo8.Circfill(g.playerX, g.playerY, 5, 8) // Red circle

	// ---- 3. Draw player shadow (transparency effect #1) ----
	shadowImg := ebiten.NewImage(12, 6)
	shadowImg.Fill(color.RGBA{0, 0, 0, 128}) // Semi-transparent black

	shadowOp := &ebiten.DrawImageOptions{}
	shadowOp.GeoM.Translate(float64(g.playerX-6), float64(g.playerY+6))
	shadowOp.Blend = ebiten.BlendSourceOver
	screen.DrawImage(shadowImg, shadowOp)

	// ---- 4. Draw ghost character (transparency effect #2) ----
	// Create ghost shape
	ghostImg := ebiten.NewImage(16, 20)

	// Draw ghost body (white with transparency)
	for y := 0; y < 20; y++ {
		for x := 0; x < 16; x++ {
			// Create a ghost shape
			if y < 12 && (x < 2 || x > 13) {
				continue // Skip corners to make rounded top
			}

			// Make the bottom wavy
			if y > 14 {
				if (x+g.tick/5)%4 == 0 {
					continue // Create wavy bottom
				}
			}

			// Set with transparency
			alpha := uint8(180 + 40*math.Sin(float64(g.tick)/15))
			ghostImg.Set(x, y, color.RGBA{255, 255, 255, alpha})
		}
	}

	// Draw ghost eyes
	for i := 0; i < 2; i++ {
		eyeX := 5 + i*6
		ghostImg.Set(eyeX, 8, color.RGBA{0, 0, 255, 255})
		ghostImg.Set(eyeX+1, 8, color.RGBA{0, 0, 255, 255})
	}

	// Draw the ghost with transparency
	ghostOp := &ebiten.DrawImageOptions{}
	ghostOp.GeoM.Translate(float64(g.ghostX), 40)
	ghostOp.Blend = ebiten.BlendSourceOver
	screen.DrawImage(ghostImg, ghostOp)

	// ---- 5. Draw water overlay (transparency effect #3) ----
	// Only in the bottom part of the screen
	waterImg := ebiten.NewImage(pigo8.GetScreenWidth(), 30)

	// Fill with semi-transparent blue
	waterImg.Fill(color.RGBA{41, 173, 255, 120}) // Light blue with transparency

	// Add some wave patterns
	for y := 0; y < 30; y++ {
		for x := 0; x < pigo8.GetScreenWidth(); x++ {
			if (x+int(g.waterOffset)+y)%8 == 0 {
				waterImg.Set(x, y, color.RGBA{41, 173, 255, 180}) // Slightly darker wave lines
			}
		}
	}

	// Draw the water
	waterOp := &ebiten.DrawImageOptions{}
	waterOp.GeoM.Translate(0, 98)
	waterOp.Blend = ebiten.BlendSourceOver
	screen.DrawImage(waterImg, waterOp)

	// ---- 6. Draw title with fade effect (transparency effect #4) ----
	// First draw the title normally
	pigo8.Print("Transparency Effects Demo", 14, 5, 7)

	// Then overlay a fading rectangle
	fadeImg := ebiten.NewImage(pigo8.GetScreenWidth(), 15)
	fadeImg.Fill(color.RGBA{29, 43, 83, g.fadeValue}) // Background color with changing alpha

	fadeOp := &ebiten.DrawImageOptions{}
	fadeOp.GeoM.Translate(0, 0)
	fadeOp.Blend = ebiten.BlendSourceOver
	screen.DrawImage(fadeImg, fadeOp)

	// ---- 7. Draw controls if enabled ----
	if g.showControls {
		// Draw semi-transparent help box
		helpImg := ebiten.NewImage(pigo8.GetScreenWidth()-20, 25)
		helpImg.Fill(color.RGBA{0, 0, 0, 200})

		helpOp := &ebiten.DrawImageOptions{}
		helpOp.GeoM.Translate(10, 20)
		helpOp.Blend = ebiten.BlendSourceOver
		screen.DrawImage(helpImg, helpOp)

		// Draw help text
		pigo8.Print("help", 15, 27, 7)
	}
}

// Helper function to draw a simple tree
func drawTree(x, y int) {
	// Draw trunk
	pigo8.Rectfill(x-2, y-15, x+2, y, 4) // Brown

	// Draw leaves
	pigo8.Circfill(x, y-18, 6, 3) // Dark green
}

func main() {
	game := NewGame()

	// Insert the game into PIGO8
	pigo8.InsertGame(game)

	// Configure settings
	settings := pigo8.NewSettings()
	settings.WindowTitle = "PIGO8 Transparency Effects Demo"
	settings.ScaleFactor = 4

	// Run the game with the configured settings
	pigo8.PlayGameWith(settings)
}
