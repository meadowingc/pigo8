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
	if color < 0 || color >= len(Pico8Palette) {
		log.Printf("ColorCollision: Invalid color index: %d. Palette has %d colors.", color, len(Pico8Palette))
		return false
	}

	// Get the color at the specified coordinates
	pixelColor := Pget(xInt, yInt)

	// Return true if the pixel color matches the specified color
	return pixelColor == color
}

// MapCollision checks if a tile at the given coordinates has the specified flag set.
// Returns true if the flag is set, false otherwise.
//
// Parameters:
//   - x: The x-coordinate to check, can be any numeric type (will be converted to tile coordinates)
//   - y: The y-coordinate to check, can be any numeric type (will be converted to tile coordinates)
//   - flag: The flag number (0-7) to check
//   - size: (optional) The size of the sprite in pixels (default: 8 for standard PICO-8 sprites)
//
// Returns:
//   - bool: true if the specified flag is set on the sprite at the tile coordinates, false otherwise
//
// This function converts the given pixel coordinates to tile coordinates (dividing by 8),
// gets the sprite number at those tile coordinates using Mget, and then checks if the
// specified flag is set on that sprite using Fget.
//
// For sprites larger than 8x8, it checks multiple points based on the size parameter.
// For example, a 16x16 sprite will check all four corners of the sprite.
//
// Example:
//
//	// Check if an 8x8 sprite is colliding with a wall (flag 0)
//	if MapCollision(sprite.x, sprite.y, Flag0) {
//	    // Sprite is colliding with a wall, handle collision
//	}
//
//	// Check if a 16x16 player sprite is colliding with a wall (flag 0)
//	if MapCollision(player.x, player.y, Flag0, 16) {
//	    // Player is colliding with a wall, handle collision
//	}
//
//	// Save original position before moving
//	origX, origY := player.x, player.y
//
//	// Apply movement
//	player.x += dx
//	player.y += dy
//
//	// Check for collision with a 16x16 sprite and restore position if needed
//	if MapCollision(player.x, player.y, Flag0, 16) {
//	    player.x, player.y = origX, origY
//	}
func MapCollision[X Number, Y Number](x X, y Y, flag int, size ...int) bool {
	// Default sprite size is 8x8 (standard PICO-8 sprite)
	spriteSize := 8

	// If size is provided, use it instead
	if len(size) > 0 && size[0] > 0 {
		spriteSize = size[0]
	}

	// Convert coordinates to float64 for calculations
	fx := float64(x)
	fy := float64(y)

	// For 8x8 sprites, just check the top-left corner
	if spriteSize <= 8 {
		// Convert coordinates to tile coordinates (divide by 8)
		tileX := Flr(fx / 8)
		tileY := Flr(fy / 8)

		// Get sprite number from map
		sprite := Mget(tileX, tileY)

		// Get flags of that sprite and check if the specific flag is set
		_, isSet := Fget(sprite, flag)

		return isSet
	}

	// For 16x16 sprites (common case), check specific points instead of all corners
	if spriteSize == 16 {
		// Check the four corners and the center of each edge
		checkPoints := [][2]float64{
			{fx, fy},           // Top-left corner
			{fx + 15, fy},      // Top-right corner
			{fx, fy + 15},      // Bottom-left corner
			{fx + 15, fy + 15}, // Bottom-right corner
			{fx + 7, fy},       // Top middle
			{fx, fy + 7},       // Left middle
			{fx + 15, fy + 7},  // Right middle
			{fx + 7, fy + 15},  // Bottom middle
		}

		// Check each point
		for _, point := range checkPoints {
			checkX, checkY := point[0], point[1]

			// Convert to tile coordinates
			tileX := Flr(checkX / 8)
			tileY := Flr(checkY / 8)

			// Get sprite number from map
			sprite := Mget(tileX, tileY)

			// Get flags of that sprite and check if the specific flag is set
			_, isSet := Fget(sprite, flag)

			// If any point collides, return true
			if isSet {
				return true
			}
		}

		// No collision found for 16x16 sprite
		return false
	}

	// For other sizes, check a grid of points
	// Calculate how many 8x8 tiles we need to check in each dimension
	tileCount := (spriteSize + 7) / 8 // Ceiling division to get tile count

	// Check each corner of the sprite
	for i := 0; i < tileCount; i++ {
		for j := 0; j < tileCount; j++ {
			// Calculate the position to check
			checkX := fx + float64(i*8)
			checkY := fy + float64(j*8)

			// Skip points outside the sprite's bounds
			if checkX >= fx+float64(spriteSize) || checkY >= fy+float64(spriteSize) {
				continue
			}

			// Convert to tile coordinates
			tileX := Flr(checkX / 8)
			tileY := Flr(checkY / 8)

			// Get sprite number from map
			sprite := Mget(tileX, tileY)

			// Get flags of that sprite and check if the specific flag is set
			_, isSet := Fget(sprite, flag)

			// If any point collides, return true
			if isSet {
				return true
			}
		}
	}

	// No collision found
	return false
}
