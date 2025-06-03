package pigo8

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/exp/constraints"
)

// SpriteInfo definition is now expected in spritesheet.go within this package

// Number covers ints, floats
type Number interface {
	constraints.Integer | constraints.Float
}

// Note: currentScreen, currentSprites, and currentDrawColor are defined in engine.go

// Spr draws a potentially fractional rectangular region of sprites,
// using the internal `currentScreen` and `currentSprites` variables.
//
// The x and y coordinates can be any integer or float type (e.g., int, float64)
// due to the use of generics [X Number, Y Number]. They are converted internally
// to float64 for drawing calculations.
//
// screen:          REMOVED (uses internal currentScreen)
// sprites:         REMOVED (uses internal currentSprites)
// spriteNumber:    The index (int) for the top-left sprite of the block.
// x:               Screen X coordinate (any Number type) for the top-left corner.
// y:               Screen Y coordinate (any Number type) for the top-left corner.
// options...:      Optional parameters (w, h, flipX, flipY)
//   - w (float64 or int): Width multiplier (default 1.0). Handled via interface{}.
//   - h (float64 or int): Height multiplier (default 1.0). Handled via interface{}.
//   - flipX (bool):       Flip horizontally (default false). Handled via interface{}.
//   - flipY (bool):       Flip vertically (default false). Handled via interface{}.
//
// Usage:
//
//	Spr(spriteNumber, x, y)
//	Spr(spriteNumber, x, y, w, h)
//	Spr(spriteNumber, x, y, w, h, flipX)
//	Spr(spriteNumber, x, y, w, h, flipX, flipY)
//
// Example:
//
//	var ix, iy int = 10, 20
//	var fx, fy float64 = 30.5, 20.0
//
//	// Draw sprite 1 at (10, 20) using int coordinates
//	Spr(1, ix, iy)
//
//	// Draw sprite 1 at (30.5, 20.0) using float64 coordinates
//	Spr(1, fx, fy)
//
//	// Draw sprite 1 at (10, 20.0) using mixed int/float64 coordinates
//	Spr(1, ix, fy)
//
//	// Draw sprite 1 and the left half of sprite 2 (w=1.5)
//	Spr(1, 50, 20, 1.5, 1.0)
//
//	// Draw a 1.5w x 1.5h block starting at sprite 0
//	Spr(0, 70, 20, 1.5, 1.5)
//
//	// Draw the same 1.5 x 1.5 block, flipped horizontally
//	Spr(0, 90, 20, 1.5, 1.5, true)
//
//	// Draw sprite 0 using a float sprite number (truncated to 0)
//	Spr(0.7, 110, 20)
//
//	// Explicitly specify generic types if needed (rarely necessary)
//	Spr[int, float64](1, 10, 20.5)
//
//	// Explicitly specify all generic types
//	Spr[float64, int, float64](1.2, 10, 20.5) // spriteNumber becomes 1
func Spr[SN Number, X Number, Y Number](spriteNumber SN, x X, y Y, options ...any) {
	// Convert generic spriteNumber, x, y to required types
	spriteNumInt := int(spriteNumber) // Cast sprite number to int
	fx := float64(x)
	fy := float64(y)

	// Apply camera offset before using coordinates for drawing
	screenFx, screenFy := applyCameraOffset(fx, fy)

	// Use internal package variables set by engine.Draw
	if currentScreen == nil {
		log.Println("Warning: Spr() called before screen was ready.")
		return
	}

	// --- Lazy Loading Logic ---
	if currentSprites == nil {
		loaded, err := loadSpritesheet() // Call the loading function from spritesheet.go
		if err != nil {
			log.Fatalf("Fatal: Failed to load required spritesheet for Spr(): %v", err)
		}
		currentSprites = loaded // Store successfully loaded sprites
	}

	// Find the sprite by ID or index
	spriteInfo := findSpriteByID(spriteNumInt)
	if spriteInfo == nil {
		// No sprite found with this ID or at this index
		return
	}

	// Parse optional arguments
	scaleW, scaleH, flipX, flipY := parseSprOptions(options)

	// Get sprite dimensions
	tileImage := spriteInfo.Image
	spriteWidth := float64(tileImage.Bounds().Dx())
	spriteHeight := float64(tileImage.Bounds().Dy())

	// Create a transparent version of the sprite
	tempImage := createTransparentSpriteImage(tileImage)

	// Calculate final dimensions
	destWidth := spriteWidth * scaleW
	destHeight := spriteHeight * scaleH

	// Setup drawing options
	opts := setupDrawOptions(screenFx, screenFy, destWidth, destHeight, scaleW, scaleH, flipX, flipY)

	// Draw the sprite
	currentScreen.DrawImage(tempImage, opts)
}

// findSpriteByID finds a sprite by its ID or falls back to using the index if ID not found
func findSpriteByID(spriteNumInt int) *SpriteInfo {
	// --- Find the Sprite by ID ---
	var spriteInfo *SpriteInfo
	for i := range currentSprites {
		if currentSprites[i].ID == spriteNumInt { // Use the integer version
			spriteInfo = &currentSprites[i]
			break
		}
	}

	if spriteInfo == nil {
		// If we can't find a sprite with the exact ID, try to use the array index as a fallback
		if spriteNumInt >= 0 && spriteNumInt < len(currentSprites) {
			spriteInfo = &currentSprites[spriteNumInt]
		}
	}

	return spriteInfo
}

// parseSprOptions parses the optional arguments for the Spr function
func parseSprOptions(options []any) (scaleW float64, scaleH float64, flipX bool, flipY bool) {
	// Default values
	scaleW = 1.0
	scaleH = 1.0
	flipX = false
	flipY = false

	// Process optional width multiplier (arg 1)
	if len(options) > 0 && options[0] != nil {
		switch val := options[0].(type) {
		case int:
			scaleW = float64(val)
		case float64:
			scaleW = val
		default:
			log.Printf("Warning: Spr() optional arg 1: expected float64 or int (width multiplier), got %T (%v)", options[0], options[0])
		}
	}

	// Process optional height multiplier (arg 2)
	if len(options) > 1 && options[1] != nil {
		switch val := options[1].(type) {
		case int:
			scaleH = float64(val)
		case float64:
			scaleH = val
		default:
			log.Printf("Warning: Spr() optional arg 2: expected float64 or int (height multiplier), got %T (%v)", options[1], options[1])
		}
	}

	// Process optional flipX (arg 3)
	if len(options) > 2 && options[2] != nil {
		switch val := options[2].(type) {
		case bool:
			flipX = val
		default:
			log.Printf("Warning: Spr() optional arg 3: expected bool (flipX), got %T (%v)", options[2], options[2])
		}
	}

	// Process optional flipY (arg 4)
	if len(options) > 3 && options[3] != nil {
		switch val := options[3].(type) {
		case bool:
			flipY = val
		default:
			log.Printf("Warning: Spr() optional arg 4: expected bool (flipY), got %T (%v)", options[3], options[3])
		}
	}

	// Warn if too many arguments
	if len(options) > 4 {
		log.Printf("Warning: Spr() called with too many arguments (%d), expected max 6 (num, x, y, w, h, fx, fy).", len(options)+3)
	}

	return scaleW, scaleH, flipX, flipY
}

// createTransparentSpriteImage creates a new image from the sprite with transparent pixels applied
func createTransparentSpriteImage(tileImage *ebiten.Image) *ebiten.Image {
	// Create a temporary image with transparency for the transparent color
	tempImage := ebiten.NewImage(tileImage.Bounds().Dx(), tileImage.Bounds().Dy())

	// Copy the sprite image to the temporary image, applying transparency
	for y := 0; y < tileImage.Bounds().Dy(); y++ {
		for x := 0; x < tileImage.Bounds().Dx(); x++ {
			// Get the color at this position
			pixelColor := tileImage.At(x, y)

			// Check if this pixel is transparent based on the palette transparency settings
			isTransparent := false

			// Find which color index this pixel matches
			for i, paletteColor := range Pico8Palette {
				if colorEquals(pixelColor, paletteColor) {
					// Check if this color is set to be transparent
					if i < len(PaletteTransparency) && PaletteTransparency[i] {
						isTransparent = true
					}
					break
				}
			}

			// Only draw non-transparent pixels
			if !isTransparent {
				tempImage.Set(x, y, pixelColor)
			}
		}
	}

	return tempImage
}

// setupDrawOptions creates and configures the drawing options for a sprite
func setupDrawOptions(fx, fy, destWidth, destHeight, scaleW, scaleH float64, flipX, flipY bool) *ebiten.DrawImageOptions {
	// Create drawing options
	opts := &ebiten.DrawImageOptions{}

	// Apply scaling
	if scaleW != 1.0 || scaleH != 1.0 {
		opts.GeoM.Scale(scaleW, scaleH)
	}

	// Apply flipping if needed
	if flipX {
		// For X flip: Scale by -1 on X axis, then translate to compensate
		opts.GeoM.Scale(-1, 1)
		opts.GeoM.Translate(destWidth, 0)
	}

	if flipY {
		// For Y flip: Scale by -1 on Y axis, then translate to compensate
		opts.GeoM.Scale(1, -1)
		opts.GeoM.Translate(0, destHeight)
	}

	// Apply final position
	opts.GeoM.Translate(fx, fy)

	// Ensure nearest-neighbor filtering for pixel-perfect rendering
	opts.Filter = ebiten.FilterNearest

	return opts
}

// GetSpriteImage returns the *ebiten.Image for a given sprite ID.
// It first tries to find a sprite with a matching ID.
// If not found, it tries to use the spriteID as an index into the spritesheet.
// Returns nil if the sprite cannot be found.
func GetSpriteImage(spriteID int) *ebiten.Image {
	allSprites := CurrentSprites() // Get sprites from engine
	if allSprites == nil {
		// This can happen if sprites haven't been loaded yet.
		// Attempt to load them, similar to Spr/Sspr.
		loaded, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: GetSpriteImage failed to load spritesheet: %v", err)
			return nil
		}
		currentSprites = loaded // Store for future calls within this package
		allSprites = currentSprites
		if allSprites == nil { // Still nil after attempt
			log.Println("Warning: GetSpriteImage called when currentSprites is nil and load failed")
			return nil
		}
	}

	var foundSpriteInfo *SpriteInfo

	// Try to find by ID first
	for i := range allSprites {
		if allSprites[i].ID == spriteID {
			foundSpriteInfo = &allSprites[i]
			break
		}
	}

	// If not found by ID, try to use spriteID as an index (fallback)
	if foundSpriteInfo == nil {
		if spriteID >= 0 && spriteID < len(allSprites) {
			// Check if the sprite at this index has been initialized (has an Image)
			if allSprites[spriteID].Image != nil {
				foundSpriteInfo = &allSprites[spriteID]
			}
		}
	}

	if foundSpriteInfo != nil && foundSpriteInfo.Image != nil {
		return foundSpriteInfo.Image
	}

	// Optionally, log if a sprite is truly not found, but be mindful of performance if called often.
	// log.Printf("Debug: GetSpriteImage could not find sprite with ID or index: %d", spriteID)
	return nil
}

// Sget returns the color number (0-15) of a pixel at the specified coordinates on the spritesheet.
// If the coordinates are outside the spritesheet bounds, it returns 0.
//
// x: the distance from the left side of the spritesheet (in pixels).
// y: the distance from the top side of the spritesheet (in pixels).
//
// Example:
//
//	// Get the color of pixel at (10,20) on the spritesheet
//	pixel_color := Sget(10, 20) // Returns color index (0-15) if pixel exists
func Sget[X Number, Y Number](x X, y Y) int {
	// Convert generic x, y to required types
	px := int(x)
	py := int(y)

	// Ensure spritesheet is loaded
	if currentSprites == nil {
		loaded, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Failed to load spritesheet for Sget(): %v", err)
			return 0 // Return 0 if spritesheet couldn't be loaded
		}
		currentSprites = loaded
	}

	// In PICO-8, sprites are arranged in a grid on the spritesheet
	// Each sprite is 8x8 pixels, and the spritesheet is 128x128 pixels (16x16 sprites)
	// Find which sprite contains the specified pixel coordinates
	spriteX := px / 8                                   // Determine which sprite column contains the pixel
	spriteY := py / 8                                   // Determine which sprite row contains the pixel
	spriteCellID := calculateSpriteID(spriteX, spriteY) // Calculate sprite ID based on dynamic dimensions

	// Calculate the pixel position within the sprite
	localX := px % 8 // X position within the sprite (0-7)
	localY := py % 8 // Y position within the sprite (0-7)

	// Find the sprite with the matching ID
	for _, sprite := range currentSprites {
		if sprite.ID == spriteCellID {
			// Get the color at the specified pixel within this sprite
			pixelColor := sprite.Image.At(localX, localY)

			// Find the matching color in the PICO-8 palette
			for i, color := range Pico8Palette {
				if colorEquals(pixelColor, color) {
					return i // Return the color index (0-15)
				}
			}
			// If no matching color found, return 0 (transparent/black)
			return 0
		}
	}

	// If no matching pixel was found, return 0
	return 0
}

// colorEquals compares two colors for equality
func colorEquals(c1, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// Color sets the current draw color to be used by subsequent drawing operations.
// The color parameter should be a number from 0 to 15 corresponding to the PICO-8 palette.
//
// Example:
//
//	Color(8) // Set current draw color to red (color 8)
//	Sset(10, 20) // Draw a red pixel at (10, 20) on the spritesheet
func Color(colorIndex int) {
	// Clamp color index to valid range (0-15)
	if colorIndex < 0 {
		colorIndex = 0
	} else if colorIndex >= len(Pico8Palette) {
		colorIndex = len(Pico8Palette) - 1
	}

	// Update both color variables to keep them in sync
	currentDrawColor = colorIndex
	cursorColor = colorIndex
}

// Sset sets the color of a pixel at the specified coordinates on the spritesheet.
// If the optional color parameter is not provided, it uses the current draw color.
//
// x: the distance from the left side of the spritesheet (in pixels).
// y: the distance from the top side of the spritesheet (in pixels).
// color: (optional) a color number from 0 to 15.
//
// Example:
//
//	Sset(10, 0, 8) // Draw a red pixel at (10,0) on the spritesheet
//	Color(12)
//	Sset(16, 0) // Draw a blue pixel at (16,0) using the current draw color

// Fget returns the flag status of a sprite.
// If flag is provided, returns true if that specific flag is set, false otherwise.
// If flag is not provided, returns the entire bitfield of all flags.
//
// spriteNum: the sprite number to check.
// flag: (optional) the flag number (0-7) to check.
//
// Example:
//
//	// Check if flag 0 is set on sprite 1
//	isSet := Fget(1, 0) // Returns true or false
//
//	// Get all flags for sprite 2 as a bitfield
//	allFlags := Fget(2) // Returns an integer (0-255)

// Fget returns the flag status of a sprite.
// Returns:
// - bitfield: the entire bitfield of all flags (0-255)
// - isSet: true if the specific flag is set (only meaningful when a flag is provided)
//
// When no flag is specified, only check the bitfield value and ignore isSet.
// When a flag is specified, check isSet for that specific flag's status.

// Fset sets the flag status of a sprite.
// If flag is provided, sets that specific flag to the value.
// If flag is not provided, sets all flags according to the value (either a boolean or a bitfield).
//
// spriteNum: the sprite number to modify.
// flag: (optional) the flag number (0-7) to set.
// value: true/false to turn the flag on/off, or a bitfield (0-255) to set multiple flags at once.
//
// Example:
//
//	// Set flag 0 to true on sprite 1
//	Fset(1, 0, true)
//
//	// Set all flags off on sprite 2
//	Fset(2, false)
//
//	// Set flags 1,3,5,7 on sprite 2 using a bitfield (170 = 2+8+32+128)
//	Fset(2, 170)

// Fset sets the flag status of a sprite.
// If flag is provided, sets that specific flag to the value.
// If flag is not provided, sets all flags according to the value (either a boolean or a bitfield).
//
// spriteNum: the sprite number to modify.
// flagOrValue: either the flag number (0-7) or a boolean/bitfield value.
// value: (optional) true/false to turn the flag on/off.
func Fset(spriteNum int, flagOrValue interface{}, value ...interface{}) {
	// Lazy-load sprites if needed
	if currentSprites == nil {
		sprites, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Fset() called but failed to load spritesheet: %v", err)
			return
		}
		currentSprites = sprites
	}

	// Find the sprite with the matching ID
	var spriteIndex = -1
	for i := range currentSprites {
		if currentSprites[i].ID == spriteNum {
			spriteIndex = i
			break
		}
	}

	// If sprite not found, return
	if spriteIndex == -1 {
		log.Printf("Warning: Fset() called with invalid sprite number: %d", spriteNum)
		return
	}

	// Case 1: Setting a specific flag
	if flagNum, ok := flagOrValue.(int); ok && len(value) > 0 {
		// Validate flag index
		if flagNum < 0 || flagNum >= 8 {
			log.Printf("Warning: Fset() called with invalid flag number: %d", flagNum)
			return
		}

		// Get the boolean value
		var boolValue bool
		switch v := value[0].(type) {
		case bool:
			boolValue = v
		case int:
			boolValue = v != 0
		default:
			log.Printf("Warning: Fset() called with invalid value type: %T", value[0])
			return
		}

		// Set the flag
		currentSprites[spriteIndex].Flags.Individual[flagNum] = boolValue

		// Update the bitfield
		if boolValue {
			// Set the bit
			currentSprites[spriteIndex].Flags.Bitfield |= 1 << flagNum
		} else {
			// Clear the bit
			currentSprites[spriteIndex].Flags.Bitfield &= ^(1 << flagNum)
		}
		return
	}

	// Case 2: Setting all flags with a boolean
	if boolValue, ok := flagOrValue.(bool); ok {
		// Set all flags to the same value
		for i := 0; i < 8; i++ {
			currentSprites[spriteIndex].Flags.Individual[i] = boolValue
		}

		// Update the bitfield
		if boolValue {
			currentSprites[spriteIndex].Flags.Bitfield = 255 // All bits set
		} else {
			currentSprites[spriteIndex].Flags.Bitfield = 0 // All bits cleared
		}
		return
	}

	// Case 3: Setting flags with a bitfield
	if intValue, ok := flagOrValue.(int); ok {
		// Clamp the value to valid range (0-255)
		if intValue < 0 {
			intValue = 0
		} else if intValue > 255 {
			intValue = 255
		}

		// Set the bitfield
		currentSprites[spriteIndex].Flags.Bitfield = intValue

		// Update individual flags
		for i := 0; i < 8; i++ {
			currentSprites[spriteIndex].Flags.Individual[i] = (intValue & (1 << i)) != 0
		}
		return
	}

	log.Printf("Warning: Fset() called with invalid arguments: %v, %v", flagOrValue, value)
}

// Fget returns the flag status of a sprite.
// Returns:
// - bitfield: the entire bitfield of all flags (0-255)
// - isSet: true if the specific flag is set (only meaningful when a flag is provided)
//
// spriteNum: the sprite number to check.
// flag: (optional) the flag number (0-7) to check.
//
// When no flag is specified, only check the bitfield value and ignore isSet.
// When a flag is specified, check isSet for that specific flag's status.
func Fget(spriteNum int, flag ...int) (bitfield int, isSet bool) {
	// Lazy-load sprites if needed
	if currentSprites == nil {
		sprites, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Fget() called but failed to load spritesheet: %v", err)
			return 0, false
		}
		currentSprites = sprites
	}

	// Find the sprite with the matching ID
	var spriteInfo *SpriteInfo
	for i := range currentSprites {
		if currentSprites[i].ID == spriteNum {
			spriteInfo = &currentSprites[i]
			break
		}
	}

	// If sprite not found, return zero values
	if spriteInfo == nil {
		log.Printf("Warning: Fget() called for non-existent sprite ID %d", spriteNum)
		return 0, false
	}

	// Get the entire bitfield
	bitfield = spriteInfo.Flags.Bitfield

	// If a specific flag is requested, check that flag
	if len(flag) > 0 {
		flagNum := flag[0]

		// Validate flag number (0-7)
		if flagNum < 0 || flagNum > 7 {
			log.Printf("Warning: Fget() called with invalid flag number %d. Valid range is 0-7.", flagNum)
			return bitfield, false
		}

		// Check if the specific flag is set
		bitMask := 1 << flagNum
		isSet = (bitfield & bitMask) != 0
	}

	return bitfield, isSet
}

// Sset sets the color of a pixel at the specified coordinates on the spritesheet.
// If the optional color parameter is not provided, it uses the current draw color.
//
// x: the distance from the left side of the spritesheet (in pixels).
// y: the distance from the top side of the spritesheet (in pixels).
// colorIndex: (optional) a color number from 0 to 15.
func Sset[X Number, Y Number](x X, y Y, colorIndex ...int) {
	// Convert generic x, y to required types
	px := int(x)
	py := int(y)

	// Determine which color to use
	colorToUse := currentDrawColor
	if len(colorIndex) > 0 {
		colorToUse = colorIndex[0]
		// Clamp color index to valid range (0-15)
		if colorToUse < 0 {
			colorToUse = 0
		} else if colorToUse >= len(Pico8Palette) {
			colorToUse = len(Pico8Palette) - 1
		}
	}

	// Ensure spritesheet is loaded
	if currentSprites == nil {
		loaded, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Failed to load spritesheet for Sset(): %v", err)
			return // Can't set pixel if spritesheet couldn't be loaded
		}
		currentSprites = loaded
	}

	// In PICO-8, sprites are arranged in a grid on the spritesheet
	// Each sprite is 8x8 pixels, and the spritesheet is 128x128 pixels (16x16 sprites)
	// Find which sprite contains the specified pixel coordinates
	spriteX := px / 8                                   // Determine which sprite column contains the pixel
	spriteY := py / 8                                   // Determine which sprite row contains the pixel
	spriteCellID := calculateSpriteID(spriteX, spriteY) // Calculate sprite ID based on dynamic dimensions

	// Calculate the pixel position within the sprite
	localX := px % 8 // X position within the sprite (0-7)
	localY := py % 8 // Y position within the sprite (0-7)

	// Find the sprite with the matching ID
	for i := range currentSprites {
		sprite := &currentSprites[i]
		if sprite.ID == spriteCellID {
			// Set the pixel color within this sprite
			sprite.Image.Set(localX, localY, Pico8Palette[colorToUse])
			return
		}
	}

	// If no sprite with the matching ID was found, log a warning
	log.Printf("Warning: Sset() called for non-existent sprite ID %d at position (%d, %d)", spriteCellID, px, py)
}

// Sspr draws a sprite from the spritesheet with custom dimensions and optional stretching and flipping.
// Mimics PICO-8's sspr(sx, sy, sw, sh, dx, dy, [dw, dh], [flip_x], [flip_y]) function.
//
// sx: sprite sheet x position (in pixels)
// sy: sprite sheet y position (in pixels)
// sw: sprite width (in pixels)
// sh: sprite height (in pixels)
// dx: how far from the left of the screen to draw the sprite
// dy: how far from the top of the screen to draw the sprite
// dw: (optional) how many pixels wide to draw the sprite (default same as sw)
// dh: (optional) how many pixels tall to draw the sprite (default same as sh)
// flip_x: (optional) boolean, if true draw the sprite flipped horizontally (default false)
// flip_y: (optional) boolean, if true draw the sprite flipped vertically (default false)
//
// Example:
//
//	// Draw a 16x16 sprite from position (8,8) on the spritesheet to position (10,20) on the screen
//	Sspr(8, 8, 16, 16, 10, 20)
//
//	// Draw a 6x5 sprite from position (8,8) on the spritesheet to position (10,20) on the screen
//	Sspr(8, 8, 6, 5, 10, 20)
//
//	// Draw a 16x16 sprite from the spritesheet, stretched to 32x32 on the screen
//	Sspr(8, 8, 16, 16, 10, 20, 32, 32)
//
//	// Draw a 16x16 sprite, flipped horizontally
//	Sspr(8, 8, 16, 16, 10, 20, 16, 16, true, false)
//
// parseSsprOptions parses the optional arguments for the Sspr function
func parseSsprOptions(options []any, sourceWidth, sourceHeight int) (destWidth, destHeight float64, flipX, flipY bool) {
	// Default values
	destWidth = float64(sourceWidth)
	destHeight = float64(sourceHeight)
	flipX = false
	flipY = false

	// Helper function for logging argument errors
	argError := func(pos int, expected string, val interface{}) {
		log.Printf("Warning: Sspr() optional arg %d: expected %s, got %T (%v)", pos+1, expected, val, val)
	}

	// Process optional dw parameter
	if len(options) >= 1 && options[0] != nil {
		dwVal, ok := options[0].(float64)
		if !ok {
			if dwInt, intOk := options[0].(int); intOk {
				dwVal = float64(dwInt)
				ok = true
			}
		}
		if !ok {
			argError(0, "float64 or int (destination width)", options[0])
		} else {
			destWidth = dwVal
		}
	}

	// Process optional dh parameter
	if len(options) >= 2 && options[1] != nil {
		dhVal, ok := options[1].(float64)
		if !ok {
			if dhInt, intOk := options[1].(int); intOk {
				dhVal = float64(dhInt)
				ok = true
			}
		}
		if !ok {
			argError(1, "float64 or int (destination height)", options[1])
		} else {
			destHeight = dhVal
		}
	}

	// Process optional flip_x parameter
	if len(options) >= 3 && options[2] != nil {
		flipXVal, ok := options[2].(bool)
		if !ok {
			argError(2, "bool (flip_x)", options[2])
		} else {
			flipX = flipXVal
		}
	}

	// Process optional flip_y parameter
	if len(options) >= 4 && options[3] != nil {
		flipYVal, ok := options[3].(bool)
		if !ok {
			argError(3, "bool (flip_y)", options[3])
		} else {
			flipY = flipYVal
		}
	}

	if len(options) > 4 {
		log.Printf("Warning: Sspr() called with too many arguments (%d), expected max 10 (sx, sy, sw, sh, dx, dy, dw, dh, flip_x, flip_y).", len(options)+6)
	}

	return destWidth, destHeight, flipX, flipY
}

// createSpriteSourceImage creates a temporary image from the specified region of the spritesheet
func createSpriteSourceImage(sourceX, sourceY, sourceWidth, sourceHeight int) *ebiten.Image {
	// Create a temporary image for the source region with transparency
	sourceImage := ebiten.NewImage(sourceWidth, sourceHeight)

	// Clear the image with transparent color
	sourceImage.Fill(color.RGBA{0, 0, 0, 0})

	// Copy the specified region from the spritesheet to the temporary image
	for y := 0; y < sourceHeight; y++ {
		for x := 0; x < sourceWidth; x++ {
			// Get the color at this position on the spritesheet
			colorIndex := Sget(sourceX+x, sourceY+y)

			// Skip transparent pixels based on the palette transparency settings
			if colorIndex >= 0 && colorIndex < len(PaletteTransparency) && PaletteTransparency[colorIndex] {
				// Skip this pixel, leaving it transparent
				continue
			}

			if colorIndex >= 0 && colorIndex < len(Pico8Palette) {
				// Set the pixel in the temporary image
				sourceImage.Set(x, y, Pico8Palette[colorIndex])
			}
		}
	}

	return sourceImage
}

// Sspr draws a sprite from the spritesheet with custom dimensions and optional stretching and flipping.
// Mimics PICO-8's sspr(sx, sy, sw, sh, dx, dy, [dw, dh], [flip_x], [flip_y]) function.
//
// sx: sprite sheet x position (in pixels)
// sy: sprite sheet y position (in pixels)
// sw: sprite width (in pixels)
// sh: sprite height (in pixels)
// dx: how far from the left of the screen to draw the sprite
// dy: how far from the top of the screen to draw the sprite
// dw: (optional) how many pixels wide to draw the sprite (default same as sw)
// dh: (optional) how many pixels tall to draw the sprite (default same as sh)
// flip_x: (optional) boolean, if true draw the sprite flipped horizontally (default false)
// flip_y: (optional) boolean, if true draw the sprite flipped vertically (default false)
//
// Example:
//
//	// Draw a 16x16 sprite from position (8,8) on the spritesheet to position (10,20) on the screen
//	Sspr(8, 8, 16, 16, 10, 20)
//
//	// Draw a 6x5 sprite from position (8,8) on the spritesheet to position (10,20) on the screen
//	Sspr(8, 8, 6, 5, 10, 20)
//
//	// Draw a 16x16 sprite from the spritesheet, stretched to 32x32 on the screen
//	Sspr(8, 8, 16, 16, 10, 20, 32, 32)
//
//	// Draw a 16x16 sprite, flipped horizontally
//	Sspr(8, 8, 16, 16, 10, 20, 16, 16, true, false)
func Sspr[SX Number, SY Number, SW Number, SH Number, DX Number, DY Number](sx SX, sy SY, sw SW, sh SH, dx DX, dy DY, options ...any) {
	// Convert generic types to required types
	sourceX := int(sx)      // Source X on spritesheet
	sourceY := int(sy)      // Source Y on spritesheet
	sourceWidth := int(sw)  // Source width on spritesheet
	sourceHeight := int(sh) // Source height on spritesheet
	destX := float64(dx)    // Destination X on screen
	destY := float64(dy)    // Destination Y on screen

	// Use internal package variables set by engine.Draw
	if currentScreen == nil {
		log.Println("Warning: Sspr() called before screen was ready.")
		return
	}

	// --- Lazy Loading Logic ---
	if currentSprites == nil {
		loaded, err := loadSpritesheet()
		if err != nil {
			log.Printf("Warning: Failed to load spritesheet for Sspr(): %v", err)
			return
		}
		currentSprites = loaded
	}

	// Parse optional arguments
	destWidth, destHeight, flipX, flipY := parseSsprOptions(options, sourceWidth, sourceHeight)

	// Validate source rectangle is within spritesheet bounds
	if !validateSpriteSheetBounds(sourceX, sourceY, sourceWidth, sourceHeight) {
		log.Printf("Warning: Sspr() source rectangle (%d,%d,%d,%d) is outside spritesheet bounds (0,0,%d,%d)",
			sourceX, sourceY, sourceWidth, sourceHeight, SpritesheetWidth, SpritesheetHeight)
		// Continue anyway, Ebiten will handle clipping
	}

	// Clamp dimensions to be non-negative
	destWidth = math.Max(0, destWidth)
	destHeight = math.Max(0, destHeight)
	if destWidth == 0 || destHeight == 0 {
		return // Don't draw if scaled to zero size
	}

	// Create a temporary image for the source region
	sourceImage := createSpriteSourceImage(sourceX, sourceY, sourceWidth, sourceHeight)

	// Set up drawing options
	op := &ebiten.DrawImageOptions{}

	// Apply camera offset to the intended top-left drawing position (dx, dy)
	screenDrawX, screenDrawY := applyCameraOffset(destX, destY)

	// Apply scaling to match the destination dimensions
	scaleX := destWidth / float64(sourceWidth)
	scaleY := destHeight / float64(sourceHeight)

	// Temporary variables for final translation, considering flips
	finalTranslateX := screenDrawX
	finalTranslateY := screenDrawY

	// Apply flip transformations if needed
	if flipX {
		scaleX *= -1.0
		finalTranslateX += destWidth // Adjust translation for horizontal flip
	}
	if flipY {
		scaleY *= -1.0
		finalTranslateY += destHeight // Adjust translation for vertical flip
	}

	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(finalTranslateX, finalTranslateY) // Use camera-adjusted and flip-adjusted coordinates

	// Draw the image to the screen
	currentScreen.DrawImage(sourceImage, op)
}
