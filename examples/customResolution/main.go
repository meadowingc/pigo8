package main

import (
	"log"
	"math/rand"

	p8 "github.com/drpaneas/pigo8"
)

// Game implements the p8.Cartridge interface
type Game struct {
	playerX    int
	playerY    int
	stars      []star
	frameCount int
}

type star struct {
	x, y  int
	color int
	speed int
}

// Init is called once at the start of the game
func (g *Game) Init() {
	log.Println("Game Boy style game initialized with custom resolution (160x144)")

	// Initialize player position in the middle of the screen
	g.playerX = 80
	g.playerY = 120

	// Create stars for the background
	g.stars = make([]star, 50)
	for i := range g.stars {
		g.stars[i] = star{
			x:     rand.Intn(p8.ScreenWidth),
			y:     rand.Intn(p8.ScreenHeight),
			color: 6 + rand.Intn(2), // Light gray or white
			speed: 1 + rand.Intn(2),
		}
	}
}

// Update is called every frame for game logic
func (g *Game) Update() {
	g.frameCount++

	// Update star positions
	for i := range g.stars {
		g.stars[i].y += g.stars[i].speed
		if g.stars[i].y > p8.ScreenHeight {
			g.stars[i].y = 0
			g.stars[i].x = rand.Intn(p8.ScreenWidth)
		}
	}

	// Handle player movement with arrow keys
	if p8.Btn(0) { // Left
		g.playerX--
		if g.playerX < 4 {
			g.playerX = 4
		}
	}
	if p8.Btn(1) { // Right
		g.playerX++
		if g.playerX > p8.ScreenWidth-4 {
			g.playerX = p8.ScreenWidth - 4
		}
	}
	if p8.Btn(2) { // Up
		g.playerY--
		if g.playerY < 4 {
			g.playerY = 4
		}
	}
	if p8.Btn(3) { // Down
		g.playerY++
		if g.playerY > p8.ScreenHeight-4 {
			g.playerY = p8.ScreenHeight - 4
		}
	}
}

// Draw is called every frame for rendering
func (g *Game) Draw() {
	// Clear screen with dark blue (Game Boy style)
	p8.Cls(1)

	// Draw stars
	for _, s := range g.stars {
		p8.Pset(s.x, s.y, s.color)
	}

	// Draw player (a simple ship)
	p8.Pset(g.playerX, g.playerY, 7)     // Center
	p8.Pset(g.playerX-1, g.playerY+1, 7) // Left wing
	p8.Pset(g.playerX+1, g.playerY+1, 7) // Right wing
	p8.Pset(g.playerX, g.playerY-1, 8)   // Red nose

	// Draw Game Boy style border
	drawBorder()

	// Draw text
	p8.Print("game boy style", 50, 5, 7)
	p8.Print("160 x 144", 60, 14, 7)
}

// drawBorder draws a simple border around the screen
func drawBorder() {
	// Top and bottom borders
	for x := range p8.ScreenWidth {
		p8.Pset(x, 0, 7)
		p8.Pset(x, p8.ScreenHeight-1, 7)
	}

	// Left and right borders
	for y := range p8.ScreenHeight {
		p8.Pset(0, y, 7)
		p8.Pset(p8.ScreenWidth-1, y, 7)
	}
}

func main() {
	// Create custom settings with Game Boy resolution (160x144)
	settings := p8.NewSettings()
	settings.TargetFPS = 60
	settings.ScreenWidth = 160
	settings.ScreenHeight = 144
	settings.WindowTitle = "Game Boy Style Demo"

	// Insert our game and play with custom settings
	p8.InsertGame(&Game{})
	p8.PlayGameWith(settings)
}
