package main

import p8 "github.com/drpaneas/pigo8"

type myGame struct{}

func (m *myGame) Init() {}

func (m *myGame) Update() {}

func (m *myGame) Draw() {
	p8.Cls(1)
	p8.Print("hello, world!", 40, 60)
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
