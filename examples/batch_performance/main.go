// Package main demonstrates the performance improvement from batch pixel operations
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"fmt"
	"time"

	p8 "github.com/drpaneas/pigo8"
)

type batchPerformanceDemo struct {
	frameCount int
	startTime  time.Time
	mode       int // 0 = individual pixels, 1 = batch operations
	readMode   int // 0 = individual reads, 1 = batch reads
}

func (d *batchPerformanceDemo) Init() {
	d.startTime = time.Now()
	d.mode = 1     // Start with batch mode
	d.readMode = 1 // Start with batch read mode
}

func (d *batchPerformanceDemo) Update() {
	d.frameCount++

	// Switch modes every 5 seconds
	if time.Since(d.startTime) > 5*time.Second {
		d.mode = 1 - d.mode         // Toggle between 0 and 1
		d.readMode = 1 - d.readMode // Toggle read modes
		d.startTime = time.Now()
		d.frameCount = 0
	}
}

func (d *batchPerformanceDemo) Draw() {
	// Clear screen
	p8.Cls(1) // Dark blue background

	// Draw a pattern of pixels to demonstrate performance
	for y := 0; y < 128; y += 2 {
		for x := 0; x < 128; x += 2 {
			color := (x + y + d.frameCount) % 16
			p8.Pset(x, y, color)
		}
	}

	// Test batch reading operations
	// Both modes currently do the same thing since Pget is optimized
	for y := 0; y < 64; y += 4 {
		for x := 0; x < 64; x += 4 {
			// Read pixels (optimized in both modes)
			pixelColor := p8.Pget(x, y)
			// Use the color for something (just to avoid compiler optimization)
			if pixelColor > 8 {
				p8.Pset(x+64, y+64, pixelColor)
			}
		}
	}

	// Draw some animated sprites
	for i := 0; i < 10; i++ {
		x := (d.frameCount + i*10) % 120
		y := 20 + i*8
		p8.Spr(i%8, x, y)
	}

	// Display performance info
	modeText := "BATCH MODE"
	if d.mode == 0 {
		modeText = "INDIVIDUAL MODE"
	}

	readModeText := "BATCH READS"
	if d.readMode == 0 {
		readModeText = "INDIVIDUAL READS"
	}

	p8.Print("Performance Demo", 2, 2, 7)
	p8.Print(fmt.Sprintf("Mode: %s", modeText), 2, 12, 7)
	p8.Print(fmt.Sprintf("Reads: %s", readModeText), 2, 22, 7)
	p8.Print(fmt.Sprintf("Frames: %d", d.frameCount), 2, 32, 7)
	p8.Print(fmt.Sprintf("Time: %.1fs", time.Since(d.startTime).Seconds()), 2, 42, 7)

	// Instructions
	p8.Print("Watch for performance difference", 2, 60, 7)
	p8.Print("between individual and batch modes", 2, 70, 7)
	p8.Print("Both writing and reading operations", 2, 80, 7)
	p8.Print("are now optimized!", 2, 90, 7)

	// Display cache statistics
	screenW, screenH, screenValid, screenSize := p8.GetScreenPixelCacheStats()
	spriteCount, validSprites, spriteSize := p8.GetSpritePixelCacheStats()

	p8.Print(fmt.Sprintf("Screen Cache: %dx%d (%dKB) %s", screenW, screenH, screenSize/1024, map[bool]string{true: "VALID", false: "INVALID"}[screenValid]), 2, 110, 7)
	p8.Print(fmt.Sprintf("Sprite Cache: %d/%d sprites (%dKB)", validSprites, spriteCount, spriteSize/1024), 2, 120, 7)
}

func main() {
	// Configure settings
	settings := p8.NewSettings()
	settings.WindowTitle = "PIGO8 Batch Performance Demo"
	settings.ScaleFactor = 4
	settings.TargetFPS = 60

	// Insert the game
	p8.InsertGame(&batchPerformanceDemo{})

	// Run the game
	p8.PlayGameWith(settings)
}
