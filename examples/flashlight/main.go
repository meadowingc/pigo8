// Package main demonstrates a flashing light animation using Sset
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type myGame struct {
	counter float64
	msg     string
}

func (m *myGame) Init() {
	// Initialize the counter
	m.counter = 0
}

func (m *myGame) Update() {
	// Increment the counter
	m.counter += 0.1

	// Create a flashing effect by alternating colors
	switch {
	case m.counter < 1:
		// Red light
		p8.Sset(12, 0, 8) // Red pixel at position (12,0) on the spritesheet
	case m.counter < 2:
		// Blue light
		p8.Sset(12, 0, 12) // Blue pixel at position (12,0) on the spritesheet
	default:
		// Reset counter
		m.counter = 0
	}

	c := p8.Sget(12, 0)
	if c == 8 {
		m.msg = "i"
	} else {
		m.msg = "ou"
	}

}

func (m *myGame) Draw() {
	// Clear screen with dark blue
	p8.Cls(3)

	// Draw sprite 1 (which contains our modified pixel)
	p8.Spr(1, 20, 30)

	p8.Print(m.msg, 5)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
