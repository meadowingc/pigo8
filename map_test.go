package pigo8

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMgetAndMset(t *testing.T) {
	// Save original map data
	originalMap := currentMap

	// Create a temporary map for testing
	testMap := &MapData{
		Version:     "1.0",
		Description: "Test Map",
		Width:       16,
		Height:      16,
		Name:        "test_map",
		Cells: []mapCell{
			{X: 0, Y: 0, Sprite: 1},
			{X: 1, Y: 1, Sprite: 2},
			{X: 2, Y: 2, Sprite: 3},
			{X: 5, Y: 7, Sprite: 42},
		},
	}

	// Set the test map as the current map
	currentMap = testMap

	// Restore original map after test
	t.Cleanup(func() {
		currentMap = originalMap
	})

	// Test Mget with different coordinate types
	t.Run("Mget with different coordinate types", func(t *testing.T) {
		// Integer coordinates
		assert.Equal(t, 1, Mget(0, 0), "Mget(0, 0) should return sprite 1")
		assert.Equal(t, 2, Mget(1, 1), "Mget(1, 1) should return sprite 2")
		assert.Equal(t, 3, Mget(2, 2), "Mget(2, 2) should return sprite 3")
		assert.Equal(t, 42, Mget(5, 7), "Mget(5, 7) should return sprite 42")

		// Float coordinates (should be converted to int)
		assert.Equal(t, 1, Mget(0.9, 0.9), "Mget(0.9, 0.9) should return sprite 1")
		assert.Equal(t, 2, Mget(1.5, 1.5), "Mget(1.5, 1.5) should return sprite 2")

		// Empty cell
		assert.Equal(t, 0, Mget(10, 10), "Mget for empty cell should return 0")
	})

	// Test Mget
	t.Run("Mget", func(t *testing.T) {
		assert.Equal(t, 1, Mget(0, 0), "Mget(0, 0) should return sprite 1")
		assert.Equal(t, 42, Mget(5, 7), "Mget(5, 7) should return sprite 42")
	})

	// Test Mset
	t.Run("Mset basic functionality", func(t *testing.T) {
		// Set a new sprite at an empty location
		Mset(10, 10, 99)
		assert.Equal(t, 99, Mget(10, 10), "After Mset(10, 10, 99), Mget should return 99")

		// Update an existing cell
		Mset(0, 0, 55)
		assert.Equal(t, 55, Mget(0, 0), "After Mset(0, 0, 55), Mget should return 55")

		// Test with float coordinates and sprite number
		Mset(11.7, 11.7, 77.9)
		assert.Equal(t, 77, Mget(11, 11), "After Mset(11.7, 11.7, 77.9), Mget(11, 11) should return 77")
	})

	// Test Mset alias
	t.Run("Mset alias", func(t *testing.T) {
		Mset(12, 12, 88)
		assert.Equal(t, 88, Mget(12, 12), "After Mset(12, 12, 88), Mget should return 88")
	})

	// Test map cell creation and update
	t.Run("Map cell creation and update", func(t *testing.T) {
		// Count initial cells
		initialCellCount := len(currentMap.Cells)

		// Set a sprite at a new position
		Mset(20, 20, 100)

		// Verify a new cell was added
		assert.Equal(t, initialCellCount+1, len(currentMap.Cells), "A new cell should be added")
		assert.Equal(t, 100, Mget(20, 20), "The new cell should have sprite 100")

		// Update the same cell
		Mset(20, 20, 101)

		// Verify the cell was updated, not added
		assert.Equal(t, initialCellCount+1, len(currentMap.Cells), "No new cell should be added when updating")
		assert.Equal(t, 101, Mget(20, 20), "The cell should be updated to sprite 101")
	})
}

// TestMgetWithMapFile tests Mget with a real map file
func TestMgetWithMapFile(t *testing.T) {
	// Skip this test if we're not in an environment with a map.json file
	if _, err := os.Stat("map.json"); os.IsNotExist(err) {
		t.Skip("Skipping test: map.json not found")
	}

	// Save original map
	originalMap := currentMap
	currentMap = nil // Force reload

	// Restore original map after test
	t.Cleanup(func() {
		currentMap = originalMap
	})

	// Create a temporary map.json for testing
	tempMap := MapData{
		Version:     "1.0",
		Description: "Temp Test Map",
		Width:       16,
		Height:      16,
		Name:        "temp_map",
		Cells: []mapCell{
			{X: 3, Y: 4, Sprite: 123},
			{X: 5, Y: 6, Sprite: 456},
		},
	}

	// Save the map to a temporary file
	tempMapData, err := json.Marshal(tempMap)
	if err != nil {
		t.Fatalf("Failed to marshal temp map: %v", err)
	}

	// Save the original map.json if it exists
	var originalMapData []byte
	if _, err := os.Stat("map.json"); err == nil {
		originalMapData, err = os.ReadFile("map.json")
		if err != nil {
			t.Fatalf("Failed to read original map.json: %v", err)
		}
	}

	// Write the temp map
	err = os.WriteFile("map.json", tempMapData, 0644)
	if err != nil {
		t.Fatalf("Failed to write temp map.json: %v", err)
	}

	// Restore the original map.json after the test
	t.Cleanup(func() {
		if len(originalMapData) > 0 {
			err := os.WriteFile("map.json", originalMapData, 0644)
			if err != nil {
				t.Logf("Warning: Failed to restore original map.json: %v", err)
			}
		} else {
			// If there was no original map.json, remove the temp one
			if err := os.Remove("map.json"); err != nil {
				t.Logf("Warning: Failed to remove temporary map.json: %v", err)
			}
		}
	})

	// Test Mget with the file-based map
	assert.Equal(t, 123, Mget(3, 4), "Mget(3, 4) should return sprite 123 from file")
	assert.Equal(t, 456, Mget(5, 6), "Mget(5, 6) should return sprite 456 from file")
	assert.Equal(t, 0, Mget(9, 9), "Mget(9, 9) should return 0 for empty cell")

	// Test Mset with the file-based map
	Mset(9, 9, 789)
	assert.Equal(t, 789, Mget(9, 9), "After Mset(9, 9, 789), Mget should return 789")
}
