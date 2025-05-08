package pigo8

import "log"

// ColorCollision checks if the pixel at coordinates (x, y) matches the specified color.
// Returns true if the pixel color matches the given color, false otherwise.
//
// Parameters:
//   - x: The x-coordinate to check (0-127), can be any numeric type
//   - y: The y-coordinate to check (0-127), can be any numeric type
//   - color: The PICO-8 color index to check against (0-15)
//
// Returns:
//   - bool: true if the pixel at (x, y) matches the specified color, false otherwise
//
// If the coordinates are outside the screen bounds (0-127), the function returns false.
// If the color index is invalid (not 0-15), the function returns false.
//
// Example:
//
//	// Check if the pixel at (10, 20) is red (color 8)
//	if ColorCollision(10, 20, 8) {
//	    // Pixel is red, handle collision
//	}
//
//	// Check if the player is touching a wall (color 3) using float coordinates
//	if ColorCollision(player.x, player.y, 3) {
//	    // Player is touching a wall, prevent movement
//	}
func ColorCollision[X Number, Y Number](x X, y Y, color int) bool {
	// Convert coordinates to int
	xInt := int(float64(x))
	yInt := int(float64(y))

	// Validate coordinates are within screen bounds
	if xInt < 0 || xInt >= ScreenWidth || yInt < 0 || yInt >= ScreenHeight {
		log.Println("ColorCollision: Invalid coordinates:", xInt, yInt)
		return false
	}

	// Validate color index
	if color < 0 || color > 15 {
		log.Println("ColorCollision: Invalid color index:", color)
		return false
	}

	// Get the color at the specified coordinates
	pixelColor := Pget(xInt, yInt)

	// Return true if the pixel color matches the specified color
	return pixelColor == color
}
