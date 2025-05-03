// Package main basic hello world example using clear screen and print statements
package main

import p8 "github.com/drpaneas/pigo8"

type myGame struct{}

func (m *myGame) Init() {}

func (m *myGame) Update() {}

func (m *myGame) Draw() {
	p8.Cls(1)
	sx := 8
	sy := 0
	sw := 16
	sh := 16
	dx := 10
	dy := 20
	p8.Sspr(sx, sy, sw, sh, dx, dy)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
