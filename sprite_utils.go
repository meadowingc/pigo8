package pigo8

// calculateSpriteID calculates the sprite ID based on the sprite's position in the grid
// using the dynamic sprite sheet dimensions.
// This function is used by Spr, Sspr, Sget, and Sset to ensure consistent sprite ID calculation
// when custom palettes change the sprite sheet dimensions.
func calculateSpriteID(spriteX, spriteY int) int {
	return spriteY*spritesheetColumns + spriteX
}

// validateSpriteSheetBounds checks if the given coordinates are within the sprite sheet bounds
// and returns true if they are valid, false otherwise.
func validateSpriteSheetBounds(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x+width <= spritesheetWidth && y+height <= spritesheetHeight
}
