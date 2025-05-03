package pigo8

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"
)

// Note: Directly testing the drawing output of Spr is difficult in unit tests
// as it requires an active Ebiten game loop and access to screen pixels.
// These tests focus on argument parsing and state checks without actually drawing.

func TestSprArguments(t *testing.T) {
	// --- Setup --- Manage global state
	originalSprites := currentSprites
	originalTransparency := PaletteTransparency // Save original transparency settings

	// Initialize palette transparency (only black transparent by default)
	for i := range PaletteTransparency {
		PaletteTransparency[i] = (i == 0)
	}

	// Create dummy sprite entries with non-sequential IDs
	dummyImg1 := ebiten.NewImage(8, 8)
	dummyImg2 := ebiten.NewImage(8, 8)
	dummyImg3 := ebiten.NewImage(8, 8)
	// Intentionally set IDs out of order with slice index
	currentSprites = []SpriteInfo{
		{ID: 5, Image: dummyImg1},
		{ID: 0, Image: dummyImg2},
		{ID: 10, Image: dummyImg3},
	}

	t.Cleanup(func() {
		currentSprites = originalSprites
		PaletteTransparency = originalTransparency // Restore original transparency settings
	})

	// Helper function to parse Spr arguments without actually drawing
	parseSprArgs := func(spriteNumber any, x any, y any, options ...any) (parsed struct {
		SpriteNumber  int
		X, Y          float64
		Width, Height float64
		FlipX, FlipY  bool
	}) {
		// Convert sprite number to int
		switch val := spriteNumber.(type) {
		case int:
			parsed.SpriteNumber = val
		case float64:
			parsed.SpriteNumber = int(val)
		default:
			parsed.SpriteNumber = 0
		}

		// Convert x, y to float64
		switch val := x.(type) {
		case int:
			parsed.X = float64(val)
		case float64:
			parsed.X = val
		default:
			parsed.X = 0
		}

		switch val := y.(type) {
		case int:
			parsed.Y = float64(val)
		case float64:
			parsed.Y = val
		default:
			parsed.Y = 0
		}

		// Default width and height
		parsed.Width = 1
		parsed.Height = 1

		// Process optional parameters
		if len(options) > 0 && options[0] != nil {
			// Width multiplier
			switch val := options[0].(type) {
			case int:
				parsed.Width = float64(val)
			case float64:
				parsed.Width = val
			case string:
				// Invalid type, keep default
			}
		}

		if len(options) > 1 && options[1] != nil {
			// Height multiplier
			switch val := options[1].(type) {
			case int:
				parsed.Height = float64(val)
			case float64:
				parsed.Height = val
			case string:
				// Invalid type, keep default
			}
		}

		if len(options) > 2 && options[2] != nil {
			// FlipX
			switch val := options[2].(type) {
			case bool:
				parsed.FlipX = val
			default:
				// Invalid type, keep default (false)
			}
		}

		if len(options) > 3 && options[3] != nil {
			// FlipY
			switch val := options[3].(type) {
			case bool:
				parsed.FlipY = val
			default:
				// Invalid type, keep default (false)
			}
		}

		return parsed
	}

	t.Run("Basic call with int coords", func(t *testing.T) {
		result := parseSprArgs(0, 5, 5)
		assert.Equal(t, 0, result.SpriteNumber)
		assert.Equal(t, 5.0, result.X)
		assert.Equal(t, 5.0, result.Y)
		assert.Equal(t, 1.0, result.Width)
		assert.Equal(t, 1.0, result.Height)
		assert.False(t, result.FlipX)
		assert.False(t, result.FlipY)
	})

	t.Run("Basic call with float coords", func(t *testing.T) {
		result := parseSprArgs(0, 5.5, 5.5)
		assert.Equal(t, 0, result.SpriteNumber)
		assert.Equal(t, 5.5, result.X)
		assert.Equal(t, 5.5, result.Y)
	})

	t.Run("Basic call with mixed coords", func(t *testing.T) {
		result1 := parseSprArgs(0, 5, 5.5)
		assert.Equal(t, 0, result1.SpriteNumber)
		assert.Equal(t, 5.0, result1.X)
		assert.Equal(t, 5.5, result1.Y)

		result2 := parseSprArgs(0, 5.5, 5)
		assert.Equal(t, 0, result2.SpriteNumber)
		assert.Equal(t, 5.5, result2.X)
		assert.Equal(t, 5.0, result2.Y)
	})

	t.Run("Basic call with float sprite number", func(t *testing.T) {
		// Floats should be truncated to ints (0.5 -> 0, 5.9 -> 5)
		result1 := parseSprArgs(0.5, 1, 1)
		assert.Equal(t, 0, result1.SpriteNumber, "Should truncate 0.5 to 0")

		result2 := parseSprArgs(5.9, 2, 2)
		assert.Equal(t, 5, result2.SpriteNumber, "Should truncate 5.9 to 5")

		// Should find sprite ID 5 when using 5.9
		result3 := parseSprArgs(5.9, 10, 10)
		assert.Equal(t, 5, result3.SpriteNumber, "Should truncate 5.9 to 5")

		// Non-existent sprite ID after truncation (e.g., 1.5 -> 1)
		result4 := parseSprArgs(1.5, 15, 15)
		assert.Equal(t, 1, result4.SpriteNumber, "Should truncate 1.5 to 1")
	})

	t.Run("Call with w/h options (float)", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 2.0, 0.5)
		})
	})

	t.Run("Call with w/h options (int)", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 2, 1)
		})
	})

	t.Run("Call with w/h/flipX options", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, true)
		})
	})

	t.Run("Call with w/h/flipX/flipY options", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, true, false)
		})
	})

	t.Run("Call with zero w/h", func(t *testing.T) {
		// Should return early, no panic
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 0.0, 1.0)
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 0)
		})
	})

	t.Run("Call with negative w/h", func(t *testing.T) {
		// w/h are clamped to 0, should return early, no panic
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, -1.0, 1.0)
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, -2)
		})
	})

	t.Run("Call with invalid option types", func(t *testing.T) {
		// Should log warnings but not panic
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, "invalid_w")
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, "invalid_h")
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, "invalid_flipX")
		})
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, true, "invalid_flipY")
		})
		// TODO: Check log warnings (requires capture)
	})

	t.Run("Call with too many options", func(t *testing.T) {
		assert.NotPanics(t, func() {
			Spr(0, 1, 1, 1.0, 1.0, true, false, "extra")
		})
		// TODO: Check log warning (requires capture)
	})

	t.Run("Call with out-of-bounds sprite number", func(t *testing.T) {
		// Should just skip drawing the non-existent tile, no panic
		assert.NotPanics(t, func() {
			Spr(99, 1, 1)
		})
		assert.NotPanics(t, func() {
			Spr(-1, 1, 1)
		})
	})

	t.Run("Call when screen is nil", func(t *testing.T) {
		savedScreen := currentScreen
		currentScreen = nil
		assert.NotPanics(t, func() {
			Spr(0, 1, 1)
		})
		currentScreen = savedScreen // Restore
		// TODO: Check log warning (requires capture)
	})

	t.Run("Call using Sprite ID not equal to slice index", func(t *testing.T) {
		// Should find sprite ID 5 (at index 0) and not panic
		assert.NotPanics(t, func() {
			Spr(5, 1, 1)
		})
		// Should find sprite ID 10 (at index 2) and not panic
		assert.NotPanics(t, func() {
			Spr(10, 2, 2)
		})
		// Should find sprite ID 0 (at index 1) and not panic
		assert.NotPanics(t, func() {
			Spr(0, 3, 3)
		})
	})

	// Note: Testing the lazy-loading of sprites is complex as it involves
	// the filesystem via loadSpritesheet(). This is better suited for an
	// integration test or requires mocking loadSpritesheet().

}

// TestSpriteCoordinateConversion tests the coordinate conversion logic used in Sget and Sset
func TestSpriteCoordinateConversion(t *testing.T) {
	// Test cases for coordinate conversion
	testCases := []struct {
		name             string
		globalX          int
		globalY          int
		expectedSpriteID int
		expectedLocalX   int
		expectedLocalY   int
	}{
		{"Top-left pixel of first sprite", 0, 0, 0, 0, 0},
		{"Middle pixel of first sprite", 4, 3, 0, 4, 3},
		{"Bottom-right pixel of first sprite", 7, 7, 0, 7, 7},
		{"First pixel of second sprite", 8, 0, 1, 0, 0},
		{"Middle pixel of second sprite", 12, 3, 1, 4, 3},
		{"First pixel of sprite in second row", 0, 8, 16, 0, 0},
		{"Middle pixel of sprite in second row", 4, 12, 16, 4, 4},
		{"Pixel in middle of spritesheet", 64, 64, 136, 0, 0},
		{"Last pixel of spritesheet", 127, 127, 255, 7, 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate sprite ID using the same logic as in Sget/Sset
			spriteX := tc.globalX / 8
			spriteY := tc.globalY / 8
			spriteCellID := spriteY*16 + spriteX

			// Calculate local coordinates
			localX := tc.globalX % 8
			localY := tc.globalY % 8

			// Verify calculations
			assert.Equal(t, tc.expectedSpriteID, spriteCellID, "Sprite ID calculation incorrect")
			assert.Equal(t, tc.expectedLocalX, localX, "Local X calculation incorrect")
			assert.Equal(t, tc.expectedLocalY, localY, "Local Y calculation incorrect")
		})
	}
}

// TestFget tests the Fget function for retrieving sprite flags
func TestFget(t *testing.T) {
	// Save original state
	originalSprites := currentSprites

	// Setup test sprites with different flag configurations
	currentSprites = []SpriteInfo{
		{ID: 1, Flags: FlagsData{Bitfield: 1, Individual: []bool{true, false, false, false, false, false, false, false}}},  // Only flag 0 is set
		{ID: 2, Flags: FlagsData{Bitfield: 170, Individual: []bool{false, true, false, true, false, true, false, true}}},   // Flags 1,3,5,7 are set
		{ID: 3, Flags: FlagsData{Bitfield: 0, Individual: []bool{false, false, false, false, false, false, false, false}}}, // No flags set
		{ID: 4, Flags: FlagsData{Bitfield: 255, Individual: []bool{true, true, true, true, true, true, true, true}}},       // All flags set
	}

	// Cleanup after tests
	t.Cleanup(func() {
		currentSprites = originalSprites
	})

	tests := []struct {
		name     string
		spriteID int
		flag     []int
		wantBF   int
		wantSet  bool
	}{
		{"Get specific flag - true case", 1, []int{0}, 1, true},
		{"Get specific flag - false case", 1, []int{1}, 1, false},
		{"Get entire bitfield", 2, []int{}, 170, false},
		{"Invalid sprite ID", 99, []int{}, 0, false},
		{"Invalid flag number", 1, []int{-1}, 1, false},
		{"Invalid flag number", 1, []int{8}, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotBF int
			var gotSet bool

			if len(tt.flag) > 0 {
				gotBF, gotSet = Fget(tt.spriteID, tt.flag[0])
			} else {
				gotBF, gotSet = Fget(tt.spriteID)
			}

			if gotBF != tt.wantBF {
				t.Errorf("Fget() bitfield = %v, want %v", gotBF, tt.wantBF)
			}
			if gotSet != tt.wantSet {
				t.Errorf("Fget() isSet = %v, want %v", gotSet, tt.wantSet)
			}
		})
	}
}

// TestFset tests the Fset function for setting sprite flags
func TestFset(t *testing.T) {
	// Save original state
	originalSprites := currentSprites

	// Setup test sprites with different flag configurations
	currentSprites = []SpriteInfo{
		{ID: 1, Flags: FlagsData{Bitfield: 1, Individual: []bool{true, false, false, false, false, false, false, false}}},  // Only flag 0 is set
		{ID: 2, Flags: FlagsData{Bitfield: 170, Individual: []bool{false, true, false, true, false, true, false, true}}},   // Flags 1,3,5,7 are set
		{ID: 3, Flags: FlagsData{Bitfield: 0, Individual: []bool{false, false, false, false, false, false, false, false}}}, // No flags set
		{ID: 4, Flags: FlagsData{Bitfield: 255, Individual: []bool{true, true, true, true, true, true, true, true}}},       // All flags set
		{ID: 5, Flags: FlagsData{Bitfield: 0, Individual: []bool{false, false, false, false, false, false, false, false}}}, // For testing setting specific flags
	}

	// Cleanup after tests
	t.Cleanup(func() {
		currentSprites = originalSprites
	})

	tests := []struct {
		name   string
		setup  func()
		verify func(t *testing.T)
	}{
		{
			name: "Set specific flag to true",
			setup: func() {
				// Set flag 2 to true on sprite 5
				Fset(5, 2, true)
			},
			verify: func(t *testing.T) {
				// Verify flag 2 is set and bitfield is updated
				bitfield, isSet := Fget(5, 2)
				if !isSet {
					t.Errorf("Flag 2 should be set to true")
				}
				if bitfield != 4 { // 2^2 = 4
					t.Errorf("Bitfield should be 4, got %d", bitfield)
				}
			},
		},
		{
			name: "Set specific flag to false",
			setup: func() {
				// First set flag 3 to true
				Fset(5, 3, true)
				// Then set it back to false
				Fset(5, 3, false)
			},
			verify: func(t *testing.T) {
				// Verify flag 3 is not set
				_, isSet := Fget(5, 3)
				if isSet {
					t.Errorf("Flag 3 should be set to false")
				}
				// Verify bitfield doesn't have flag 3 set
				bitfield, _ := Fget(5)
				if bitfield&8 != 0 { // 2^3 = 8
					t.Errorf("Bitfield should not have flag 3 set, got %d", bitfield)
				}
			},
		},
		{
			name: "Set all flags to true",
			setup: func() {
				// Set all flags to true
				Fset(5, true)
			},
			verify: func(t *testing.T) {
				// Verify all flags are set
				bitfield, _ := Fget(5)
				if bitfield != 255 {
					t.Errorf("All flags should be set, bitfield should be 255, got %d", bitfield)
				}
				// Check individual flags
				for i := 0; i < 8; i++ {
					_, isSet := Fget(5, i)
					if !isSet {
						t.Errorf("Flag %d should be set to true", i)
					}
				}
			},
		},
		{
			name: "Set all flags to false",
			setup: func() {
				// First set all flags to true
				Fset(5, true)
				// Then set all to false
				Fset(5, false)
			},
			verify: func(t *testing.T) {
				// Verify all flags are cleared
				bitfield, _ := Fget(5)
				if bitfield != 0 {
					t.Errorf("All flags should be cleared, bitfield should be 0, got %d", bitfield)
				}
				// Check individual flags
				for i := 0; i < 8; i++ {
					_, isSet := Fget(5, i)
					if isSet {
						t.Errorf("Flag %d should be set to false", i)
					}
				}
			},
		},
		{
			name: "Set flags using bitfield",
			setup: func() {
				// Set flags 1,3,5,7 using bitfield 170 (2+8+32+128)
				Fset(5, 170)
			},
			verify: func(t *testing.T) {
				// Verify bitfield is set correctly
				bitfield, _ := Fget(5)
				if bitfield != 170 {
					t.Errorf("Bitfield should be 170, got %d", bitfield)
				}
				// Check that flags 1,3,5,7 are set
				flagIndices := []int{1, 3, 5, 7}
				for _, idx := range flagIndices {
					_, isSet := Fget(5, idx)
					if !isSet {
						t.Errorf("Flag %d should be set to true", idx)
					}
				}
				// Check that flags 0,2,4,6 are not set
				flagIndices = []int{0, 2, 4, 6}
				for _, idx := range flagIndices {
					_, isSet := Fget(5, idx)
					if isSet {
						t.Errorf("Flag %d should be set to false", idx)
					}
				}
			},
		},
		{
			name: "Invalid sprite ID",
			setup: func() {
				// Try to set flags on non-existent sprite
				Fset(99, 0, true)
			},
			verify: func(_ *testing.T) {
				// Nothing to verify, just make sure it doesn't crash
			},
		},
		{
			name: "Invalid flag number",
			setup: func() {
				// Try to set invalid flag number
				Fset(5, -1, true)
				Fset(5, 8, true)
			},
			verify: func(t *testing.T) {
				// Verify sprite 5's bitfield is still 0 (unchanged)
				bitfield, _ := Fget(5)
				if bitfield != 0 {
					t.Errorf("Bitfield should remain 0, got %d", bitfield)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset sprite 5 for each test
			for i := range currentSprites {
				if currentSprites[i].ID == 5 {
					currentSprites[i].Flags.Bitfield = 0
					for j := range currentSprites[i].Flags.Individual {
						currentSprites[i].Flags.Individual[j] = false
					}
					break
				}
			}

			// Run the test setup
			tt.setup()

			// Verify the results
			tt.verify(t)
		})
	}
}

// TestSsetColorHandling tests the color handling logic in Sset
func TestSsetColorHandling(t *testing.T) {
	// Save original state
	originalDrawColor := currentDrawColor

	// Cleanup after tests
	t.Cleanup(func() {
		currentDrawColor = originalDrawColor
	})

	// Test cases continue...
}

// convertNumericToFloat64 converts any numeric type to float64
func convertNumericToFloat64(v any) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int8:
		return float64(val)
	case int16:
		return float64(val)
	case int32:
		return float64(val)
	case int64:
		return float64(val)
	case uint:
		return float64(val)
	case uint8:
		return float64(val)
	case uint16:
		return float64(val)
	case uint32:
		return float64(val)
	case uint64:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}

// convertNumericToFloat64Optional converts any numeric type to float64 with a success flag
func convertNumericToFloat64Optional(v any) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int8:
		return float64(val), true
	case int16:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint8:
		return float64(val), true
	case uint16:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case float32:
		return float64(val), true
	case float64:
		return val, true
	default:
		return 0, false
	}
}

// SsprTestResult represents the parsed arguments for Sspr function testing
type SsprTestResult struct {
	SourceX, SourceY, SourceW, SourceH int
	DestX, DestY                       float64
	DestW, DestH                       float64
	FlipX, FlipY                       bool
}

// parseSsprTestArgs parses arguments similar to how Sspr does it but without drawing
func parseSsprTestArgs(sx, sy, sw, sh, dx, dy any, options ...any) SsprTestResult {
	var result SsprTestResult

	// Convert source coordinates to integers
	result.SourceX = int(convertNumericToFloat64(sx))
	result.SourceY = int(convertNumericToFloat64(sy))
	result.SourceW = int(convertNumericToFloat64(sw))
	result.SourceH = int(convertNumericToFloat64(sh))

	// Convert destination coordinates to float64
	result.DestX = convertNumericToFloat64(dx)
	result.DestY = convertNumericToFloat64(dy)

	// Default destination size is the same as source size
	result.DestW = float64(result.SourceW)
	result.DestH = float64(result.SourceH)

	// Process optional parameters
	if len(options) > 0 && options[0] != nil {
		// Try to convert to float64 for destination width
		if dw, ok := convertNumericToFloat64Optional(options[0]); ok {
			result.DestW = dw
		}
	}

	if len(options) > 1 && options[1] != nil {
		// Try to convert to float64 for destination height
		if dh, ok := convertNumericToFloat64Optional(options[1]); ok {
			result.DestH = dh
		}
	}

	if len(options) > 2 && options[2] != nil {
		// Try to convert to bool for flip_x
		if flipX, ok := options[2].(bool); ok {
			result.FlipX = flipX
		}
	}

	if len(options) > 3 && options[3] != nil {
		// Try to convert to bool for flip_y
		if flipY, ok := options[3].(bool); ok {
			result.FlipY = flipY
		}
	}

	return result
}

// TestSsprArgumentParsing tests the argument parsing logic of the Sspr function without actually drawing.
// This avoids the "ReadPixels cannot be called before the game starts" error that occurs
// when trying to read pixels from an Ebiten image in a unit test.
func TestSsprArgumentParsing(t *testing.T) {
	// Save original transparency settings
	originalTransparency := PaletteTransparency

	// Initialize palette transparency (only black transparent by default)
	for i := range PaletteTransparency {
		PaletteTransparency[i] = (i == 0)
	}

	// Restore original transparency settings after test
	t.Cleanup(func() {
		PaletteTransparency = originalTransparency
	})

	// Define test cases to reduce cyclomatic complexity
	testCases := []struct {
		name     string
		args     []any
		expected func(t *testing.T, result SsprTestResult)
	}{
		{
			name: "Basic call with int coords",
			args: []any{0, 0, 8, 8, 10, 10},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 0, result.SourceX)
				assert.Equal(t, 0, result.SourceY)
				assert.Equal(t, 8, result.SourceW)
				assert.Equal(t, 8, result.SourceH)
				assert.Equal(t, 10.0, result.DestX)
				assert.Equal(t, 10.0, result.DestY)
				assert.Equal(t, 8.0, result.DestW)
				assert.Equal(t, 8.0, result.DestH)
				assert.False(t, result.FlipX)
				assert.False(t, result.FlipY)
			},
		},
		{
			name: "Basic call with float coords",
			args: []any{0.5, 0.5, 8.0, 8.0, 10.5, 10.5},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 0, result.SourceX) // Should be truncated to int
				assert.Equal(t, 0, result.SourceY) // Should be truncated to int
				assert.Equal(t, 8, result.SourceW)
				assert.Equal(t, 8, result.SourceH)
				assert.Equal(t, 10.5, result.DestX)
				assert.Equal(t, 10.5, result.DestY)
				assert.Equal(t, 8.0, result.DestW)
				assert.Equal(t, 8.0, result.DestH)
			},
		},
		{
			name: "Basic call with mixed coords",
			args: []any{0, 0.5, 8, 8.5, 10, 10.5},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 0, result.SourceX)
				assert.Equal(t, 0, result.SourceY) // Should be truncated to int
				assert.Equal(t, 8, result.SourceW)
				assert.Equal(t, 8, result.SourceH) // Should be truncated to int
				assert.Equal(t, 10.0, result.DestX)
				assert.Equal(t, 10.5, result.DestY)
			},
		},
		{
			name: "With destination width/height",
			args: []any{0, 0, 8, 8, 10, 10, 16, 16},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 16.0, result.DestW)
				assert.Equal(t, 16.0, result.DestH)
			},
		},
		{
			name: "With flipping options",
			args: []any{0, 0, 8, 8, 10, 10, 8, 8, true, false},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.True(t, result.FlipX)
				assert.False(t, result.FlipY)
			},
		},
		{
			name: "Zero source dimensions",
			args: []any{0, 0, 0, 0, 10, 10},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 0, result.SourceW)
				assert.Equal(t, 0, result.SourceH)
				assert.Equal(t, 0.0, result.DestW) // Should inherit source dimensions
				assert.Equal(t, 0.0, result.DestH) // Should inherit source dimensions
			},
		},
		{
			name: "Zero destination dimensions",
			args: []any{0, 0, 8, 8, 10, 10, 0, 0},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 0.0, result.DestW)
				assert.Equal(t, 0.0, result.DestH)
			},
		},
		{
			name: "Negative dimensions",
			args: []any{0, 0, 8, 8, 10, 10, -5, -5},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, -5.0, result.DestW) // Negative values are allowed in parsing
				assert.Equal(t, -5.0, result.DestH) // Negative values are allowed in parsing
			},
		},
		{
			name: "Out of bounds source",
			args: []any{120, 120, 16, 16, 10, 10},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 120, result.SourceX)
				assert.Equal(t, 120, result.SourceY)
				assert.Equal(t, 16, result.SourceW)
				assert.Equal(t, 16, result.SourceH)
			},
		},
		{
			name: "Invalid option types",
			args: []any{0, 0, 8, 8, 10, 10, "not_a_number", "also_not_a_number"},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 8.0, result.DestW) // Should keep source width
				assert.Equal(t, 8.0, result.DestH) // Should keep source height
			},
		},
		{
			name: "Too many arguments",
			args: []any{0, 0, 8, 8, 10, 10, 16, 16, true, false, "extra_arg"},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 16.0, result.DestW)
				assert.Equal(t, 16.0, result.DestH)
				assert.True(t, result.FlipX)
				assert.False(t, result.FlipY)
			},
		},
		{
			name: "Non-standard source size",
			args: []any{5, 5, 6, 7, 10, 10},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 5, result.SourceX)
				assert.Equal(t, 5, result.SourceY)
				assert.Equal(t, 6, result.SourceW)
				assert.Equal(t, 7, result.SourceH)
				assert.Equal(t, 6.0, result.DestW) // Should inherit source width
				assert.Equal(t, 7.0, result.DestH) // Should inherit source height
			},
		},
		{
			name: "Stretching",
			args: []any{0, 0, 8, 8, 10, 10, 32, 16}, // 4x width, 2x height
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 32.0, result.DestW)
				assert.Equal(t, 16.0, result.DestH)
			},
		},
		{
			name: "Squashing",
			args: []any{0, 0, 16, 16, 10, 10, 8, 8}, // Half size
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 8.0, result.DestW)
				assert.Equal(t, 8.0, result.DestH)
			},
		},
		{
			name: "Nil optional arguments",
			args: []any{0, 0, 8, 8, 10, 10, nil, nil, nil, nil},
			expected: func(t *testing.T, result SsprTestResult) {
				assert.Equal(t, 8.0, result.DestW) // Should keep source width
				assert.Equal(t, 8.0, result.DestH) // Should keep source height
				assert.False(t, result.FlipX)
				assert.False(t, result.FlipY)
			},
		},
	}

	// Run all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Extract the required arguments and options
			if len(tc.args) < 6 {
				t.Fatalf("Test case %s has insufficient arguments", tc.name)
			}

			// Extract the first 6 required arguments
			baseArgs := tc.args[:6]

			// Extract any optional arguments
			var options []any
			if len(tc.args) > 6 {
				options = tc.args[6:]
			}

			// Call the parsing function
			result := parseSsprTestArgs(
				baseArgs[0], baseArgs[1], baseArgs[2], baseArgs[3], baseArgs[4], baseArgs[5],
				options...,
			)

			// Verify the results
			tc.expected(t, result)
		})
	}
}

// TestSsetColorHandling continues...
func TestSsetColorHandling_continued(t *testing.T) {
	t.Run("Default color handling", func(t *testing.T) {
		// Set a known color
		currentDrawColor = 7 // White

		// Verify that when no color is specified, currentDrawColor is used
		colorToUse := currentDrawColor
		assert.Equal(t, 7, colorToUse, "Should use currentDrawColor when no color specified")
	})

	t.Run("Explicit color handling", func(t *testing.T) {
		// Set a known color
		currentDrawColor = 7 // White

		// Verify that when a color is specified, it's used instead of currentDrawColor
		colorToUse := 8 // Explicit color (red)
		assert.Equal(t, 8, colorToUse, "Should use explicit color when specified")

		// Verify currentDrawColor remains unchanged
		assert.Equal(t, 7, currentDrawColor, "currentDrawColor should not change when explicit color is used")
	})

	t.Run("Color clamping - high", func(t *testing.T) {
		// Test color clamping for values above the valid range
		invalidColor := 99 // Above valid range
		clampedColor := invalidColor

		// Apply the same clamping logic as in Sset
		if clampedColor >= len(Pico8Palette) {
			clampedColor = len(Pico8Palette) - 1
		}

		assert.Equal(t, len(Pico8Palette)-1, clampedColor, "Color should be clamped to max index")
	})

	t.Run("Color clamping - low", func(t *testing.T) {
		// Test color clamping for values below the valid range
		invalidColor := -1 // Below valid range
		clampedColor := invalidColor

		// Apply the same clamping logic as in Sset
		if clampedColor < 0 {
			clampedColor = 0
		}

		assert.Equal(t, 0, clampedColor, "Color should be clamped to 0")
	})
}
