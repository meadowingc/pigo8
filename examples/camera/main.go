// Package main demonstrates the camera functionality of the PICO-8 engine
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

// Game holds the state of our game
type Game struct {
}

// Init is called once at the beginning of the game
func (m *Game) Init() {
}

// Update is called once per frame
func (m *Game) Update() {
}

// Draw is called once per frame
func (m *Game) Draw() {
	p8.Cls()
	p8.Camera()
	p8.Rectfill(0, 0, 127, 127, 2)
	p8.Rect(0, 0, 127, 127, 8)
	p8.Print("camera(0,0)", 2, 2)

	p8.Camera(63, 63)
	p8.Rect(63, 63, 127+63, 127+63, 11)
	p8.Print("camera(63,63)", 136, 182)

}

func main() {
	p8.InsertGame(&Game{})
	p8.Play()
}
