package pigo8

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Note: The global currentDrawColor is defined in engine.go and set by the Color() function

// warningsSeen tracks which warning messages have already been shown
var warningsSeen = make(map[string]bool)

// logWarningOnce logs a warning message only once, even if called multiple times with the same message
func logWarningOnce(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	if !warningsSeen[msg] {
		log.Print(msg)
		warningsSeen[msg] = true
	}
}

// parseRectArgs parses common arguments for Rect and Rectfill.
// It returns the calculated top-left corner (x, y), dimensions (w, h),
// the PICO-8 color index to use, and whether parsing was successful.
func parseRectArgs(x1, y1, x2, y2 float64, options []interface{}) (float32, float32, float32, float32, int, bool) {
	// Determine drawing color
	drawColorIndex := currentDrawColor // Use the global current draw color set by Color()
	if len(options) >= 1 {
		// Try to handle different numeric types for color
		switch v := options[0].(type) {
		case int:
			// Handle integer color directly
			if v >= 0 && v < len(Pico8Palette) {
				drawColorIndex = v
				// Update both color variables to keep them in sync
				currentDrawColor = v
				cursorColor = v
			} else {
				logWarningOnce("Warning: Rect/Rectfill optional color %d out of range (0-15). Using current color %d.", v, drawColorIndex)
			}
		case float64:
			// Convert float64 to int for color
			intVal := int(v)
			if intVal >= 0 && intVal < len(Pico8Palette) {
				drawColorIndex = intVal
				// Update both color variables to keep them in sync
				currentDrawColor = intVal
				cursorColor = intVal
			} else {
				logWarningOnce("Warning: Rect/Rectfill optional color %d out of range (0-15). Using current color %d.", intVal, drawColorIndex)
			}
		case float32:
			// Convert float32 to int for color
			intVal := int(v)
			if intVal >= 0 && intVal < len(Pico8Palette) {
				drawColorIndex = intVal
				// Update both color variables to keep them in sync
				currentDrawColor = intVal
				cursorColor = intVal
			} else {
				logWarningOnce("Warning: Rect/Rectfill optional color %d out of range (0-15). Using current color %d.", intVal, drawColorIndex)
			}
		default:
			logWarningOnce("Warning: Rect/Rectfill optional color argument expected numeric type, got %T. Using current color %d.", options[0], drawColorIndex)
		}
	}
	if len(options) > 1 {
		logWarningOnce("Warning: Rect/Rectfill called with too many arguments (%d), expected max 5.", len(options)+4)
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

	// Apply camera offset
	fx1, fy1 = applyCameraOffset(fx1, fy1)
	fx2, fy2 = applyCameraOffset(fx2, fy2)

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

	// Apply camera offset
	fx1, fy1 = applyCameraOffset(fx1, fy1)
	fx2, fy2 = applyCameraOffset(fx2, fy2)

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

// parseLineArgs parses common arguments for Line function.
// It returns the PICO-8 color index to use and whether parsing was successful.
func parseLineArgs(options []interface{}) (int, bool) {
	// Determine drawing color
	drawColorIndex := currentDrawColor // Use the global current draw color set by Color()
	if len(options) >= 1 {
		// Try to handle different numeric types for color
		switch v := options[0].(type) {
		case int:
			// Handle integer color directly
			if v >= 0 && v < len(Pico8Palette) {
				drawColorIndex = v
				// Update the global drawing color to match PICO-8 behavior
				currentDrawColor = v
			} else {
				logWarningOnce("Warning: Line optional color %d out of range (0-15). Using current color %d.", v, drawColorIndex)
			}
		case float64:
			// Convert float64 to int for color
			intVal := int(v)
			if intVal >= 0 && intVal < len(Pico8Palette) {
				drawColorIndex = intVal
				// Update the global drawing color to match PICO-8 behavior
				currentDrawColor = intVal
			} else {
				logWarningOnce("Warning: Line optional color %d out of range (0-15). Using current color %d.", intVal, drawColorIndex)
			}
		case float32:
			// Convert float32 to int for color
			intVal := int(v)
			if intVal >= 0 && intVal < len(Pico8Palette) {
				drawColorIndex = intVal
				// Update the global drawing color to match PICO-8 behavior
				currentDrawColor = intVal
			} else {
				logWarningOnce("Warning: Line optional color %d out of range (0-15). Using current color %d.", intVal, drawColorIndex)
			}
		default:
			logWarningOnce("Warning: Line optional color argument expected numeric type, got %T. Using current color %d.", options[0], drawColorIndex)
		}
	}
	if len(options) > 1 {
		logWarningOnce("Warning: Line called with too many arguments (%d), expected max 5.", len(options)+4)
	}

	return drawColorIndex, true
}

// Line draws a line between two points.
// Mimics PICO-8's line(x1, y1, x2, y2, [color]) function.
//
// x1, y1: Coordinates of the starting point (any Number type).
// x2, y2: Coordinates of the ending point (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Line[X1 Number, Y1 Number, X2 Number, Y2 Number](x1 X1, y1 Y1, x2 X2, y2 Y2, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Line() called before screen was ready.")
		return
	}

	// Convert to float64 for calculations
	fx1, fy1, fx2, fy2 := float64(x1), float64(y1), float64(x2), float64(y2)

	// Parse optional color argument
	drawColorIndex, ok := parseLineArgs(options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(Pico8Palette) {
		actualColor = Pico8Palette[drawColorIndex]
	} else {
		actualColor = Pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Line(). Defaulting to black.", drawColorIndex)
	}

	// Draw the line using Ebitengine's vector package
	vector.StrokeLine(
		currentScreen,
		float32(fx1),
		float32(fy1),
		float32(fx2),
		float32(fy2),
		1.0, // Line width of 1 pixel to match PICO-8
		actualColor,
		false, // No anti-aliasing to match PICO-8's pixel-perfect style
	)
}

// parseCircArgs parses common arguments for Circ and Circfill.
// It returns the center coordinates (x, y), radius, the PICO-8 color index to use,
// and whether parsing was successful.
func parseCircArgs(x, y, radius float64, options []interface{}) (float32, float32, float32, int, bool) {
	// Determine drawing color
	drawColorIndex := currentDrawColor // Use the global current draw color set by Color()
	if len(options) >= 1 {
		// Try to handle different numeric types for color
		switch v := options[0].(type) {
		case int:
			// Handle integer color directly
			if v >= 0 && v < len(Pico8Palette) {
				drawColorIndex = v
				// Update both color variables to keep them in sync
				currentDrawColor = v
				cursorColor = v
			} else {
				logWarningOnce("Warning: Circ/Circfill optional color %d out of range (0-15). Using current color %d.", v, drawColorIndex)
			}
		case float64:
			// Convert float64 to int for color
			intVal := int(v)
			if intVal >= 0 && intVal < len(Pico8Palette) {
				drawColorIndex = intVal
				// Update both color variables to keep them in sync
				currentDrawColor = intVal
				cursorColor = intVal
			} else {
				logWarningOnce("Warning: Circ/Circfill optional color %d out of range (0-15). Using current color %d.", intVal, drawColorIndex)
			}
		case float32:
			// Convert float32 to int for color
			intVal := int(v)
			if intVal >= 0 && intVal < len(Pico8Palette) {
				drawColorIndex = intVal
				// Update both color variables to keep them in sync
				currentDrawColor = intVal
				cursorColor = intVal
			} else {
				logWarningOnce("Warning: Circ/Circfill optional color %d out of range (0-15). Using current color %d.", intVal, drawColorIndex)
			}
		default:
			logWarningOnce("Warning: Circ/Circfill optional color argument expected numeric type, got %T. Using current color %d.", options[0], drawColorIndex)
		}
	}
	if len(options) > 1 {
		logWarningOnce("Warning: Circ/Circfill called with too many arguments (%d), expected max 4.", len(options)+3)
	}

	return float32(x), float32(y), float32(radius), drawColorIndex, true
}

// Circ draws an outline circle.
// Mimics PICO-8's circ(x, y, radius, [color]) function.
//
// x, y: Coordinates of the center point (any Number type).
// radius: Radius of the circle (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Circ[X Number, Y Number, R Number](x X, y Y, radius R, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Circ() called before screen was ready.")
		return
	}

	fx, fy, fr := float64(x), float64(y), float64(radius)

	// Apply camera offset
	fx, fy = applyCameraOffset(fx, fy)

	circX, circY, circR, drawColorIndex, ok := parseCircArgs(fx, fy, fr, options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(Pico8Palette) {
		actualColor = Pico8Palette[drawColorIndex]
	} else {
		actualColor = Pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Circ(). Defaulting to black.", drawColorIndex)
	}

	// Draw the circle outline using Ebitengine vector graphics
	vector.StrokeCircle(
		currentScreen,
		circX,
		circY,
		circR,
		1.0, // Stroke width of 1 pixel to match PICO-8's style
		actualColor,
		false, // No anti-aliasing to match PICO-8's pixel-perfect style
	)
}

// Circfill draws a filled circle.
// Mimics PICO-8's circfill(x, y, radius, [color]) function.
//
// x, y: Coordinates of the center point (any Number type).
// radius: Radius of the circle (any Number type).
// options...:
//   - color (int): Optional PICO-8 color index (0-15). If omitted or invalid,
//     uses the current drawing color (defaults to 7 - white currently).
func Circfill[X Number, Y Number, R Number](x X, y Y, radius R, options ...interface{}) {
	if currentScreen == nil {
		log.Println("Warning: Circfill() called before screen was ready.")
		return
	}

	fx, fy, fr := float64(x), float64(y), float64(radius)

	// Apply camera offset
	fx, fy = applyCameraOffset(fx, fy)

	circX, circY, circR, drawColorIndex, ok := parseCircArgs(fx, fy, fr, options)
	if !ok {
		return // Argument parsing logged an issue
	}

	// Get the actual color from the palette
	var actualColor color.Color
	if drawColorIndex >= 0 && drawColorIndex < len(Pico8Palette) {
		actualColor = Pico8Palette[drawColorIndex]
	} else {
		actualColor = Pico8Palette[0] // Fallback to black
		log.Printf("Error: Invalid effective drawing color index %d for Circfill(). Defaulting to black.", drawColorIndex)
	}

	// Draw the filled circle using Ebitengine vector graphics
	vector.DrawFilledCircle(
		currentScreen,
		circX,
		circY,
		circR,
		actualColor,
		false, // No anti-aliasing to match PICO-8's pixel-perfect style
	)
}
