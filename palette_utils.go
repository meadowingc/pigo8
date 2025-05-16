package pigo8

import (
	"image/color"
	"math"
)

// weights based on ITU-R BT.709 standard for the RGB to grayscale conversion for HDTVs.
// These weights account for human perception of brightness in different color channels.
const (
	rWeight = 0.2126
	gWeight = 0.7152
	bWeight = 0.0722
)

// getColorLuminance calculates the perceived brightness (luminance) of a color
// using the ITU-R BT.709 standard weights for RGB to grayscale conversion.
// Returns a value between 0 (darkest) and 255 (brightest).
// The alpha channel is explicitly ignored in the calculation.
func getColorLuminance(c color.Color) float64 {
	// Convert to RGBA to ensure we're working with RGB values
	rgba := color.RGBAModel.Convert(c).(color.RGBA)

	// Calculate luminance using the standard weights with explicit RGB components
	// The alpha channel is not used in the calculation
	return rWeight*float64(rgba.R) + gWeight*float64(rgba.G) + bWeight*float64(rgba.B)
}

// findDarkestColorIndex returns the index of the darkest color in the current palette,
// excluding the color at index 0. If the darkest color is at index 0, it returns the
// next darkest color. Returns 1 if no other colors are available.
// The darkness is determined by calculating the luminance of each color.
// In case of a tie, the color with the lower index is returned.
func findDarkestColorIndex() int {
	darkestIndex := 1 // Start with index 1 as default (skip index 0)
	lowestLuminance := math.MaxFloat64

	// Iterate through all colors in the palette, starting from index 1
	for i := 1; i < GetPaletteSize(); i++ {
		c := GetPaletteColor(i)
		if c == nil {
			continue
		}

		luminance := getColorLuminance(c)

		// Update if this color is darker (lower luminance)
		if luminance < lowestLuminance {
			lowestLuminance = luminance
			darkestIndex = i
		}
	}

	// If we only have one color (index 0), return index 1 (which would be out of bounds,
	// but that's handled by GetPaletteColor)
	if GetPaletteSize() <= 1 {
		return 1
	}

	return darkestIndex
}

// findLightestColorIndex returns the index of the lightest color in the current palette,
// excluding the color at index 0. If the lightest color is at index 0, it returns the
// next lightest color. Returns 1 if no other colors are available.
// The lightness is determined by calculating the luminance of each color.
// In case of a tie, the color with the lower index is returned.
func findLightestColorIndex() int {
	lightestIndex := 1 // Start with index 1 as default (skip index 0)
	highestLuminance := -1.0

	// Iterate through all colors in the palette, starting from index 1
	for i := 1; i < GetPaletteSize(); i++ {
		c := GetPaletteColor(i)
		if c == nil {
			continue
		}

		luminance := getColorLuminance(c)

		// Update if this color is lighter (higher luminance)
		if luminance > highestLuminance {
			highestLuminance = luminance
			lightestIndex = i
		}
	}

	// If we only have one color (index 0), return index 1 (which would be out of bounds,
	// but that's handled by GetPaletteColor)
	if GetPaletteSize() <= 1 {
		return 1
	}

	return lightestIndex
}

// findMidToneColorIndex returns the index of a color that is neither the lightest nor the darkest
// in the current palette, excluding the color at index 0. If no such color exists (e.g., palette has
// only 2 colors), it returns the lightest color. Returns 1 if no other colors are available.
// The function calculates the luminance of each color and finds the one closest to the middle
// of the luminance range.
func findMidToneColorIndex() int {
	if GetPaletteSize() <= 2 {
		return 1
	}

	// Find the lightest and darkest colors to determine the range
	lightestIdx := findLightestColorIndex()
	darkestIdx := findDarkestColorIndex()

	// If we only have two colors, return the lighter one
	if lightestIdx == darkestIdx || GetPaletteSize() <= 3 {
		return lightestIdx
	}

	// Get the luminance range
	lightestLum := getColorLuminance(GetPaletteColor(lightestIdx))
	darkestLum := getColorLuminance(GetPaletteColor(darkestIdx))
	midLum := (lightestLum + darkestLum) / 2.0

	// Find the color with luminance closest to the middle
	midToneIndex := 1
	closestDiff := math.MaxFloat64

	// Iterate through all colors in the palette, starting from index 1
	for i := 1; i < GetPaletteSize(); i++ {
		// Skip the lightest and darkest colors
		if i == lightestIdx || i == darkestIdx {
			continue
		}

		c := GetPaletteColor(i)
		if c == nil {
			continue
		}

		luminance := getColorLuminance(c)
		diff := math.Abs(luminance - midLum)

		// Update if this color is closer to the middle
		if diff < closestDiff {
			closestDiff = diff
			midToneIndex = i
		}
	}

	// If we didn't find any mid-tone colors (shouldn't happen with >2 colors),
	// return the lightest one as a fallback
	if midToneIndex == 1 && GetPaletteSize() > 2 {
		return lightestIdx
	}

	return midToneIndex
}
