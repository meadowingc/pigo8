// Package main basic hello world example using clear screen and print statements
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import p8 "github.com/drpaneas/pigo8"

type myGame struct{}

func (m *myGame) Init() {}

func (m *myGame) Update() {}

func (m *myGame) Draw() {
	p8.Cls(0)
	p8.Spr(2, 20, 22)
	p8.Spr(3, 28, 22)
	p8.Spr(34, 20, 30)
	p8.Spr(35, 28, 30)

	p8.Sspr(16, 0, 16, 16, 50, 50)
	p8.Sspr(64, 56, 32, 32, 80, 80)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
