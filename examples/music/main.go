package main

//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .

import (
	"log"

	p8 "github.com/drpaneas/pigo8"
)

// Game implements the PIGO8 Cartridge interface
type Game struct {
	frame int
}

// Init is called once at the start of the game
func (g *Game) Init() {
	log.Println("Music example initialized")
}

// Update is called every frame for game logic
func (g *Game) Update() {
	g.frame++

	// Play music 1 when space is pressed
	if p8.Btn(p8.UP) {
		log.Println("Playing music 1")
		p8.Music(1)
	}

	// Play music 2 when Z is pressed
	if p8.Btn(p8.DOWN) {
		log.Println("Playing music 2")
		p8.Music(2)
	}

	// Play music 3 exclusively (stops other music) when C is pressed
	if p8.Btn(p8.LEFT) {
		log.Println("Playing music 3 exclusively")
		p8.Music(3, true)
	}

	// Stop all music when V is pressed
	if p8.Btn(p8.RIGHT) {
		log.Println("Stopping all music")
		p8.Music(-1)
	}
}

// Draw is called every frame for rendering
func (g *Game) Draw() {
	// Clear the screen with dark blue
	p8.Cls(1)

	// Draw instructions
	p8.Print("Music Example", 30, 10, 7)
	p8.Print("Up to play 1", 10, 30, 7)
	p8.Print("Down to play 2", 10, 40, 7)
	p8.Print("Left to play only 3", 10, 50, 7)
	p8.Print("Right to stop all", 10, 60, 7)

}

func main() {
	// Create and insert our game
	game := &Game{}
	p8.InsertGame(game)

	// Configure the PIGO8 console
	settings := p8.NewSettings()
	settings.WindowTitle = "PIGO8 Music Example"
	settings.ScaleFactor = 4

	// Start the game
	p8.PlayGameWith(settings)
}
