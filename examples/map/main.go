// Package main demonstrates the map functionality of the PICO-8 engine
package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type myGame struct{}

func (m *myGame) Init() {
}

func (m *myGame) Update() {
}

func (m *myGame) Draw() {
	p8.Cls(0)
	p8.Map(18, 7, 40, 70, 5, 4)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
