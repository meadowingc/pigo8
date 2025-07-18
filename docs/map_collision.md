# Map Collision Detection

Map collision detection is a custom function in PIGO8 that allows you to detect collisions between game objects and map tiles based on the tiles' flag values.

## Overview

In many games, you need to check if a character or object is colliding with solid elements in the game map, such as walls, platforms, or obstacles. The `MapCollision` function simplifies this process by checking if a point or sprite overlaps with map tiles that have specific flags set.

## Function Signature

```go
func MapCollision[X Number, Y Number](x X, y Y, flag int, size ...int) bool
```

### Parameters

* `x`: The x-coordinate to check, can be any numeric type (will be converted to tile coordinates)
* `y`: The y-coordinate to check, can be any numeric type (will be converted to tile coordinates)
* `flag`: The flag number (0-7) to check
* `size`: (optional) The size of the sprite in pixels (default: 8 for standard PICO-8 sprites)

### Return Value

* `bool`: Returns `true` if the specified flag is set on the sprite at the tile coordinates, `false` otherwise

## Example Usage

```go
package main

import p8 "github.com/drpaneas/pigo8"

type Game struct {
    playerX, playerY float64
    playerSize       int
}

func (g *Game) Init() {
    g.playerX = 64
    g.playerY = 64
    g.playerSize = 16 // 16x16 player sprite
}

func (g *Game) Update() {
    // Store the current position
    prevX, prevY := g.playerX, g.playerY
    
    // Move player based on input
    if p8.Btn(p8.RIGHT) {
        g.playerX++
    }
    
    // Check for collision with solid map tiles (using flag 0 for solid tiles)
    if p8.MapCollision(g.playerX, g.playerY, 0, g.playerSize) {
        // Collision detected, revert to previous position
        g.playerX, g.playerY = prevX, prevY
    }
}
```

## Setting Up Map Flags

To use map collision detection, you need to set up flags for your map tiles:

1. In the PIGO8 editor, select a sprite and toggle the flags you want to set
2. Use these sprites in your map
3. In your game code, check for collisions with specific flags

For example, you might use:

* Flag0 for solid/blocking tiles
* Flag1 for damage tiles (spikes, lava)
* Flag2 for collectible tiles
* Flag3 for special interaction tiles

## Multiple Flag Checks

You can check for multiple flags by combining them with bitwise OR:

```go
// Check if a tile has either Flag0 OR Flag1 set
if p8.MapRectCollision(g.playerX, g.playerY, g.playerWidth, g.playerHeight, p8.Flag0|p8.Flag1) {
    // Handle collision with either solid tiles or damage tiles
}
```

## Complete Example

You can find a complete example of map collision detection in the [examples/map_layers](https://github.com/drpaneas/pigo8/tree/main/examples/map_layers) directory.
