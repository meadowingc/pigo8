package main

import (
	p8 "github.com/drpaneas/pigo8"
)

type Entity struct {
	sprite, x, y, timing, speed, first, last float64
}

func NewEntity(sprite, x, y, timing, speed, first, last float64) Entity {
	return Entity{
		sprite: sprite,
		timing: timing,
		first:  first,
		last:   last,
		x:      x,
		y:      y,
		speed:  speed,
	}
}

func (ae *Entity) Animate() {
	ae.sprite += ae.timing
	if ae.sprite >= ae.last {
		ae.sprite = ae.first
	}
}

func (ae *Entity) Move(offset float64) {
	ae.x += offset
	if ae.x > 128 {
		ae.x = -8
	}
}

func (ae *Entity) Draw() {
	p8.Spr(ae.sprite, ae.x, ae.y)
}

var player Entity
var enemies = []Entity{}
var items = []Entity{}

type myGame struct{}

func (m *myGame) Init() {
	player = NewEntity(1, -8, 59, 0.25, 1, 1, 5)
	enemy1 := NewEntity(5, -20, 5, 0.1, 1.25, 5, 9)
	enemy2 := NewEntity(9, -14, 30, 0.2, 0.4, 9, 13)
	enemy3 := NewEntity(13, -11, 90, 0.4, 0.75, 13, 17)
	enemies = append(enemies, enemy1, enemy2, enemy3)
	item1 := NewEntity(48, 30, 110, 0.3, 48, 50, 56)
	item2 := NewEntity(56, 60, 110, 0.25, 54, 56, 60)
	item3 := NewEntity(60, 90, 110, 0.15, 4, 60, 64)
	items = append(items, item1, item2, item3)
}

func (m *myGame) Update() {
	player.Animate()
	player.Move(player.speed)

	for i := range enemies {
		enemies[i].Animate()
		enemies[i].Move(enemies[i].speed)
	}

	for i := range items {
		items[i].Animate()
	}
}

func (m *myGame) Draw() {
	p8.Cls(0)
	player.Draw()

	for _, enemy := range enemies {
		enemy.Draw()
	}

	for _, item := range items {
		item.Draw()
	}
}

func main() {
	p8.InsertGame(&myGame{})
	p8.Play()
}
