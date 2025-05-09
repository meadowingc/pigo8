package pigo8

import (
	"bytes"
	_ "embed"

	// "fmt" // Not needed for this version

	"image/color"
	_ "image/png" // Keep in case other PNGs are loaded
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2/text/v2" // Re-add text/v2
)

//go:embed pico-8.ttf
var pico8FontTTF []byte

var (
	// Pico8Palette defines the standard 16 PICO-8 colors.
	Pico8Palette = []color.Color{
		color.RGBA{R: 0, G: 0, B: 0, A: 255},       // 0 black
		color.RGBA{R: 29, G: 43, B: 83, A: 255},    // 1 dark-blue
		color.RGBA{R: 126, G: 37, B: 83, A: 255},   // 2 dark-purple
		color.RGBA{R: 0, G: 135, B: 81, A: 255},    // 3 dark-green
		color.RGBA{R: 171, G: 82, B: 54, A: 255},   // 4 brown
		color.RGBA{R: 95, G: 87, B: 79, A: 255},    // 5 dark-gray
		color.RGBA{R: 194, G: 195, B: 199, A: 255}, // 6 light-gray
		color.RGBA{R: 255, G: 241, B: 232, A: 255}, // 7 white
		color.RGBA{R: 255, G: 0, B: 77, A: 255},    // 8 red
		color.RGBA{R: 255, G: 163, B: 0, A: 255},   // 9 orange
		color.RGBA{R: 255, G: 236, B: 39, A: 255},  // 10 yellow
		color.RGBA{R: 0, G: 228, B: 54, A: 255},    // 11 green
		color.RGBA{R: 41, G: 173, B: 255, A: 255},  // 12 blue
		color.RGBA{R: 131, G: 118, B: 156, A: 255}, // 13 indigo
		color.RGBA{R: 255, G: 119, B: 168, A: 255}, // 14 pink
		color.RGBA{R: 255, G: 204, B: 170, A: 255}, // 15 peach
	}

	// PaletteTransparency defines which colors in the Pico8Palette should be treated as transparent
	// By default, only color 0 (black) is transparent
	PaletteTransparency = []bool{true, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false}

	// pico8FaceSource is the loaded source for the PICO-8 TTF font.
	pico8FaceSource *text.GoTextFaceSource

	// DefaultFontSize is the default size used for the Print function.
	// PICO-8 font is typically 6px high.
	DefaultFontSize = 6.0

	// These variables hold the internal state for the cursor used by Print.
	// They were moved from main.go.
	cursorX     int
	cursorY     int
	cursorColor = 7 // Default to white (PICO-8 color 7)
)

func init() {
	// Initialize the font face source from embedded TTF.
	s, err := text.NewGoTextFaceSource(bytes.NewReader(pico8FontTTF))
	if err != nil {
		log.Fatalf("Failed to create font face source from pico-8.ttf: %v", err)
	}
	pico8FaceSource = s
}

// Cls clears the current drawing screen with a specified PICO-8 color index.
// Uses the internal `currentScreen` variable set by the engine.
// If no colorIndex is provided, it defaults to 0 (Black).
func Cls(colorIndex ...int) {
	if currentScreen == nil {
		log.Println("Warning: Cls() called before screen was ready.")
		return
	}
	idx := 0 // Default to black (index 0)
	if len(colorIndex) > 0 {
		idx = colorIndex[0]
	}

	if idx < 0 || idx >= len(Pico8Palette) {
		log.Printf("Warning: Cls() called with invalid color index %d. Defaulting to 0.", idx)
		idx = 0
	}
	currentScreen.Fill(Pico8Palette[idx])

	// Reset the global print cursor position
	cursorX = 0
	cursorY = 0
}

// Pget returns the PICO-8 color index (0-15) of the pixel at coordinates (x, y)
// on the current drawing screen.
// Uses the internal `currentScreen` variable.
//
// If the coordinates (x, y) are outside the screen bounds, it returns 0 (black).
// If the color at (x, y) does not exactly match any color in the Pico8Palette,
// it returns 0.
//
// Example:
//
//	func DrawExample() {
//	    // Set pixel at (10, 20) to red (index 8)
//	    Pset(10, 20, 8)
//
//	    // Get the color index of the pixel we just set
//	    pixelColorIndex := Pget(10, 20)
//
//	    // Print the retrieved color index (will print 8)
//	    // (Requires fmt package)
//	    // fmt.Printf("Color index at (10, 20): %d\n", pixelColorIndex)
//	}
func Pget(x, y int) int {
	if currentScreen == nil {
		log.Println("Warning: Pget() called before screen was ready.")
		return 0
	}
	bounds := currentScreen.Bounds()
	// Check if coordinates are outside the image bounds
	if x < bounds.Min.X || x >= bounds.Max.X || y < bounds.Min.Y || y >= bounds.Max.Y {
		return 0 // PICO-8 pget returns 0 for out-of-bounds
	}

	// Get the color at the specified pixel
	pixelColor := currentScreen.At(x, y)
	// Retrieve RGBA values as uint32 (premultiplied alpha, 0-65535 range)
	r1, g1, b1, a1 := pixelColor.RGBA()

	// Iterate through the palette to find a match
	for i, paletteColor := range Pico8Palette {
		// Retrieve RGBA values for the palette color in the same format
		r2, g2, b2, a2 := paletteColor.RGBA()
		// Compare the raw uint32 values
		if r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2 {
			return i // Return the matching palette index
		}
	}

	// Optional: Log if a non-palette color is encountered, though ideally only palette colors are used.
	// log.Printf("Warning: Pget(%d, %d) found color %v not in Pico8Palette.", x, y, pixelColor)

	// If no exact match is found in the palette, return 0
	return 0
}

// Pset draws a single pixel at coordinates (x, y) on the current drawing screen
// using the specified PICO-8 color index or the current cursorColor.
// Uses the internal `currentScreen` variable.
//
// The color is specified by its index (0-15) in the standard Pico8Palette.
// If no colorIndex is provided, the current cursorColor is used.
//
// If the coordinates (x, y) are outside the screen bounds, the function does nothing.
// If an invalid colorIndex is provided (e.g., < 0 or > 15), a warning is logged,
// and the function does nothing.
//
// Example:
//
//	Cursor(0, 0, 8) // Set current color to red
//	Pset(10, 20) // Draws a red pixel at (10, 20)
//	Pset(50, 50, 12) // Draws a blue pixel at (50, 50), color overrides cursorColor
func Pset(x, y int, colorIndex ...int) {
	// Check if screen is ready
	if currentScreen == nil {
		log.Println("Warning: Pset() called before screen was ready.")
		return
	}

	// Determine color to use
	color := cursorColor // Default to current cursor color
	if len(colorIndex) > 0 {
		color = colorIndex[0]
		if color < 0 || color >= len(Pico8Palette) {
			log.Printf("Warning: Pset() called with invalid color index %d. Palette has %d colors. Ignoring.", color, len(Pico8Palette))
			return
		}
	}

	// Apply camera offset
	fx, fy := ApplyCameraOffset(float64(x), float64(y))
	x, y = int(fx), int(fy)

	// Check bounds
	if x < 0 || x >= ScreenWidth || y < 0 || y >= ScreenHeight {
		return // Silently ignore out-of-bounds pixels
	}

	// Set the pixel with proper alpha handling
	pixelColor := Pico8Palette[color]

	// Check if this is a transparent color (binary transparency from PaletteTransparency)
	if PaletteTransparency[color] {
		// Don't draw transparent pixels (binary transparency)
		return
	}

	// Draw the pixel with proper alpha blending
	currentScreen.Set(x, y, pixelColor)
}

const (
	// CharWidthApproximation approximates character width for PICO-8 font for measurement.
	CharWidthApproximation = 4.0
)

// Cursor sets the implicit print cursor position (x, y) and optionally the default draw color.
// It mimics the PICO-8 CURSOR(x, y, [color]) function.
// Calling Cursor() with no arguments resets the cursor position to (0, 0) but leaves the color unchanged.
//
// Args:
//   - args: Optional arguments interpreted as [x, y] or [x, y, colorIndex].
//   - If len(args) == 0: Resets cursor position to (0, 0).
//   - If len(args) == 2: Sets cursor position to (args[0], args[1]).
//   - If len(args) >= 3: Sets cursor position to (args[0], args[1]) and sets currentDrawColor to args[2].
//
// Example:
//
//	Cursor(10, 20)     // Set cursor to (10, 20)
//	Cursor(30, 40, 5) // Set cursor to (30, 40) and draw color to 5 (dark gray)
//	Cursor()          // Reset cursor position to (0, 0)
func Cursor(args ...int) {
	switch len(args) {
	case 0:
		cursorX = 0
		cursorY = 0
	case 2:
		cursorX = args[0]
		cursorY = args[1]
	case 3:
		cursorX = args[0]
		cursorY = args[1]
		// Validate and set color
		col := args[2]
		if col < 0 || col >= len(Pico8Palette) {
			log.Printf("Warning: Cursor() called with invalid color index %d. Color not changed.", col)
		} else {
			// Update both color variables to keep them in sync
			cursorColor = col
			currentDrawColor = col
		}
	default:
		log.Printf("Warning: Cursor() called with invalid number of arguments (%d). Expected 0, 2, or 3.", len(args))
	}
}

// Print draws the given string onto the current drawing screen.
// Uses the internal `currentScreen` variable.
// It mimics the PICO-8 PRINT(str, [x, y], [color]) function, including implicit cursor tracking.
// It returns the X and Y coordinates of the pixel immediately following the printed string.
//
// Args:
//   - str: The string to print.
//   - args: Optional arguments interpreted based on PICO-8 logic:
//   - If len(args) == 0: Prints at current cursor (cursorX, cursorY) with current cursorColor.
//   - If len(args) == 1: Prints at current cursor (cursorX, cursorY) with color args[0] (overrides cursorColor).
//   - If len(args) == 2: Prints starting at (args[0], args[1]) with current cursorColor.
//   - If len(args) >= 3: Prints starting at (args[0], args[1]) with color args[2] (overrides cursorColor).
//
// Returns:
//   - int: The X coordinate after the string (drawX + stringWidth).
//   - int: The Y coordinate after the string (drawY + fontHeight).
//
// Limitations:
//   - Returned Width: Width calculation is an approximation (char_count * 4).
//
// Example:
//
//	// Assume cursor starts at (0, 0), color is 7 (white)
//	Cursor(0, 0, 6) // Set current color to light gray
//	_, _ = Print("1 HELLO")         // Draws at (0,0) in light gray, cursor moves to (0, 6).
//	_, _ = Print("2 WORLD", 8)      // Draws at (0,6) in red, cursor moves to (0, 12).
//	_, _ = Print("3 AT", 20, 20)     // Draws at (20,20) in light gray, cursor moves to (20, 26).
//	endX, endY := Print("4 DONE")    // Draws at (20, 26) in light gray, cursor moves to (20, 32).
func Print(str string, args ...int) (int, int) {
	// Check if screen is ready
	if currentScreen == nil {
		log.Println("Warning: Print() called before screen was ready.")

		// Calculate return values based on arguments without changing cursor state
		posX, posY := cursorX, cursorY
		if len(args) >= 2 {
			posX, posY = args[0], args[1]
		}

		// Approximate measurement for return value
		advance := float64(len([]rune(str))) * CharWidthApproximation
		endX := int(math.Ceil(float64(posX) + advance))
		endY := posY + int(DefaultFontSize)

		return endX, endY
	}

	// Parse arguments
	posX, posY, col := cursorX, cursorY, cursorColor // Use global cursorColor as default

	// If a new position is provided, override posX and posY.
	if len(args) >= 2 {
		posX, posY = args[0], args[1]
	}
	// If a color is provided (in len(args)==1 for color only,
	// or len(args)==3 for position and color), use the last argument.
	if len(args) == 1 || len(args) == 3 {
		col = args[len(args)-1]
		// Update both color variables to keep them in sync
		if col >= 0 && col < len(Pico8Palette) {
			cursorColor = col
			currentDrawColor = col
		}
	}

	// Validate the color index.
	if col < 0 || col >= len(Pico8Palette) {
		log.Printf("Warning: Print() called with invalid color index %d. Defaulting to cursorColor (%d).", col, cursorColor)
		col = cursorColor // Default to current cursorColor if invalid index given
	}

	// Apply camera offset
	fx, fy := ApplyCameraOffset(float64(posX), float64(posY))
	drawX, drawY := int(fx), int(fy)

	// --- Prepare for Drawing ---
	face := &text.GoTextFace{
		Source: pico8FaceSource,
		Size:   DefaultFontSize,
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(drawX), float64(drawY))
	op.ColorScale.ScaleWithColor(Pico8Palette[col])

	// --- Approximate Measurement for Return Value ---
	advance := float64(len([]rune(str))) * CharWidthApproximation
	endX := int(math.Ceil(float64(posX) + advance))
	endY := posY + int(DefaultFontSize)

	// --- Draw ---
	text.Draw(currentScreen, str, face, op)

	// --- Update Cursor Position ---
	// If a position was explicitly provided, use that; otherwise, keep the current cursorX.
	if len(args) >= 2 {
		cursorX = args[0]
	} else {
		cursorX = posX
	}
	cursorY = endY

	return endX, endY
}

// Palt sets the transparency for a specific color in the palette.
// When called with no arguments, it resets all colors to default transparency
// (only black is transparent).
//
// Args:
//   - color: A color number from the PICO-8 palette (0-15).
//   - transparent: true to make the color transparent, false to make it opaque.
//
// Example:
//
//	Spr(1, 10, 10)  // Draw sprite with default transparency (black is transparent)
//	Palt(8, true)   // Make red (color 8) transparent
//	Spr(1, 20, 10)  // Draw sprite with red transparent
//	Palt()          // Reset to default transparency
func Palt(args ...interface{}) {
	// If called with no arguments, reset to default transparency settings
	if len(args) == 0 {
		// Default: only black (color 0) is transparent
		for i := range PaletteTransparency {
			PaletteTransparency[i] = (i == 0)
		}
		return
	}

	// Extract color argument
	colorIndex, ok := args[0].(int)
	if !ok {
		// Try to convert from float64 if provided
		if colorFloat, floatOk := args[0].(float64); floatOk {
			colorIndex = int(colorFloat)
		} else {
			log.Printf("Warning: Palt() called with invalid color type: %T", args[0])
			return
		}
	}

	// Validate color index
	if colorIndex < 0 || colorIndex >= len(PaletteTransparency) {
		log.Printf("Warning: Palt() called with out-of-range color index: %d", colorIndex)
		return
	}

	// Extract transparent argument
	if len(args) < 2 {
		log.Printf("Warning: Palt() called without transparency value")
		return
	}

	transparent, ok := args[1].(bool)
	if !ok {
		log.Printf("Warning: Palt() called with invalid transparency type: %T", args[1])
		return
	}

	// Set the transparency for the specified color
	PaletteTransparency[colorIndex] = transparent
}

// --- Palette Management Functions ---

// SetPalette replaces the current color palette with a new one.
// This also resizes the transparency array to match the new palette size,
// setting only the first color (index 0) as transparent by default.
//
// newPalette: Slice of color.Color values to use as the new palette.
//
// Example:
//
//	// Create a 4-color grayscale palette
//	grayscale := []color.Color{
//		color.RGBA{0, 0, 0, 255},       // Black
//		color.RGBA{85, 85, 85, 255},    // Dark Gray
//		color.RGBA{170, 170, 170, 255}, // Light Gray
//		color.RGBA{255, 255, 255, 255}, // White
//	}
//	SetPalette(grayscale)
func SetPalette(newPalette []color.Color) {
	if len(newPalette) == 0 {
		log.Println("Warning: Attempted to set empty palette. Ignoring.")
		return
	}

	// Replace the palette
	Pico8Palette = newPalette

	// Resize transparency array to match
	oldTransparency := PaletteTransparency
	PaletteTransparency = make([]bool, len(newPalette))

	// Copy over existing transparency settings for indices that still exist
	for i := 0; i < len(PaletteTransparency) && i < len(oldTransparency); i++ {
		PaletteTransparency[i] = oldTransparency[i]
	}

	// Ensure at least the first color is transparent if we have any colors
	if len(PaletteTransparency) > 0 {
		PaletteTransparency[0] = true
	}
}

// Transparency is controlled using the Palt() function

// GetPaletteSize returns the current number of colors in the palette.
//
// Example:
//
//	// Get the current palette size
//	size := GetPaletteSize()
//	Print(fmt.Sprintf("Palette has %d colors", size), 10, 10, 7)
func GetPaletteSize() int {
	return len(Pico8Palette)
}

// GetPaletteColor returns the color.Color at the specified index in the palette.
// Returns nil if the index is out of range.
//
// colorIndex: Index of the color to retrieve.
//
// Example:
//
//	// Get the color at index 3
//	color3 := GetPaletteColor(3)
func GetPaletteColor(colorIndex int) color.Color {
	if colorIndex >= 0 && colorIndex < len(Pico8Palette) {
		return Pico8Palette[colorIndex]
	}
	return nil
}

// SetPaletteColor replaces a single color in the palette at the specified index.
// If the index is out of range, the function does nothing.
//
// colorIndex: Index of the color to replace.
// newColor: The new color.Color to use.
//
// Example:
//
//	// Change color 7 (white) to a light blue
//	SetPaletteColor(7, color.RGBA{200, 220, 255, 255})
func SetPaletteColor(colorIndex int, newColor color.Color) {
	if colorIndex >= 0 && colorIndex < len(Pico8Palette) {
		Pico8Palette[colorIndex] = newColor
	} else {
		log.Printf("Warning: Attempted to set color at out-of-range index %d. Palette has %d colors.",
			colorIndex, len(Pico8Palette))
	}
}

// --- Transparency Functions ---

// No alpha transparency functions needed - using only binary transparency
// where colors are either fully visible or fully transparent
