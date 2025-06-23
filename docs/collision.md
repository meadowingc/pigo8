# Collision Detection in PIGO8

PIGO8 provides collision detection capabilities that allow you to create interactive games with objects that can collide with each other. This guide explains how to use the collision detection system in PIGO8.

## Color-Based Collision Detection

The simplest form of collision detection in PIGO8 is color-based collision detection. This method checks if a pixel at specific coordinates matches a certain color.

### The `ColorCollision` Function

```go
func ColorCollision[X Number, Y Number](x X, y Y, color int) bool
```

**Parameters:**

- `x`: The x-coordinate to check (0-127), can be any numeric type (int, float64, etc.)
- `y`: The y-coordinate to check (0-127), can be any numeric type (int, float64, etc.)
- `color`: The PICO-8 color index to check against (0-15)

**Returns:**

- `bool`: `true` if the pixel at (x, y) matches the specified color, `false` otherwise

If the coordinates are outside the screen bounds (0-127), the function returns `false`.
If the color index is invalid (not 0-15), the function returns `false`.

### Example Usage

```go
// Check if the player has collided with a red wall (color 8)
if p8.ColorCollision(player.x, player.y, 8) {
    // Player is touching a red wall, prevent movement
    player.x = previousX
    player.y = previousY
}
```

## Color Collision Example

PIGO8 includes a complete example of color-based collision detection in the `examples/colorCollision` directory. This example demonstrates how to use the `ColorCollision` function to create a simple game where the player navigates through a maze of colored lines.

### Running the Example

To run the color collision example:

```bash
go run github.com/pigo8/examples/colorCollision@latest
```

Use the arrow keys to move the player (blue dot) through the maze. The player will collide with the colored lines and be prevented from moving through them.

### How the Example Works

The example uses the `ColorCollision` function to check if the player is touching a colored line. If a collision is detected, the player's movement is prevented in that direction.

```go
if p8.Btn(p8.ButtonLeft) {
    g.player.x -= g.player.speed
    if p8.ColorCollision(g.player.x, g.player.y, g.player.collisionColor) {
        g.player.x = beforeX
    }
}
```

This approach allows for simple collision detection without the need for complex collision shapes or physics calculations.

## Tips for Using Color-Based Collision Detection

1. **Multiple Collision Points**: For larger objects, check multiple points around the object for collisions rather than just the center.

2. **Store Previous Position**: Always store the previous position of objects before moving them, so you can revert to that position if a collision is detected.

3. **Separate Horizontal and Vertical Movement**: Check for collisions after horizontal movement and vertical movement separately, which gives better control when moving near corners.

4. **Use Different Colors**: Use different colors for different types of collisions (e.g., walls, hazards, collectibles) to create more complex gameplay.
