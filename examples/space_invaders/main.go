// Package space_invaders is a simple implementation of the classic Space Invaders game
// using the PIGO8 game engine. This example demonstrates basic game mechanics
// including player movement, shooting, collision detection, and sprite rendering.
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"fmt"

	"github.com/drpaneas/pigo8"
)

// Note: We're using basic drawing functions since Sprite functions aren't available in the public API

// Game represents the current state of the Space Invaders game.
// It manages the player, aliens, bullets, and game state.
type Game struct {
	playerX      int
	playerY      int
	bullets      []bullet
	aliens       []alien
	alienBullets []bullet
	score        int
	lives        int
	gameOver     bool
	paused       bool
	menuItem     int // 0 = resume, 1 = quit
}

// bullet represents a projectile in the game, which can be fired by either the player or aliens.
type bullet struct {
	x, y  int
	speed int
}

// alien represents an enemy in the game with position, status, and sprite information.
type alien struct {
	x, y   int
	alive  bool
	sprite int
}

// NewGame initializes and returns a new instance of the Space Invaders game
// with default starting values for the player, score, and lives.
func NewGame() *Game {
	return &Game{
		playerX: 64,
		playerY: 120,
		score:   0,
		lives:   7,
	}
}

// Init is called once at the start of the game
func (g *Game) Init() {
	g.initAliens()
}

func (g *Game) initAliens() {
	g.aliens = nil
	// Create 5 rows of 11 aliens each
	for row := 0; row < 5; row++ {
		for col := 0; col < 11; col++ {
			g.aliens = append(g.aliens, alien{
				x:      16 + col*10,
				y:      16 + row*10,
				alive:  true,
				sprite: (row % 3) + 1, // Cycle through 3 alien sprites
			})
		}
	}
}

// processInputs handles player inputs for pausing, menu navigation, game reset, movement, and shooting.
// It returns false if the game update loop should be short-circuited (e.g., game is paused or over).
func (g *Game) processInputs() bool {
	if pigo8.Btnp(pigo8.START) {
		g.paused = !g.paused
	}

	if g.paused {
		if pigo8.Btnp(pigo8.UP) || pigo8.Btnp(pigo8.LEFT) {
			g.menuItem = (g.menuItem - 1 + 2) % 2
		}
		if pigo8.Btnp(pigo8.DOWN) || pigo8.Btnp(pigo8.RIGHT) {
			g.menuItem = (g.menuItem + 1) % 2
		}
		if pigo8.Btnp(pigo8.O) {
			switch g.menuItem {
			case 0: // Resume
				g.paused = false
			case 1: // Quit
				g.gameOver = true // Or os.Exit(0) if you want to close the game
			}
		}
		return false // Don't continue main update loop if paused
	}

	if g.gameOver {
		if pigo8.Btnp(pigo8.O) {
			g.resetGame()
		}
		return false // Don't continue main update loop if game is over
	}

	// Player movement
	if pigo8.Btn(pigo8.LEFT) && g.playerX > 8 {
		g.playerX -= 2
	}
	if pigo8.Btn(pigo8.RIGHT) && g.playerX < 120 {
		g.playerX += 2
	}

	// Player shooting
	if pigo8.Btnp(pigo8.O) {
		g.bullets = append(g.bullets, bullet{
			x:     g.playerX,
			y:     g.playerY - 8,
			speed: 4,
		})
	}
	return true // Continue with the main update loop
}

// updatePlayerBullets updates the position of player bullets and removes them if they go off-screen.
func (g *Game) updatePlayerBullets() {
	for i := 0; i < len(g.bullets); {
		g.bullets[i].y -= g.bullets[i].speed
		if g.bullets[i].y < 0 {
			g.bullets = append(g.bullets[:i], g.bullets[i+1:]...)
			// No increment for i here, as the slice has shifted
		} else {
			i++
		}
	}
}

// updateAliensAndAlienBullets handles alien behavior, including shooting, and updates alien bullets.
func (g *Game) updateAliensAndAlienBullets() {
	// Alien movement and shooting
	for i := range g.aliens {
		if !g.aliens[i].alive {
			continue
		}

		// Simple AI: random chance to shoot
		if pigo8.Rnd(100) == 0 {
			g.alienBullets = append(g.alienBullets, bullet{
				x:     g.aliens[i].x,
				y:     g.aliens[i].y + 8,
				speed: 2,
			})
		}

		// Check if aliens reached bottom
		if g.aliens[i].y > 110 { // Assuming screen height or player level makes this game over
			g.gameOver = true
			return // Game over, no need to process further alien logic this frame
		}
	}

	// Update alien bullets and check for collision with player
	for i := 0; i < len(g.alienBullets); {
		bullet := &g.alienBullets[i]
		bullet.y += bullet.speed

		if bullet.y > 128 { // Off-screen
			g.alienBullets = append(g.alienBullets[:i], g.alienBullets[i+1:]...)
			continue
		}

		// Check collision with player
		if bullet.x > g.playerX-4 && bullet.x < g.playerX+8 &&
			bullet.y > g.playerY-8 && bullet.y < g.playerY+8 {
			g.lives--
			pigo8.Music(1)                                                       // Play hit sound (using Music function)
			g.alienBullets = append(g.alienBullets[:i], g.alienBullets[i+1:]...) // Remove bullet
			if g.lives <= 0 {
				g.gameOver = true
				return // Player is dead, game over
			}
			continue // Bullet hit, removed, process next
		}
		i++
	}
}

// checkPlayerBulletCollisionsAndWin checks for collisions between player bullets and aliens,
// and checks if all aliens are defeated.
func (g *Game) checkPlayerBulletCollisionsAndWin() {
	// Check collision between player bullets and aliens
	for i := 0; i < len(g.bullets); {
		bullet := g.bullets[i]
		hit := false
		for j := range g.aliens {
			if g.aliens[j].alive &&
				bullet.x > g.aliens[j].x-4 && bullet.x < g.aliens[j].x+8 &&
				bullet.y > g.aliens[j].y-8 && bullet.y < g.aliens[j].y+8 {
				g.aliens[j].alive = false
				g.score += 10
				pigo8.Music(0)                                        // Play explosion sound (using Music function)
				g.bullets = append(g.bullets[:i], g.bullets[i+1:]...) // Remove bullet
				hit = true
				break // Bullet can only hit one alien
			}
		}
		if !hit {
			i++ // Only increment if bullet was not removed
		}
	}

	// Check if all aliens are defeated
	allAliensDead := true
	for _, alien := range g.aliens {
		if alien.alive {
			allAliensDead = false
			break
		}
	}
	if allAliensDead {
		g.initAliens() // Reset aliens for a new wave
	}
}

// Update progresses the game state by one frame.
func (g *Game) Update() {
	if !g.processInputs() {
		return // Game is paused or over, or an action was taken in menus
	}

	g.updatePlayerBullets()
	g.updateAliensAndAlienBullets()
	if g.gameOver { // Check if game became over during alien/bullet updates
		return
	}
	g.checkPlayerBulletCollisionsAndWin()
}

// Draw renders all game objects to the screen, including the player, aliens, bullets, and UI elements.
// It's called every frame to update the visual representation of the game state.
func (g *Game) Draw() {
	// Clear screen with black
	pigo8.Cls(0)

	if g.paused {
		// Draw pause menu
		pigo8.Print("paused", 50, 40, 7)

		// Draw menu items with selection indicator
		if g.menuItem == 0 {
			pigo8.Print("> resume", 45, 60, 7)
		} else {
			pigo8.Print("  resume", 45, 60, 7)
		}

		if g.menuItem == 1 {
			pigo8.Print("> quit", 45, 70, 7)
		} else {
			pigo8.Print("  quit", 45, 70, 7)
		}
		return
	}

	// Draw player (triangle using lines)
	// Top point to bottom left
	pigo8.Line(g.playerX+4, g.playerY-8, g.playerX, g.playerY, 7)
	// Top point to bottom right
	pigo8.Line(g.playerX+4, g.playerY-8, g.playerX+8, g.playerY, 7)
	// Bottom line
	pigo8.Line(g.playerX, g.playerY, g.playerX+8, g.playerY, 7)

	// Draw bullets
	for _, b := range g.bullets {
		pigo8.Rectfill(b.x, b.y, b.x+2, b.y+4, 7) // White bullets
	}

	// Draw alien bullets
	for _, b := range g.alienBullets {
		pigo8.Rectfill(b.x, b.y, b.x+2, b.y+4, 8) // Red bullets
	}

	// Draw aliens as simple shapes
	for _, a := range g.aliens {
		if a.alive {
			// Draw alien as a square with a different color for each type
			color := 2 + a.sprite // Different colors for different alien types
			pigo8.Rect(a.x, a.y, a.x+8, a.y+8, color)
			pigo8.Rectfill(a.x+2, a.y+2, a.x+6, a.y+6, color)
		}
	}

	// Draw UI
	pigo8.Print("score: "+fmt.Sprint(g.score), 4, 4, 7)
	pigo8.Print("lives: "+fmt.Sprint(g.lives), 80, 4, 7)

	if g.gameOver {
		pigo8.Print("game over", 40, 60, 7)
		pigo8.Print("press o to restart", 20, 70, 7)
	}
}

func (g *Game) resetGame() {
	g.playerX = 64
	g.playerY = 120
	g.bullets = nil
	g.alienBullets = nil
	g.score = 0
	g.lives = 7
	g.gameOver = false
	g.paused = false
	g.menuItem = 0
	g.initAliens()
}

func main() {
	game := NewGame()
	pigo8.InsertGame(game)
	pigo8.Play()
}
