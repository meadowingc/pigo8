// main package for the pause_menu example
package main

import (
	"math/rand"
	"time"

	p8 "github.com/drpaneas/pigo8"
)

// Game implements the pigo8.Cartridge interface
type Game struct {
	playerX, playerY int
	stars            []Star
}

// Star represents a background star
type Star struct {
	x, y   int
	color  int
	speed  int
	blink  int
	active bool
}

// Init initializes the game
func (g *Game) Init() {
	// Initialize player position in the center of the screen
	g.playerX = p8.GetScreenWidth() / 2
	g.playerY = p8.GetScreenHeight() / 2

	// Initialize stars
	g.stars = make([]Star, 50)
	for i := range g.stars {
		g.stars[i] = Star{
			x:      rand.Intn(p8.GetScreenWidth()),
			y:      rand.Intn(p8.GetScreenHeight()),
			color:  5 + rand.Intn(3), // Colors 5, 6, 7 (light blue, light gray, white)
			speed:  1 + rand.Intn(2),
			blink:  rand.Intn(30),
			active: true,
		}
	}
}

// Update updates the game state
func (g *Game) Update() {
	// Update stars
	for i := range g.stars {
		// Move stars down
		g.stars[i].y += g.stars[i].speed

		// Wrap stars around when they go off screen
		if g.stars[i].y > p8.GetScreenHeight() {
			g.stars[i].y = 0
			g.stars[i].x = rand.Intn(p8.GetScreenWidth())
		}

		// Make stars blink
		g.stars[i].blink--
		if g.stars[i].blink <= 0 {
			g.stars[i].active = !g.stars[i].active
			g.stars[i].blink = rand.Intn(30)
		}
	}

	// Handle player movement with arrow keys
	if p8.Btn(p8.ButtonLeft) && g.playerX > 0 {
		g.playerX--
	}
	if p8.Btn(p8.ButtonRight) && g.playerX < p8.GetScreenWidth()-8 {
		g.playerX++
	}
	if p8.Btn(p8.ButtonUp) && g.playerY > 0 {
		g.playerY--
	}
	if p8.Btn(p8.ButtonDown) && g.playerY < p8.GetScreenHeight()-8 {
		g.playerY++
	}

	// Note: The pause state is now handled internally by the engine
}

// Draw draws the game
func (g *Game) Draw() {
	// Clear screen with dark blue
	p8.Cls(0)

	// Draw stars
	for _, star := range g.stars {
		if star.active {
			p8.Pset(star.x, star.y, star.color)
		}
	}

	// Draw player (simple spaceship)
	p8.Spr(1, g.playerX, g.playerY)

	// Draw instructions
	p8.Print("PRESS START TO PAUSE", 10, 2, 7)
}

func main() {
	// Initialize random number generator
	rand.New(rand.NewSource(time.Now().UnixNano()))

	// Create game instance
	game := &Game{}

	// Insert the game into PIGO8
	p8.InsertGame(game)

	// Configure settings
	settings := p8.NewSettings()
	settings.WindowTitle = "PIGO8 Pause Menu Demo"
	settings.ScaleFactor = 4

	// Run the game with the configured settings
	p8.PlayGameWith(settings)
}
