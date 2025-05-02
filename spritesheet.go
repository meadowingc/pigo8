package pigo8

import (
	"encoding/json" // Keep color import
	"fmt"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

// --- Structs to match spritesheet.json ---

// FlagsData holds sprite flag information.
// Exported because it's part of the exported SpriteInfo struct.
type FlagsData struct { // Exported
	Bitfield   int    `json:"bitfield"`
	Individual []bool `json:"individual"`
}

// spriteData holds the raw data for a single sprite from JSON.
// Kept internal as it's only used during loading.
type spriteData struct { // Internal
	ID     int       `json:"id"`
	X      int       `json:"x"`
	Y      int       `json:"y"`
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Pixels [][]int   `json:"pixels"`
	Flags  FlagsData `json:"flags"` // Uses exported FlagsData
	Used   bool      `json:"used"`
}

// spriteSheet holds the overall structure of the JSON file.
// Kept internal.
type spriteSheet struct { // Internal
	Sprites []spriteData `json:"sprites"`
}

// --- Target struct to hold processed sprite info ---

// SpriteInfo holds the processed, ready-to-use sprite data.
// Exported for use in main.go.
type SpriteInfo struct { // Exported
	ID    int
	Image *ebiten.Image
	Flags FlagsData
}

// --- Functions to load and process the spritesheet ---

// loadSpritesheetFromData processes sprite data provided as a byte slice.
// This allows users to load the spritesheet.json using go:embed or other methods
// in their own code (enabling build-time checks) and pass the data directly.
func loadSpritesheetFromData(data []byte) ([]SpriteInfo, error) {
	// Basic check if data is empty
	if len(data) == 0 {
		return nil, fmt.Errorf("provided spritesheet data is empty")
	}

	// Unmarshal the JSON data
	var sheet spriteSheet
	err := json.Unmarshal(data, &sheet)
	if err != nil {
		// Return a clear error about unmarshalling
		return nil, fmt.Errorf("error unmarshalling provided spritesheet data: %w", err)
	}

	// Add a check to see if sprites were loaded
	if len(sheet.Sprites) == 0 {
		// Log warning here as it's about content, not file loading
		log.Printf(
			"Warning: No sprites found after unmarshalling spritesheet data. Check JSON format and tags.",
		)
		// Return empty slice, not necessarily an error
		return []SpriteInfo{}, nil
	}
	// Check if pixel data is present for the first sprite (if any)
	if len(sheet.Sprites) > 0 && len(sheet.Sprites[0].Pixels) == 0 {
		log.Printf(
			"Warning: First sprite has empty pixel data after unmarshalling. Check JSON tags, especially for 'pixels'.",
		)
	}

	// Process used sprites
	var loadedSprites []SpriteInfo
	for _, spriteData := range sheet.Sprites {
		if !spriteData.Used {
			continue // Skip unused sprites
		}

		// Check if pixel data is empty for this specific sprite
		if len(spriteData.Pixels) == 0 ||
			(len(spriteData.Pixels) > 0 && len(spriteData.Pixels[0]) == 0) {
			log.Printf(
				"Warning: Skipping sprite %d due to missing or empty pixel data.",
				spriteData.ID,
			)
			continue
		}

		// Create a new Ebiten image for the sprite
		img := ebiten.NewImage(spriteData.Width, spriteData.Height)

		// Iterate over pixels and set colors based on the palette
		for y, row := range spriteData.Pixels {
			for x, colorIndex := range row {
				// Use Pico8Palette (defined in screen.go, same package)
				if colorIndex >= 0 && colorIndex < len(Pico8Palette) {
					// PICO-8 color 0 is often transparent
					if colorIndex != 0 {
						img.Set(x, y, Pico8Palette[colorIndex])
					}
				} else {
					log.Printf("Warning: Sprite %d has out-of-range color index %d at (%d, %d)", spriteData.ID, colorIndex, x, y)
				}
			}
		}

		// Create the SpriteInfo struct
		info := SpriteInfo{
			ID:    spriteData.ID,
			Image: img,
			Flags: spriteData.Flags,
		}
		loadedSprites = append(loadedSprites, info)
	}

	if len(loadedSprites) == 0 &&
		len(sheet.Sprites) > 0 { // Only warn if sprites existed but none were 'used'
		log.Printf(
			"Warning: No 'used' sprites were processed. Check the 'used' field in your spritesheet data.",
		)
	}

	return loadedSprites, nil
}

// loadSpritesheet loads sprite data from spritesheet.json in the current directory.
// This performs a runtime check for the file. For build-time verification,
// consider using go:embed in your application code and calling LoadSpritesheetFromData.
func loadSpritesheet() ([]SpriteInfo, error) {
	const spritesheetFilename = "spritesheet.json"
	
	// Log memory before loading spritesheet
	logMemory("before spritesheet load", false)
	
	data, err := os.ReadFile(spritesheetFilename)
	if err != nil {
		if os.IsNotExist(err) {
			// Create the detailed error for file not found
			detailedError := fmt.Errorf(
				"%w. Ensure '%s' exists in your game's base directory. "+
					"This file is required for sprite functionality and can be generated using the 'parsepico' tool (found at https://github.com/drpaneas/parsepico)",
				err, spritesheetFilename,
			)
			return nil, detailedError // Return the specific error
		}
		// Handle other potential read errors - log here as it's unexpected at runtime
		log.Printf("Error reading %s: %v", spritesheetFilename, err)
		return nil, fmt.Errorf(
			"error reading %s: %w",
			spritesheetFilename,
			err,
		) // Wrap other errors
	}
	
	// Log memory after reading file
	logMemory("after reading spritesheet file", false)
	
	// Call the data-processing function
	sprites, err := loadSpritesheetFromData(data)
	if err != nil {
		return nil, err
	}
	
	// Always log when spritesheet is loaded, regardless of memory change
	fileSize := float64(len(data)) / 1024
	log.Printf("Spritesheet: %d sprites (%.1f KB)", len(sprites), fileSize)
	logMemory("after spritesheet load", true)
	
	return sprites, nil
}
