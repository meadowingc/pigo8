# PIGO8 Documentation

Welcome to the PIGO8 documentation. This directory contains detailed information about the PIGO8 library's features, functions, and usage examples.

## Table of Contents

### Core Features
- [Palette Management](palette-management.md) - Dynamic color palette management
- [Transparency](transparency.md) - Color transparency control

## Getting Started

PIGO8 is a Go library inspired by PICO-8, designed to make 2D game development simple and fun. It provides a set of functions for drawing, input handling, sound, and more.

### Basic Game Structure

```go
package main

import "github.com/drpaneas/pigo8"

// Game represents our game state
type Game struct {
    // Your game state variables here
}

// Init initializes the game
func (g *Game) Init() {
    // Initialization code
}

// Update updates the game state
func (g *Game) Update() {
    // Update game logic
}

// Draw draws the game
func (g *Game) Draw() {
    // Drawing code
}

func main() {
    game := &Game{}
    
    // Insert the game into PIGO8
    pigo8.InsertGame(game)
    
    // Configure settings
    settings := pigo8.NewSettings()
    settings.WindowTitle = "My PIGO8 Game"
    settings.ScaleFactor = 4
    
    // Run the game with the configured settings
    pigo8.PlayGameWith(settings)
}
```

## Examples

Check out the `examples` directory in the repository for complete, working examples of PIGO8 games and demos.

## Contributing

Contributions to PIGO8 and its documentation are welcome! Please feel free to submit pull requests or open issues on the GitHub repository.
