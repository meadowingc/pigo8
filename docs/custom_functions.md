# Custom Functions

PIGO8 extends the original PICO-8 API with additional custom functions that provide enhanced capabilities for your games. These functions are not part of the official PICO-8, but they follow the same design philosophy and integrate seamlessly with the rest of the PIGO8 API.

## Available Custom Functions

* **Color Collision Detection**: Detect collisions between sprites based on their non-transparent pixels
* **Map Collision Detection**: Detect collisions with map tiles using flags
* **Flag Constants**: Pre-defined constants for easier flag operations
* **Sget/Sset Functions**: Get and set individual pixels on the spritesheet
* **Alpha Transparency**: Create semi-transparent colors with alpha values
* **Fade System**: Create smooth transitions between scenes or palettes

## Why Custom Functions?

While PIGO8 aims to recreate the feel of PICO-8, it also takes advantage of Go's capabilities to provide additional features that can make game development easier and more powerful. These custom functions:

1. Solve common game development challenges
2. Reduce boilerplate code
3. Enable effects and gameplay mechanics that would be difficult to implement otherwise
4. Maintain the spirit of PICO-8 while extending its capabilities

## Using Custom Functions

Custom functions follow the same naming conventions and design patterns as the standard PIGO8 functions. They can be used alongside the standard functions without any special setup.

```go
// Example using both standard and custom functions
func (g *Game) Update() {
    // Standard PICO-8 function to move the player
    if p8.Btn(p8.ButtonRight) {
        g.playerX++
    }
    
    // Custom function to check for collision with map
    if p8.MapCollision(g.playerX, g.playerY, 0) {
        // Handle collision
    }
}
```

Refer to the specific function documentation pages for detailed usage examples and implementation details.
