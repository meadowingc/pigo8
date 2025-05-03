package pigo8

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Note: The global currentDrawColor is defined in engine.go and set by the Color() function

// parseRectArgs parses common arguments for Rect and Rectfill.
// It returns the calculated top-left corner (x, y), dimensions (w, h),
// the PICO-8 color index to use, and whether parsing was successful.
func parseRectArgs(x1, y1, x2, y2 float64, options []interface{}) (float32, float32, float32, float32, int, bool) {
	// Determine drawing color
	drawColorIndex := currentDrawColor // Use the global current draw color set by Color()
	if len(options) >= 1 {
		if cIdx, ok := options[0].(int); ok {
			if cIdx >= 0 && cIdx < len(Pico8Palette) {
				drawColorIndex = cIdx
				// Update the global drawing color to match PICO-8 behavior
				currentDrawColor = cIdx
			} else {
				log.Printf("Warning: Rect/Rectfill optional color %d out of range (0-15). Using current color %d.", cIdx, drawColorIndex)
			}
		} else {
			log.Printf("Warning: Rect/Rectfill optional color argument expected int, got %T. Using current color %d.", options[0], drawColorIndex)
		}
	}
	if len(options) > 1 {
		log.Printf("Warning: Rect/Rectfill called with too many arguments (%d), expected max 5.", len(options)+4)
	}

	// Calculate top-left corner (x, y) and dimensions (width, height)
	// PICO-8 rect/rectfill is inclusive, so add 1 to the difference.
	rectX := math.Min(x1, x2)
	rectY := math.Min(y1, y2)
	rectW := math.Abs(x2-x1) + 1
	rectH := math.Abs(y2-y1) + 1

	return float32(rectX), float32(rectY), float32(rectW), float32(rectH), drawColorIndex, true
}

// Rect draws an outline rectangle using two corner points.
// Mimics PICO-8's rect(x1, y1, x2, y2, [color]) function.
//
// x1, y1, x2, y2: Coordinates of two opposing corners (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Rect[X1 Number, Y1 Number, X2 Number, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Rect() called before screen was ready.")
		return
	}

	fx1, fy1, fx2, fy2 := float64(x1), float64(y1), float64(x2), float64(y2)
	rectX, rectY, rectW, rectH, drawColorIndex, ok := parseRectArgs(fx1, fy1, fx2, fy2, options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(Pico8Palette) {
		actualColor = Pico8Palette[drawColorIndex]
	} else {
		actualColor = Pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Rect(). Defaulting to black.", drawColorIndex)
	}

	// Draw outline by drawing four 1-pixel thick filled rectangles
	topY := rectY
	bottomY := rectY + rectH - 1
	leftX := rectX
	rightX := rectX + rectW - 1

	// Top horizontal line
	vector.DrawFilledRect(currentScreen, leftX, topY, rectW, 1, actualColor, false)
	// Bottom horizontal line
	vector.DrawFilledRect(currentScreen, leftX, bottomY, rectW, 1, actualColor, false)
	// Left vertical line (height adjusted to avoid drawing corners twice)
	vector.DrawFilledRect(currentScreen, leftX, topY+1, 1, rectH-2, actualColor, false)
	// Right vertical line (height adjusted to avoid drawing corners twice)
	vector.DrawFilledRect(currentScreen, rightX, topY+1, 1, rectH-2, actualColor, false)

	/* // Original StrokeRect implementation - might clip at edges
	strokeWidth := float32(1.0) // PICO-8 rect outline is 1 pixel thick
	vector.StrokeRect(
		currentScreen,
		rectX,
		rectY,
		rectW,
		rectH,
		strokeWidth,
		actualColor,
		false, // No anti-aliasing
	)
	*/
}

// Rectfill draws a filled rectangle using two corner points.
// Mimics PICO-8's rectfill(x1, y1, x2, y2, [color]) function.
//
// x1, y1, x2, y2: Coordinates of two opposing corners (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Rectfill[X1 Number, Y1 Number, X2 Number, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Rectfill() called before screen was ready.")
		return
	}

	fx1, fy1, fx2, fy2 := float64(x1), float64(y1), float64(x2), float64(y2)
	rectX, rectY, rectW, rectH, drawColorIndex, ok := parseRectArgs(fx1, fy1, fx2, fy2, options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(Pico8Palette) {
		actualColor = Pico8Palette[drawColorIndex]
	} else {
		actualColor = Pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Rectfill(). Defaulting to black.", drawColorIndex)
	}

	// Draw filled rectangle using Ebitengine vector graphics
	vector.DrawFilledRect(
		currentScreen,
		rectX,
		rectY,
		rectW,
		rectH,
		actualColor,
		false, // No anti-aliasing
	)
}
