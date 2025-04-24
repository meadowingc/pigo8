package pigo8

import (
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
//	// Explicitly specify generic types if needed (rarely necessary)
//	Spr[int, float64](1, 10, 20.5)
func Spr[X Number, Y Number](spriteNumber int, x X, y Y, options ...interface{}) {
	// Convert generic x, y to float64 for internal calculations
	fx := float64(x)
	fy := float64(y)

	// Use internal package variables set by engine.Draw
	if currentScreen == nil {
		log.Println("Warning: Spr() called before screen was ready.")
		return
	}

	// --- Lazy Loading Logic ---
	if currentSprites == nil {
		log.Println("Spr(): Attempting to load spritesheet...")
		loaded, err := loadSpritesheet() // Call the loading function from spritesheet.go
		if err != nil {
			log.Fatalf("Fatal: Failed to load required spritesheet for Spr(): %v", err)
		}
		log.Printf("Spr(): Successfully loaded %d sprites.", len(loaded))
		currentSprites = loaded // Store successfully loaded sprites

	}

	// --- Find the Sprite by ID ---
	var spriteInfo *SpriteInfo
	for i := range currentSprites {
		if currentSprites[i].ID == spriteNumber {
			spriteInfo = &currentSprites[i]
			break
		}
	}

	if spriteInfo == nil || spriteInfo.Image == nil {
		// log.Printf("Warning: Spr() called for non-existent or unloaded sprite ID %d.", spriteNumber)
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
