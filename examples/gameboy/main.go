// Package main demonstrates a Game Boy style game with sprite animation and collision detection
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"image"

	p8 "github.com/drpaneas/pigo8"
)

// Game holds all game state
type Game struct {
	// Player state
	pos       p8.Vector2D // x, y position (in pixels)
	speed     float64     // movement speed
	spritePos image.Point // sprite sheet coordinates (integer coordinates on spritesheet)
	flipX     bool        // horizontal flip
	dir       int         // 0=LEFT, 1=RIGHT, 2=UP, 3=DOWN
}

// Sprite positions on the spritesheet (x, y) for each direction/animation
var spritePositions = map[string]image.Point{
	"up":        {24, 0},
	"down":      {8, 0},
	"leftIdle":  {56, 0},
	"rightIdle": {56, 0},
	"walkFrame": {40, 0},
}

// Init initializes the game state
func (g *Game) Init() {
	// Initialize player at center with default sprite
	g.pos = p8.NewVector2D(60, 60)
	g.speed = 1
	g.spritePos = spritePositions["leftIdle"]
	g.dir = p8.LEFT
}

// Update handles game logic each frame including input, collision detection, and animation
func (g *Game) Update() {
	isMoving := g.handleMovement()
	g.updateAnimation(isMoving)

}

// getInputDirection reads button presses and returns a movement vector
func (g *Game) getInputDirection() (dx, dy float64, dir int) {
	// Start with current direction to maintain it when not moving
	dir = g.dir

	// Get movement input for both axes
	if p8.Btn(p8.LEFT) {
		dx = -1
	} else if p8.Btn(p8.RIGHT) {
		dx = 1
	}

	if p8.Btn(p8.UP) {
		dy = -1
	} else if p8.Btn(p8.DOWN) {
		dy = 1
	}

	// Only update direction if there's actual movement
	if dx != 0 || dy != 0 {
		// Set direction based on movement, prioritizing horizontal for diagonals
		// For diagonal movement, we'll use the horizontal animation
		if dx != 0 {
			// Horizontal movement takes priority for animation reasons
			if dx < 0 {
				dir = p8.LEFT
			} else {
				dir = p8.RIGHT
			}
		} else if dy != 0 {
			// Only use vertical animation when not moving horizontally
			if dy < 0 {
				dir = p8.UP
			} else {
				dir = p8.DOWN
			}
		}
	}

	return dx, dy, dir
}

// handleMovement processes input and updates player position
// Returns: isMoving (bool)
func (g *Game) handleMovement() (isMoving bool) {
	previousPos := g.pos
	dx, dy, dir := g.getInputDirection()
	g.dir = dir

	// Check if there was movement
	isMoving = dx != 0 || dy != 0

	// Normalize diagonal movement
	if dx != 0 && dy != 0 {
		mag := p8.Sqrt(dx*dx + dy*dy)
		dx, dy = dx/mag, dy/mag
	}

	// Apply movement
	if isMoving {
		// Create a movement vector and add it to position
		moveVec := p8.NewVector2D(dx, dy).Scale(g.speed)
		g.pos = g.pos.Add(moveVec)
	}

	// Check for collision with walls (flag 0) using a 16x16 sprite size
	if p8.MapCollision(g.pos.X, g.pos.Y, 0, 16) {
		g.pos = previousPos // Restore position if collision detected
	}

	return isMoving
}

// updateAnimation updates the player's sprite based on direction and movement
func (g *Game) updateAnimation(isMoving bool) {
	// Animation frame toggle (6 FPS walking animation)
	anim := (p8.Flr(p8.T()*6) % 2) == 0

	// Set default sprite based on direction
	switch g.dir {
	case p8.UP:
		g.spritePos = spritePositions["up"]
		g.flipX = isMoving && anim
	case p8.DOWN:
		g.spritePos = spritePositions["down"]
		g.flipX = isMoving && anim
	case p8.LEFT:
		g.spritePos = spritePositions["leftIdle"]
		g.flipX = false
		if isMoving && anim {
			g.spritePos = spritePositions["walkFrame"]
		}
	case p8.RIGHT:
		g.spritePos = spritePositions["rightIdle"]
		g.flipX = true
		if isMoving && anim {
			g.spritePos = spritePositions["walkFrame"]
		}
	}
}

// Draw renders the game state to the screen
func (g *Game) Draw() {
	p8.Cls(2) // Clear screen with color 2
	p8.Map()  // Draw the map
	p8.Sspr(g.spritePos.X, g.spritePos.Y, 16, 16, g.pos.X, g.pos.Y, 16, 16, g.flipX, false)
}

func main() {
	// Start the game with Game Boy resolution
	settings := p8.NewSettings()
	settings.TargetFPS = 60
	settings.ScreenWidth = 160
	settings.ScreenHeight = 144
	settings.WindowTitle = "Game Boy Style Demo"

	p8.InsertGame(&Game{})
	p8.PlayGameWith(settings)
}
