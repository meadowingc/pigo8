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

	// --- Find the Sprite by ID ---
	var spriteInfo *SpriteInfo
	for i := range currentSprites {
		if currentSprites[i].ID == spriteNumInt { // Use the integer version
			spriteInfo = &currentSprites[i]
			break
		}
	}

	if spriteInfo == nil || spriteInfo.Image == nil {
		// log.Printf("Warning: Spr() called for non-existent or unloaded sprite ID %d.", spriteNumInt) // Use the integer version
		// Don't log by default, PICO-8 doesn't warn for drawing non-existent sprites
		return // Sprite ID not found or image wasn't loaded
	}

	// Default values for optional arguments
	wMultiplier := 1.0
	hMultiplier := 1.0
	flipX := false
	flipY := false

	// --- Argument Processing ---
	argError := func(pos int, expected string, val interface{}) {
		log.Printf("Warning: Spr() optional arg %d: expected %s, got %T (%v)", pos+1, expected, val, val)
	}

	if len(options) >= 1 {
		wVal, ok := options[0].(float64)
		if !ok {
			if wInt, intOk := options[0].(int); intOk {
				wVal = float64(wInt)
				ok = true
			}
		}
		if !ok {
			argError(0, "float64 or int (width multiplier)", options[0])
			// Don't return on error, just use default
		} else {
			wMultiplier = wVal
		}
	}
	if len(options) >= 2 {
		hVal, ok := options[1].(float64)
		if !ok {
			if hInt, intOk := options[1].(int); intOk {
				hVal = float64(hInt)
				ok = true
			}
		}
		if !ok {
			argError(1, "float64 or int (height multiplier)", options[1])
			// Don't return on error, just use default
		} else {
			hMultiplier = hVal
		}
	}
	if len(options) >= 3 {
		flipXVal, ok := options[2].(bool)
		if !ok {
			argError(2, "bool (flipX)", options[2])
		} else {
			flipX = flipXVal
		}
	}
	if len(options) >= 4 {
		flipYVal, ok := options[3].(bool)
		if !ok {
			argError(3, "bool (flipY)", options[3])
		} else {
			flipY = flipYVal
		}
	}
	if len(options) > 4 {
		log.Printf("Warning: Spr() called with too many arguments (%d), expected max 6 (num, x, y, w, h, fx, fy).", len(options)+3)
	}

	// Clamp multipliers to be non-negative
	wMultiplier = math.Max(0, wMultiplier)
	hMultiplier = math.Max(0, hMultiplier)
	if wMultiplier == 0 || hMultiplier == 0 {
		return // Don't draw if scaled to zero size
	}

	// --- Drawing Logic ---
	tileImage := spriteInfo.Image
	spriteWidth := float64(tileImage.Bounds().Dx())
	spriteHeight := float64(tileImage.Bounds().Dy())

	op := &ebiten.DrawImageOptions{}

	// Apply scaling
	// Note: PICO-8's w/h arguments are multipliers for the base 8x8 sprite size.
	// Our sprites might technically not be 8x8 if the JSON differs, but we assume they are loaded as such.
	// The scaling factor multiplies the sprite's *actual* loaded dimensions.
	scaleX := wMultiplier
	scaleY := hMultiplier

	// Centre point for flipping (relative to the sprite's top-left corner)
	centerX := spriteWidth / 2.0
	centerY := spriteHeight / 2.0

	// Apply flip transformations by scaling around the center
	if flipX {
		scaleX *= -1.0
	}
	if flipY {
		scaleY *= -1.0
	}

	// Translate to center, scale (applies scaling and flip), translate back
	op.GeoM.Translate(-centerX, -centerY)
	op.GeoM.Scale(scaleX, scaleY)
	op.GeoM.Translate(centerX, centerY)

	// Translate to final position on screen
	op.GeoM.Translate(fx, fy)

	// Draw the image
	currentScreen.DrawImage(tileImage, op)
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
	spriteX := px / 8 // Determine which sprite column contains the pixel
	spriteY := py / 8 // Determine which sprite row contains the pixel
	spriteCellID := spriteY*16 + spriteX // Calculate sprite ID based on position (16 sprites per row)

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
	
	currentDrawColor = colorIndex
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
	spriteX := px / 8 // Determine which sprite column contains the pixel
	spriteY := py / 8 // Determine which sprite row contains the pixel
	spriteCellID := spriteY*16 + spriteX // Calculate sprite ID based on position (16 sprites per row)
	
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
