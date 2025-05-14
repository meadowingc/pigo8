# Color Collision Detection

Color collision detection is a custom function in PIGO8 that allows you to check if a pixel at specific coordinates matches a specified color. This is useful for color-based collision detection in games.

## Overview

The `ColorCollision` function checks if the pixel at the given coordinates (x, y) matches a specific color from the PICO-8 palette. This can be used for various collision detection scenarios, such as checking if a character is touching a specific terrain type or obstacle represented by a particular color.

## Function Signature

```go
func ColorCollision[X Number, Y Number](x X, y Y, color int) bool
```

### Parameters

* `x`: The x-coordinate to check (0-127), can be any numeric type
* `y`: The y-coordinate to check (0-127), can be any numeric type
* `color`: The PICO-8 color index to check against (0-15)

### Return Value

* `bool`: Returns `true` if the pixel at (x, y) matches the specified color, `false` otherwise

## Example Usage

```go
package main

import p8 "github.com/drpaneas/pigo8"

type Game struct {
    playerX, playerY float64
    wallColor        int
}

func (g *Game) Init() {
    g.playerX = 64
    g.playerY = 64
    g.wallColor = 3 // Assuming color 3 represents walls
}

func (g *Game) Update() {
    // Store original position
    origX, origY := g.playerX, g.playerY
    
    // Move player based on input
    if p8.Btn(p8.ButtonRight) {
        g.playerX++
    }
    
    // Check if the player is touching a wall (color 3)
    if p8.ColorCollision(g.playerX, g.playerY, g.wallColor) {
        // Player is touching a wall, revert to previous position
        g.playerX, g.playerY = origX, origY
        
        // Play collision sound
        p8.Music(0)
    }
}
```

## How It Works

The color collision detection function:

1. Validates that the coordinates are within the screen bounds (0-127)
2. Validates that the color index is valid (0-15)
3. Gets the color of the pixel at the specified coordinates using `Pget`
4. Compares the pixel color with the specified color
5. Returns true if they match, false otherwise

## Use Cases

Color collision detection is useful for:

* Terrain-based collision (e.g., detecting water, lava, or solid ground)
* Color-coded obstacle detection
* Pixel-perfect collision in games with detailed environments
* Detecting when a character enters specific areas marked by color

## Performance Considerations

For optimal performance:

* Use this function sparingly, as checking individual pixels can be CPU-intensive
* Consider checking only key points of your game objects (e.g., corners or center) rather than every pixel
* For larger objects, combine with bounding box checks first

## Complete Example

You can find a complete example of color collision detection in the [examples/colorCollision](https://github.com/drpaneas/pigo8/tree/main/examples/colorCollision) directory.
