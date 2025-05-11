// Package main basic sprite editor
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"time"

	p8 "github.com/drpaneas/pigo8"
)

type myGame struct {
	currentColor  int   // Current selected color from palette
	currentSprite int   // Current selected sprite from spritesheet (0-255)
	hoverX        int   // X coordinate of the pixel being hovered over (-1 if none)
	hoverY        int   // Y coordinate of the pixel being hovered over (-1 if none)
	gridSize      int   // Size of the working grid (1=8x8, 2=16x16, 4=32x32, 8=64x64)
	lastWheelTime int64 // Last time the mouse wheel was scrolled (for debouncing)
	mapMode       bool  // Whether we are in map mode

	// Map editor state
	mapCameraX    int            // Camera X position in the map (in sprites)
	mapCameraY    int            // Camera Y position in the map (in sprites)
	mapData       [320][320]int  // The map data - stores sprite indices

	// Popup notification
	showSavePopup bool   // Whether to show the save popup
	popupTimer    int    // Timer for the popup (in frames)
	popupMessage  string // Message to show in the popup
}

type MapData struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Name        string    `json:"name"`
	Cells       []mapCell `json:"cells"`
}

type mapCell struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Sprite int `json:"sprite"`
}

func (m *myGame) Init() {
	initSquareColors()

	// Initialize sprite flags to false
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			for flag := range 8 {
				spriteFlags[row][col][flag] = false
			}
		}
	}

	// Initialize spritesheet (will also load from file if available)
	initSpritesheet()

	// Try to load map data from map.json
	if err := m.loadMapData(); err != nil {
		fmt.Println("No map.json found, starting with empty map")
	}

	m.currentColor = 8  // Default to color 8 (usually red in PICO-8 palette)
	m.currentSprite = 1 // Default to first non-transparent sprite (sprite 0 is reserved)
	m.hoverX = -1       // No hover initially
	m.hoverY = -1       // No hover initially
	m.gridSize = 1      // Start with 8x8 grid (1 sprite)

	// Ensure grid size is never less than 1
	if m.gridSize < 1 {
		m.gridSize = 1
	}
	m.lastWheelTime = 0 // Initialize wheel time

	// Initialize the drawing canvas with the default sprite (1)
	updateDrawingCanvas(m)
}

// Define the sprite structure to match PIGO8's format
type spriteData struct {
	ID     int          `json:"id"`
	X      int          `json:"x"`
	Y      int          `json:"y"`
	Width  int          `json:"width"`
	Height int          `json:"height"`
	Used   bool         `json:"used"`
	Flags  p8.FlagsData `json:"flags"`
	Pixels [][]int      `json:"pixels"`
}

// Define the spritesheet structure
type spriteSheetData struct {
	// Custom spritesheet dimensions
	SpriteSheetColumns int          `json:"SpriteSheetColumns"`
	SpriteSheetRows    int          `json:"SpriteSheetRows"`
	SpriteSheetWidth   int          `json:"SpriteSheetWidth"`
	SpriteSheetHeight  int          `json:"SpriteSheetHeight"`
	Sprites            []spriteData `json:"sprites"`
}

// loadSpritesheet loads the spritesheet from spritesheet.json if it exists
func loadSpritesheet() {
	// Check if spritesheet.json exists
	data, err := os.ReadFile("spritesheet.json")
	if err != nil {
		// File doesn't exist or can't be read, just use the default empty spritesheet
		fmt.Println("No spritesheet.json found, using empty spritesheet")
		return
	}

	// Parse the JSON data
	var sheet spriteSheetData
	err = json.Unmarshal(data, &sheet)
	if err != nil {
		fmt.Println("Error parsing spritesheet.json:", err)
		return
	}

	// Check if the spritesheet has custom dimensions
	if sheet.SpriteSheetColumns > 0 && sheet.SpriteSheetRows > 0 {
		// Update the global variables with the dimensions from the JSON file
		spriteSheetCols = sheet.SpriteSheetColumns
		spriteSheetRows = sheet.SpriteSheetRows
		fmt.Printf("Loaded custom spritesheet dimensions: %dx%d sprites (%dx%d pixels)\n",
			spriteSheetCols, spriteSheetRows, sheet.SpriteSheetWidth, sheet.SpriteSheetHeight)
	}

	// Load the sprites into the spritesheet
	for _, sprite := range sheet.Sprites {
		// Calculate row and column from sprite ID
		row := sprite.ID / spriteSheetCols
		col := sprite.ID % spriteSheetCols

		// Make sure the sprite is within bounds
		if row >= 0 && row < spriteSheetRows && col >= 0 && col < spriteSheetCols {
			// Load pixel data
			for r := range 8 {
				for c := range 8 {
					// Make sure we have pixel data for this position
					if r < len(sprite.Pixels) && c < len(sprite.Pixels[r]) {
						spritesheet[row][col][r][c] = sprite.Pixels[r][c]
					}
				}
			}

			// Load flag data
			for i := range 8 {
				if i < len(sprite.Flags.Individual) {
					// Get the flag value from the individual array
					flagValue := sprite.Flags.Individual[i]
					// Store it in the spriteFlags array
					spriteFlags[row][col][i] = flagValue
				}
			}
		}
	}

	fmt.Println("Loaded spritesheet from spritesheet.json")
}

// saveSpritesheet saves the current spritesheet to a JSON file
func saveSpritesheet(g *myGame) error {
	// Create the spritesheet structure following the PIGO8 format
	var sheet spriteSheetData

	// Set the spritesheet dimensions
	sheet.SpriteSheetColumns = spriteSheetCols
	sheet.SpriteSheetRows = spriteSheetRows
	sheet.SpriteSheetWidth = spriteSheetCols * 8  // Each sprite is 8x8 pixels
	sheet.SpriteSheetHeight = spriteSheetRows * 8 // Each sprite is 8x8 pixels

	// Initialize the sprites slice with the correct capacity
	sprites := make([]spriteData, spriteSheetRows*spriteSheetCols)

	// Fill in the data for each sprite
	for row := 0; row < spriteSheetRows; row++ {
		for col := 0; col < spriteSheetCols; col++ {
			// Calculate sprite index
			spriteIndex := row*spriteSheetCols + col

			// Create a new sprite
			sprite := spriteData{
				ID:     spriteIndex,
				X:      col * 8,
				Y:      row * 8,
				Width:  8,
				Height: 8,
				Used:   true, // Mark all sprites as used
				Flags: p8.FlagsData{
					Bitfield:   0, // Will be calculated below
					Individual: make([]bool, 8),
				},
				Pixels: make([][]int, 8),
			}

			// Fill in the flags
			bitfield := 0
			for i := 0; i < 8; i++ {
				flagValue := spriteFlags[row][col][i]
				sprite.Flags.Individual[i] = flagValue
				if flagValue {
					bitfield |= 1 << i
				}
			}
			sprite.Flags.Bitfield = bitfield

			// Fill in the pixel data
			for r := 0; r < 8; r++ {
				sprite.Pixels[r] = make([]int, 8)
				for c := 0; c < 8; c++ {
					sprite.Pixels[r][c] = spritesheet[row][col][r][c]
				}
			}

			// Add the sprite to the sprites slice
			sprites[spriteIndex] = sprite
		}
	}

	// Assign the sprites to the sheet
	sheet.Sprites = sprites

	// Marshal the sheet to JSON
	data, err := json.MarshalIndent(sheet, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshalling spritesheet: %w", err)
	}

	// Write the JSON to a file
	err = os.WriteFile("spritesheet.json", data, 0644)
	if err != nil {
		return fmt.Errorf("error writing spritesheet.json: %w", err)
	}

	// Trigger the save popup notification
	g.showSavePopup = true
	g.popupTimer = 60 // Show for about 1 second (60 frames at 60fps)

	return nil
}

func (m *myGame) Update() {
	// Check if 'X' button is pressed to switch to map mode
	if p8.Btnp(p8.X) {
		m.mapMode = !m.mapMode
	}

	// Handle map mode controls
	if m.mapMode {
		// Save map when 'O' is pressed
		if p8.Btnp(p8.O) {
			if err := m.saveMapData(); err != nil {
				fmt.Println("Error saving map:", err)
				m.showPopup("Error saving map!")
			} else {
				fmt.Println("Map saved to map.json")
				m.showPopup("Map saved!")
			}
		}

		// Move camera with arrow keys (full screen = 16 sprites = 128 pixels)
		visibleSpritesX := mapViewWidth / unit
		if p8.Btnp(p8.LEFT) && m.mapCameraX > 0 {
			m.mapCameraX -= visibleSpritesX // Move left by one screen
		}
		if p8.Btnp(p8.RIGHT) && m.mapCameraX < 320-visibleSpritesX {
			m.mapCameraX += visibleSpritesX // Move right by one screen
		}
		visibleSpritesY := mapViewHeight / unit
		if p8.Btnp(p8.UP) && m.mapCameraY > 0 {
			m.mapCameraY -= visibleSpritesY // Move up by one screen
		}
		if p8.Btnp(p8.DOWN) && m.mapCameraY < 320-visibleSpritesY {
			m.mapCameraY += visibleSpritesY // Move down by one screen
		}

		// Handle sprite placement
		mx, my := p8.Mouse()

		// Check if mouse is within map bounds
		if mx >= 10 && mx < 10+mapViewWidth && my >= 10 && my < 10+mapViewHeight {
			// Calculate map coordinates from mouse position
			mapX := m.mapCameraX + (mx - 10) / 8
			mapY := m.mapCameraY + (my - 10) / 8

			// Handle right click to erase (set to sprite 0)
			if p8.Btn(p8.MouseRight) {
				// Check if target position is within map bounds
				if mapX >= 0 && mapX < 320 && mapY >= 0 && mapY < 320 {
					p8.Mset(mapX, mapY, 0) // Set to sprite 0 (empty/transparent)
					m.mapData[mapY][mapX] = 0 // Update internal map data
				}
				return
			}

			// Place sprite on left click
			// Place sprite(s) on left click
			if p8.Btn(p8.MouseLeft) {
				// Calculate grid dimensions based on current grid size
				gridWidth := 1
				gridHeight := 1
				switch m.gridSize {
				case 2: // 16x16
					gridWidth, gridHeight = 2, 2
				case 4: // 32x32
					gridWidth, gridHeight = 4, 4
				}

				// Calculate map coordinates from mouse position
				mapX := m.mapCameraX + (mx - 10) / 8
				mapY := m.mapCameraY + (my - 10) / 8

				// Get the base sprite (top-left of selection)
				baseSprite := m.currentSprite

				// Place all sprites in the grid if they fit within map bounds
				for dy := 0; dy < gridHeight; dy++ {
					for dx := 0; dx < gridWidth; dx++ {
						targetX := mapX + dx
						targetY := mapY + dy

						// Check if target position is within map bounds
						if targetX >= 0 && targetX < 320 && targetY >= 0 && targetY < 320 {
							// Calculate base sprite's position in the spritesheet
							baseRow := baseSprite / spriteSheetCols
							baseCol := baseSprite % spriteSheetCols

							// Calculate the correct sprite index based on position in grid
							spriteRow := baseRow + dy
							spriteCol := baseCol + dx
							spriteIndex := spriteRow*spriteSheetCols + spriteCol

							// Place the sprite if it's within spritesheet bounds
							if spriteRow < spriteSheetRows && spriteCol < spriteSheetCols {
								p8.Mset(targetX, targetY, spriteIndex)
							}
							m.mapData[targetY][targetX] = spriteIndex // Update internal map data
						}
					}
				}
			}
		}

		return // Skip rest of update when in map mode
	}

	// Check if 'O' button is pressed to save the spritesheet
	if p8.Btnp(p8.O) {
		err := saveSpritesheet(m)
		if err != nil {
			fmt.Println("Error saving spritesheet:", err)
		} else {
			fmt.Println("Spritesheet saved to spritesheet.json")
		}
	}

	// Update popup timer
	if m.showSavePopup {
		m.popupTimer--
		if m.popupTimer <= 0 {
			m.showSavePopup = false
		}
	}

	// Get mouse position
	mx, my := p8.Mouse()

	// Calculate draw grid boundaries for reference
	gridStartX := 10
	gridStartY := 10
	gridEndY := gridStartY + 8*12 - 2

	// Calculate spritesheet grid boundaries
	// Using global spritesheetStartX variable
	spritesheetStartY := 10
	// We don't need these variables in Update, but they're used in Draw
	// spritesheetEndX := spritesheetStartX + spriteSheetCols*spriteCellSize
	// spritesheetEndY := spritesheetStartY + spriteSheetRows*spriteCellSize

	// Calculate checkbox area boundaries
	checkboxStartY := gridEndY + 15
	checkboxSize := 8 // Smaller checkboxes

	// Check if mouse is over the checkboxes
	for i := range 8 {
		checkboxX := gridStartX + i*checkboxSize*3/2 // Space them out a bit
		checkboxY := checkboxStartY

		// Check if mouse is over this checkbox
		if mx >= checkboxX && mx < checkboxX+checkboxSize &&
			my >= checkboxY && my < checkboxY+checkboxSize {

			// If left mouse button is clicked, toggle the flag
			if p8.Btnp(p8.MouseLeft) {
				// Calculate the base sprite (top-left of the selection)
				baseSprite := m.currentSprite
				baseRow := baseSprite / spriteSheetCols
				baseCol := baseSprite % spriteSheetCols

				// Get the current state of the flag from the base sprite
				currentState := spriteFlags[baseRow][baseCol][i]
				// Toggle to the opposite state
				newState := !currentState

				// Apply the flag change to all selected sprites based on grid size
				// If gridSize is less than 1, default to 1 to ensure at least one sprite is affected
				effectiveGridSize := m.gridSize
				if effectiveGridSize < 1 {
					effectiveGridSize = 1
				}

				for r := 0; r < effectiveGridSize; r++ {
					for c := 0; c < effectiveGridSize; c++ {
						// Calculate the sprite position
						sprRow := baseRow + r
						sprCol := baseCol + c

						// Make sure we don't go out of bounds
						if sprRow >= 0 && sprRow < spriteSheetRows && sprCol >= 0 && sprCol < spriteSheetCols {
							// Set the flag to the new state
							spriteFlags[sprRow][sprCol][i] = newState
						}
					}
				}
			}
		}
	}

	// Calculate which square the mouse is over in the drawing grid
	// Calculate the cell size based on the grid size
	gridSize := 8 * m.gridSize  // Actual pixel dimensions (8, 16, 32, or 64)
	cellSize := 96 / gridSize   // 96 is the total space (8*12) divided by number of cells
	cellSize = max(1, cellSize) // Ensure cell size is at least 1 pixel

	// Calculate row and column based on cell size
	row := (my - gridStartY) / cellSize
	col := (mx - gridStartX) / cellSize

	// Check if mouse is over the drawing grid
	if row >= 0 && row < gridSize && col >= 0 && col < gridSize {
		// Calculate the base sprite (top-left of the selection)
		baseSprite := m.currentSprite
		baseRow := baseSprite / spriteSheetCols
		baseCol := baseSprite % spriteSheetCols

		// Calculate which sprite this pixel belongs to
		spriteRow := row / 8 // Which sprite row (0 for first sprite, 1 for second, etc.)
		spriteCol := col / 8 // Which sprite column

		// Calculate the position within that sprite (0-7)
		spritePixelRow := row % 8
		spritePixelCol := col % 8

		// Calculate the actual sprite coordinates in the spritesheet
		sprRow := baseRow + spriteRow
		sprCol := baseCol + spriteCol

		// Calculate absolute pixel coordinates for hover display
		m.hoverX = sprCol*8 + spritePixelCol
		m.hoverY = sprRow*8 + spritePixelRow

		// Make sure we don't go out of bounds
		if sprRow < spriteSheetRows && sprCol < spriteSheetCols {
			// Handle mouse clicks for drawing
			if p8.Btn(p8.MouseLeft) { // Left mouse button
				// Update both the visible drawing grid and the selected sprite
				setSquareColor(row, col, m.currentColor)                                     // Set color in the visible grid
				spritesheet[sprRow][sprCol][spritePixelRow][spritePixelCol] = m.currentColor // Set color in the sprite
				// Update the sprite in PIGO8
				p8.Sset(sprCol*8+spritePixelCol, sprRow*8+spritePixelRow, m.currentColor)
				// Update any map tiles using this sprite
				spriteIndex := sprRow*spriteSheetCols + sprCol
				updateMapSprites(spriteIndex)
				// Update the drawing canvas to reflect changes
				updateDrawingCanvas(m)
			} else if p8.Btn(p8.MouseRight) { // Right mouse button
				// Update both the visible drawing grid and the selected sprite
				setSquareColor(row, col, 0)                                     // Reset color in the visible grid
				spritesheet[sprRow][sprCol][spritePixelRow][spritePixelCol] = 0 // Reset color in the sprite
				// Update the sprite in PIGO8
				p8.Sset(sprCol*8+spritePixelCol, sprRow*8+spritePixelRow, 0)
				// Update any map tiles using this sprite
				spriteIndex := sprRow*spriteSheetCols + sprCol
				updateMapSprites(spriteIndex)
				// Update the drawing canvas to reflect changes
				updateDrawingCanvas(m)
			}
		}
	} else {
		// Not hovering over the grid
		m.hoverX = -1
		m.hoverY = -1
	}

	// Calculate which sprite the mouse is over in the spritesheet
	// Use spriteCellSize (8) instead of the old value (12)
	sprRow := (my - spritesheetStartY) / spriteCellSize
	sprCol := (mx - spritesheetStartX) / spriteCellSize

	// Check if mouse is over the spritesheet grid
	if sprRow >= 0 && sprRow < spriteSheetRows && sprCol >= 0 && sprCol < spriteSheetCols {
		// If left mouse button is clicked, select this sprite
		if p8.Btnp(p8.MouseLeft) {
			// Calculate the sprite index
			spriteIndex := sprRow*spriteSheetCols + sprCol
			// Don't allow selecting sprite 0 (reserved as transparent)
			if spriteIndex > 0 {
				m.currentSprite = spriteIndex
			}

			// Update the entire drawing canvas to reflect the new sprite selection
			// This ensures all pixels (including transparent ones) are properly refreshed
			updateDrawingCanvas(m)
		}
	}

	// Check if mouse is over the palette (positioned below the checkboxes)
	paletteY := gridEndY + 40 // Position palette 40 pixels below the grid (below checkboxes)
	paletteRow := (my - paletteY) / 12
	paletteCol := (mx - gridStartX) / 12

	// Always use 8 columns for the palette (must match drawPalette function)
	colorsPerRow := 8
	totalPaletteSize := p8.GetPaletteSize()
	totalRows := (totalPaletteSize + colorsPerRow - 1) / colorsPerRow // Ceiling division

	// Check if mouse is over the palette area
	if paletteRow >= 0 && paletteRow < totalRows && paletteCol >= 0 && paletteCol < colorsPerRow {
		// Calculate the color index based on row and column
		colorIndex := paletteRow*colorsPerRow + paletteCol

		// Make sure the color index is valid
		if colorIndex < totalPaletteSize {
			// If left mouse button is clicked, select this color
			if p8.Btnp(p8.MouseLeft) {
				m.currentColor = colorIndex
			}
		}
	}

	// Handle mouse wheel scrolling to adjust grid size with debouncing
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	// Only process wheel events if enough time has passed (500ms delay)
	if currentTime-m.lastWheelTime > 500 {
		if p8.Btnp(p8.MouseWheelUp) {
			// Increase grid size: 1 (8x8) -> 2 (16x16) -> 4 (32x32)
			// Limit to 4 (32x32) as the maximum grid size
			if m.gridSize < 4 {
				// Ensure grid size is at least 1 before multiplying
				if m.gridSize < 1 {
					m.gridSize = 1
				}
				m.gridSize *= 2
				// Update the drawing canvas to show the selected sprites
				updateDrawingCanvas(m)
				// Update the last wheel time
				m.lastWheelTime = currentTime
			}
		} else if p8.Btnp(p8.MouseWheelDown) {
			// Decrease grid size: 4 (32x32) -> 2 (16x16) -> 1 (8x8)
			if m.gridSize > 1 {
				m.gridSize /= 2
				// Ensure grid size is never less than 1
				if m.gridSize < 1 {
					m.gridSize = 1
				}
				// Update the drawing canvas to show the selected sprites
				updateDrawingCanvas(m)
				// Update the last wheel time
				m.lastWheelTime = currentTime
			}
		}
	}
}

func (g *myGame) Draw() {
	p8.ClsRGBA(color.RGBA{R: 25, G: 25, B: 25, A: 255})

	// If we are in map mode, draw the spritemap 
	if g.mapMode {
		// Define viewport boundaries
		const viewportX = 10  // Left margin of sprite editor viewport
		const viewportY = 10  // Top margin of sprite editor viewport

		// Draw the visible portion of the map
		visibleSpritesX := mapViewWidth / unit  // Number of sprites visible horizontally
		visibleSpritesY := mapViewHeight / unit // Number of sprites visible vertically
		for y := 0; y < visibleSpritesY; y++ {
			for x := 0; x < visibleSpritesX; x++ {
				// Get sprite at current map position
				sprite := p8.Mget(g.mapCameraX+x, g.mapCameraY+y)
				// Draw the sprite at viewport position
				p8.Spr(sprite, float64(viewportX+x*8), float64(viewportY+y*8))
			}
		}

		// Get mouse position for hover highlight
		mx, my := p8.Mouse()

		// Calculate grid dimensions based on current grid size
		gridWidth := 1
		gridHeight := 1
		switch g.gridSize {
		case 2: // 16x16
			gridWidth, gridHeight = 2, 2
		case 4: // 32x32
			gridWidth, gridHeight = 4, 4
		}

		// Draw hover highlight for multi-sprite placement
		hoverX := (mx - viewportX) / 8
		hoverY := (my - viewportY) / 8
		if hoverX >= 0 && hoverX < visibleSpritesX && hoverY >= 0 && hoverY < visibleSpritesY {
			// Draw hover highlight for each sprite in the grid
			for dy := 0; dy < gridHeight; dy++ {
				for dx := 0; dx < gridWidth; dx++ {
					if int(hoverX)+dx < visibleSpritesX && int(hoverY)+dy < visibleSpritesY {
						// Draw hover highlight
						p8.Rect(float64(viewportX+(int(hoverX)+dx)*8), float64(viewportY+(int(hoverY)+dy)*8),
							float64(viewportX+(int(hoverX)+dx+1)*8-1), float64(viewportY+(int(hoverY)+dy+1)*8-1), 7)
					}
				}
			}
		}

		// Draw a border around the current screen (128x128 pixels)
		p8.Rect(float64(viewportX), float64(viewportY), float64(viewportX+mapViewWidth), float64(viewportY+mapViewHeight), 7) // White border

		// Reset camera for UI elements
		p8.Camera()

		// Display map information
		p8.Color(7) // White text
		// Show current screen coordinates
		screenX := g.mapCameraX / (mapViewWidth / unit)
		screenY := g.mapCameraY / (mapViewHeight / unit)
		// Calculate text positions relative to the viewport
		textY := viewportY + mapViewHeight + 10 // 10 pixels below the viewport
		p8.Print(fmt.Sprintf("Screen: %d,%d", screenX, screenY), viewportX, textY, 7)

		// Show mouse coordinates in map space
		// Use existing mx, my from earlier in the function
		mapX := g.mapCameraX + (mx - viewportX) / 8
		mapY := g.mapCameraY + (my - viewportY) / 8
		// Only show coordinates if mouse is within map bounds
		if mx >= viewportX && mx < viewportX+mapViewWidth && my >= viewportY && my < viewportY+mapViewHeight {
			p8.Print(fmt.Sprintf("Map: %d,%d", mapX, mapY), viewportX + 90, textY, 7)
			// Show sprite at current position
			if mapX >= 0 && mapX < 128 && mapY >= 0 && mapY < 128 {
				sprite := p8.Mget(mapX, mapY)
				p8.Print(fmt.Sprintf("Sprite: %d", sprite), viewportX + 180, textY, 7)
			}
		}
		return
	}

	// Ensure drawing canvas is up to date
	updateDrawingCanvas(g)

	// Calculate drawing grid boundaries
	gridStartX := 10
	gridStartY := 10
	gridEndX := gridStartX + 8*12 - 2 // 8 columns * 12 pixels - 2 for border alignment
	gridEndY := gridStartY + 8*12 - 2 // 8 rows * 12 pixels - 2 for border alignment

	// Calculate spritesheet grid boundaries
	// Using global spritesheetStartX variable
	spritesheetStartY := 10
	// Each sprite is exactly 8x8 pixels with no gap or border
	spritesheetEndX := spritesheetStartX + spriteSheetCols*spriteCellSize
	spritesheetEndY := spritesheetStartY + spriteSheetRows*spriteCellSize

	// Draw the drawing canvas based on the current grid size
	gridSize := 8 * g.gridSize // Actual pixel dimensions (8, 16, 32, or 64)

	// Calculate the cell size to fit the grid in the same space
	cellSize := 96 / gridSize   // 96 is the total space (8*12) divided by number of cells
	cellSize = max(1, cellSize) // Ensure cell size is at least 1 pixel

	// Draw the grid
	for row := range gridSize {
		for col := range gridSize {
			c := getSquareColor(row, col)  // Get the color of the square
			x := gridStartX + col*cellSize // col determines the x-coordinate
			y := gridStartY + row*cellSize // row determines the y-coordinate

			// Fill the square with its color
			if cellSize > 1 {
				p8.Rectfill(x, y, x+cellSize-1, y+cellSize-1, c)
			} else {
				// For very small cells, just set a pixel
				p8.Pset(x, y, c)
			}
		}
	}

	// Display pixel coordinates if hovering over the grid
	if g.hoverX >= 0 && g.hoverY >= 0 {
		coordText := "pixel: (" + strconv.Itoa(g.hoverX) + "," + strconv.Itoa(g.hoverY) + ")"
		p8.Print(coordText, gridStartX, gridStartY-10, 7,)
	}

	// Draw perimeter around the entire drawing grid
	p8.Rect(gridStartX-1, gridStartY-1, gridEndX+1, gridEndY+1, 7) // White perimeter

	// Draw the spritesheet grid with just the sprites, no borders
	for sprRow := range spriteSheetRows {
		for sprCol := range spriteSheetCols {
			// Calculate the position for this sprite in the spritesheet
			sprX := spritesheetStartX + sprCol*spriteCellSize
			sprY := spritesheetStartY + sprRow*spriteCellSize

			// Draw the sprite content
			for r := range 8 {
				for c := range 8 {
					// Get the color from the sprite
					color := spritesheet[sprRow][sprCol][r][c]

					// Calculate exact pixel position
					pixelX := sprX + c
					pixelY := sprY + r

					// Draw the pixel (black for transparent)
					if color == 0 {
						p8.Pset(pixelX, pixelY, 0)
					} else {
						p8.Pset(pixelX, pixelY, color)
					}
				}
			}
		}
	}

	// Calculate the base sprite (top-left of the selection)
	baseSprite := g.currentSprite
	baseRow := baseSprite / spriteSheetCols
	baseCol := baseSprite % spriteSheetCols

	// Only draw selection borders if we have sprites selected
	if g.gridSize >= 1 {
		// Calculate the top-left and bottom-right corners of the entire selection
		topLeftX := spritesheetStartX + baseCol*spriteCellSize - 1
		topLeftY := spritesheetStartY + baseRow*spriteCellSize - 1
		bottomRightX := topLeftX + g.gridSize*spriteCellSize + 1
		bottomRightY := topLeftY + g.gridSize*spriteCellSize + 1

		// Make sure we don't go out of bounds
		if bottomRightX > spritesheetStartX+spriteSheetCols*spriteCellSize {
			bottomRightX = spritesheetStartX + spriteSheetCols*spriteCellSize
		}
		if bottomRightY > spritesheetStartY+spriteSheetRows*spriteCellSize {
			bottomRightY = spritesheetStartY + spriteSheetRows*spriteCellSize
		}

		// Draw a single white border around the entire selection
		p8.Rect(topLeftX, topLeftY, bottomRightX, bottomRightY, 7) // White border
	}

	// Draw perimeter around the entire spritesheet grid
	p8.Rect(spritesheetStartX-1, spritesheetStartY-1, spritesheetEndX+1, spritesheetEndY+1, 7)

	// Draw a label for the spritesheet and show selected sprite number and grid size
	gridSizeText := "8x8"
	switch g.gridSize {
	case 2:
		gridSizeText = "16x16"
	case 4:
		gridSizeText = "32x32"
	}
	p8.Print("spritesheet - sprite: "+strconv.Itoa(g.currentSprite)+" - grid: "+gridSizeText,
		spritesheetStartX, spritesheetEndY+4, 7,)

	// Draw the checkboxes for flags between the grid and the palette
	g.drawCheckboxes(gridStartX, gridEndY+15)

	// Draw the color palette below the checkboxes
	g.drawPalette(gridStartX, gridEndY+40)

	// Draw the save popup if active
	if g.showSavePopup {
		g.drawSavePopup()
	}
}

var (
	width           = 52 // Increased to accommodate the larger spritesheet and more space
	height          = 27 // Increased to accommodate the taller spritesheet
	mapViewWidth    = 128 // Default map viewport width in pixels (16 sprites)
	mapViewHeight   = 128 // Default map viewport height in pixels (16 sprites)
	unit            = 8
	spriteCellSize  = 8  // Size of each sprite cell in the spritesheet
	spriteSheetCols = 32 // Number of columns in the spritesheet
	spriteSheetRows = 24 // Number of rows in the spritesheet

	// Position of the spritesheet grid
	spritesheetStartX = 120 // Position spritesheet (adjusted 10px to the left)
)

var squareColors [64][64]int      // Up to 64x64 grid to store square colors
var spritesheet [24][32][8][8]int // 24x32 grid of 8x8 sprites
var spriteFlags [24][32][8]bool   // Flags for each sprite [row][col][flag0-7]

func initSquareColors() {
	for row := range 64 {
		for col := range 64 {
			squareColors[row][col] = 0 // Default color
		}
	}
}

func initSpritesheet() {
	// Try to load the spritesheet from a file
	loadSpritesheet()

	// Initialize the spritesheet with default values if needed
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			for r := range 8 {
				for c := range 8 {
					// Initialize with transparent color (0)
					spritesheet[row][col][r][c] = 0
				}
			}
		}
	}
	
	// Initialize the spritesheet in PIGO8 to ensure sprites exist
	initPico8Spritesheet()
}

func initPico8Spritesheet() {
	// Create a temporary spritesheet.json file that PIGO8 can load
	createTempSpritesheet()

	// Now initialize all sprites with our data
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			// Set each pixel in the sprite
			for r := 0; r < 8; r++ {
				for c := 0; c < 8; c++ {
					// Calculate the absolute pixel position
					px := col*8 + c
					py := row*8 + r
					// Set the pixel color in PIGO8
					p8.Sset(px, py, spritesheet[row][col][r][c])
				}
			}
		}
	}
}

// createTempSpritesheet creates a temporary spritesheet.json file
// that PIGO8 can load to initialize its sprite system
func createTempSpritesheet() {
	// Create a basic spritesheet structure
	var sheet spriteSheetData

	// Set dimensions
	sheet.SpriteSheetColumns = spriteSheetCols
	sheet.SpriteSheetRows = spriteSheetRows
	sheet.SpriteSheetWidth = spriteSheetCols * 8
	sheet.SpriteSheetHeight = spriteSheetRows * 8

	// Create sprites array
	sprites := make([]spriteData, spriteSheetRows*spriteSheetCols)

	// Fill in basic sprite data
	for row := 0; row < spriteSheetRows; row++ {
		for col := 0; col < spriteSheetCols; col++ {
			spriteIndex := row*spriteSheetCols + col
			
			// Create a sprite with basic data
			sprite := spriteData{
				ID:     spriteIndex,
				X:      col * 8,
				Y:      row * 8,
				Width:  8,
				Height: 8,
				Used:   true,
				Flags:  p8.FlagsData{
					Bitfield:   0,
					Individual: make([]bool, 8),
				},
				Pixels: make([][]int, 8),
			}
			
			// Initialize pixel data
			for r := 0; r < 8; r++ {
				sprite.Pixels[r] = make([]int, 8)
				for c := 0; c < 8; c++ {
					sprite.Pixels[r][c] = spritesheet[row][col][r][c]
				}
			}
			
			// Add to sprites array
			sprites[spriteIndex] = sprite
		}
	}
	
	// Set sprites in sheet
	sheet.Sprites = sprites
	
	// Convert to JSON
	jsonData, err := json.MarshalIndent(sheet, "", "  ")
	if err != nil {
		fmt.Println("Error creating temporary spritesheet JSON:", err)
		return
	}
	
	// Write to file
	err = os.WriteFile("spritesheet.json", jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing temporary spritesheet file:", err)
		return
	}
	
	// The spritesheet.json file will be loaded automatically the next time
	// a sprite-related function like Spr() or Sset() is called
}

func updateMapSprites(spriteIndex int) {
	// Scan through the entire map and update any instances of this sprite
	for y := 0; y < 128; y++ {
		for x := 0; x < 128; x++ {
			// Check if this map cell uses the modified sprite
			if p8.Mget(x, y) == spriteIndex {
				// Force a redraw of this sprite by setting it to itself
				p8.Mset(x, y, spriteIndex)
			}
		}
	}
}

func getSquareColor(row, col int) int {
	if row < 0 || row >= 64 || col < 0 || col >= 64 {
		return -1
	}
	return squareColors[row][col]
}

func setSquareColor(row, col, color int) {
	// Ensure the coordinates are within the grid
	if row < 0 || row >= 64 || col < 0 || col >= 64 {
		return
	}
	// Update the color in the squareColors array
	squareColors[row][col] = color
}

// drawCheckboxes draws the 8 checkboxes for sprite flags
func (g *myGame) drawCheckboxes(x, y int) {
	checkboxSize := 8 // Smaller checkboxes

	// Calculate the base sprite (top-left of the selection)
	baseSprite := g.currentSprite
	baseRow := baseSprite / spriteSheetCols
	baseCol := baseSprite % spriteSheetCols

	// Draw 8 checkboxes in a row
	for i := range 8 {
		checkboxX := x + i*checkboxSize*3/2 // Space them out a bit
		checkboxY := y

		// Draw checkbox outline
		p8.Rect(checkboxX, checkboxY, checkboxX+checkboxSize-1, checkboxY+checkboxSize-1, 7)

		// Check the flag state across all selected sprites
		allTrue := true
		allFalse := true

		// Loop through all selected sprites based on grid size
		// If gridSize is less than 1, default to 1 to ensure at least one sprite is checked
		effectiveGridSize := g.gridSize
		if effectiveGridSize < 1 {
			effectiveGridSize = 1
		}

		for r := 0; r < effectiveGridSize; r++ {
			for c := 0; c < effectiveGridSize; c++ {
				// Calculate the sprite position
				sprRow := baseRow + r
				sprCol := baseCol + c

				// Make sure we don't go out of bounds
				if sprRow >= 0 && sprRow < spriteSheetRows && sprCol >= 0 && sprCol < spriteSheetCols {
					// Check flag state
					if spriteFlags[sprRow][sprCol][i] {
						allFalse = false // At least one is true
					} else {
						allTrue = false // At least one is false
					}
				}
			}
		}

		// Fill the checkbox based on state
		if allTrue {
			// All sprites have this flag set - fill with solid color
			p8.Rectfill(checkboxX+2, checkboxY+2, checkboxX+checkboxSize-3, checkboxY+checkboxSize-3, 8)
		} else if !allFalse {
			// Mixed state - some sprites have this flag set, others don't - show a pattern
			p8.Rectfill(checkboxX+2, checkboxY+2, checkboxX+checkboxSize-3, checkboxY+checkboxSize-3, 6) // Use a different color
			// Draw a pattern to indicate mixed state
			p8.Line(checkboxX+2, checkboxY+2, checkboxX+checkboxSize-3, checkboxY+checkboxSize-3, 8)
			p8.Line(checkboxX+checkboxSize-3, checkboxY+2, checkboxX+2, checkboxY+checkboxSize-3, 8)
		}

		// Draw flag number
		p8.Print(strconv.Itoa(i), checkboxX+1, checkboxY+checkboxSize+2, 7,)
	}

	// Draw label
	p8.Print("flags", x, y-10, 7,)
}

// drawPalette draws the color palette below the grid
func (g *myGame) drawPalette(x, y int) {
	// Get the total palette size
	totalColors := p8.GetPaletteSize()

	// Always use 8 columns for the palette
	colorsPerRow := 8

	// Draw each color in the palette
	for i := range totalColors {
		// Calculate the row and column for this color
		row := i / colorsPerRow
		col := i % colorsPerRow

		// Calculate the position
		px := x + col*12
		py := y + row*12

		// Draw the color square
		p8.Rectfill(px, py, px+10, py+10, i)

		// Highlight the currently selected color with a white border
		if i == g.currentColor {
			p8.Rect(px-1, py-1, px+11, py+11, 7) // White highlight border
		}
	}
}

// updateDrawingCanvas updates the drawing canvas to show the selected sprites based on current grid size
func updateDrawingCanvas(g *myGame) {
	// Calculate the base sprite (top-left of the selection)
	baseSprite := g.currentSprite
	baseRow := baseSprite / spriteSheetCols
	baseCol := baseSprite % spriteSheetCols

	// Clear the drawing canvas
	for row := range 64 {
		for col := range 64 {
			squareColors[row][col] = 0
		}
	}

	// Calculate the number of sprites to show based on grid size
	spritesPerRow := 1
	spritesPerCol := 1
	switch g.gridSize {
	case 2: // 16x16
		spritesPerRow = 2
		spritesPerCol = 2
	case 4: // 32x32
		spritesPerRow = 4
		spritesPerCol = 4
	}

	// Update the drawing canvas for each selected sprite
	for row := 0; row < spritesPerCol; row++ {
		for col := 0; col < spritesPerRow; col++ {
			// Calculate the sprite index for this position
			spriteRow := baseRow + row
			spriteCol := baseCol + col
			spriteIndex := spriteRow*spriteSheetCols + spriteCol

			// Skip if we're outside the spritesheet bounds
			if spriteIndex >= spriteSheetCols*spriteSheetRows {
				continue
			}

			// Copy the sprite's pixels to the drawing canvas
			for pixelRow := 0; pixelRow < 8; pixelRow++ {
				for pixelCol := 0; pixelCol < 8; pixelCol++ {
					// Calculate the position in the drawing canvas
					canvasRow := row*8 + pixelRow
					canvasCol := col*8 + pixelCol

					// Get the color from the sprite
					color := p8.Sget(spriteCol*8+pixelCol, spriteRow*8+pixelRow)

					// Update the drawing canvas
					squareColors[canvasRow][canvasCol] = color
				}
			}
		}
	}
}

// drawSavePopup draws a popup notification
func (m *myGame) drawSavePopup() {
	if !m.showSavePopup {
		return
	}

	// Calculate popup dimensions and position
	popupWidth := 120
	popupHeight := 30
	popupX := (width*8 - popupWidth) / 2
	popupY := (height*8 - popupHeight) / 2

	// Draw popup background
	p8.Rectfill(popupX-2, popupY-2, popupX+popupWidth+2, popupY+popupHeight+2, 0) // Black border
	p8.Rectfill(popupX, popupY, popupX+popupWidth, popupY+popupHeight, 7)         // White background

	// Draw text
	textX := popupX + popupWidth/2
	textY := popupY + popupHeight/2
	p8.Print(m.popupMessage, textX-40, textY-3, 0) // Center the text
	// Draw close button
	closeX := popupX + popupWidth - 20
	closeY := popupY + 5
	p8.Rectfill(closeX, closeY, closeX+15, closeY+15, 8) // Red button
	p8.Print("X", closeX+5, closeY+5, 7,)                 // White X

	// Check for mouse click on close button
	mx, my := p8.Mouse()
	if p8.Btnp(p8.MouseLeft) {
		if mx >= closeX && mx <= closeX+15 && my >= closeY && my <= closeY+15 {
			m.showSavePopup = false
		}
	}
}

// showPopup displays a popup message for 120 frames (2 seconds)
func (m *myGame) showPopup(message string) {
	m.showSavePopup = true
	m.popupTimer = 120
	m.popupMessage = message
}

// saveMapData saves the current map to map.json
func (m *myGame) saveMapData() error {
	// Create a map data structure that matches PIGO8's format
	mapData := MapData{
		Version:     "1.0",
		Description: "Map created with PIGO8 editor",
		Width:       320,
		Height:      320,
		Name:        "map",
		Cells:       []mapCell{},
	}

	// Convert our map data to PIGO8's format
	for y := 0; y < 320; y++ {
		for x := 0; x < 320; x++ {
			sprite := m.mapData[y][x]
			// Only save non-zero sprites to keep the file size smaller
			if sprite != 0 {
				mapData.Cells = append(mapData.Cells, mapCell{
					X:      x,
					Y:      y,
					Sprite: sprite,
				})
			}
		}
	}

	// Convert to JSON
	data, err := json.MarshalIndent(mapData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling map data: %w", err)
	}

	// Write to file
	if err := os.WriteFile("map.json", data, 0644); err != nil {
		return fmt.Errorf("error writing map.json: %w", err)
	}

	return nil
}

// loadMapData loads the map from map.json if it exists
func (m *myGame) loadMapData() error {
	// Try to read the map file
	data, err := os.ReadFile("map.json")
	if err != nil {
		return fmt.Errorf("error reading map.json: %w", err)
	}

	// Parse the JSON data
	var mapData MapData
	if err := json.Unmarshal(data, &mapData); err != nil {
		return fmt.Errorf("error parsing map.json: %w", err)
	}

	// Initialize map with zeros
	for y := range m.mapData {
		for x := range m.mapData[y] {
			m.mapData[y][x] = 0
		}
	}

	// Load the cells into our map data
	for _, cell := range mapData.Cells {
		// Make sure coordinates are within bounds
		if cell.X >= 0 && cell.X < 320 && cell.Y >= 0 && cell.Y < 320 {
			m.mapData[cell.Y][cell.X] = cell.Sprite
			// Also update the PIGO8 map
			p8.Mset(cell.X, cell.Y, cell.Sprite)
		}
	}

	fmt.Printf("Loaded map from map.json: %dx%d with %d cells\n", mapData.Width, mapData.Height, len(mapData.Cells))
	return nil
}

func main() {
	// Parse command line flags
	widthFlag := flag.Int("w", mapViewWidth, "map viewport width in pixels")
	heightFlag := flag.Int("h", mapViewHeight, "map viewport height in pixels")
	flag.Parse()

	// Update map viewport dimensions
	mapViewWidth = *widthFlag
	mapViewHeight = *heightFlag

	// Ensure dimensions are multiples of 8 (sprite size)
	mapViewWidth = (mapViewWidth / unit) * unit
	mapViewHeight = (mapViewHeight / unit) * unit

	settings := p8.NewSettings()
	settings.ScreenWidth = width * unit
	settings.ScreenHeight = height * unit
	settings.ScaleFactor = 5
	p8.InsertGame(&myGame{})
	p8.PlayGameWith(settings)
}
