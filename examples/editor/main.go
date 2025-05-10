// Package main basic sprite editor
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"encoding/json"
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
	showSavePopup bool // Whether to show the save popup
	popupTimer    int  // Timer for the popup (in frames)
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

	// Initialize the first sprite with the drawing canvas
	for row := range 8 {
		for col := range 8 {
			squareColors[row][col] = spritesheet[0][0][row][col] // Load from first sprite
		}
	}
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
		// Move camera with arrow keys (full screen = 16 sprites = 128 pixels)
		if p8.Btnp(p8.LEFT) && m.mapCameraX > 0 {
			m.mapCameraX -= 16 // Move left by one screen
		}
		if p8.Btnp(p8.RIGHT) && m.mapCameraX < 320-16 {
			m.mapCameraX += 16 // Move right by one screen
		}
		if p8.Btnp(p8.UP) && m.mapCameraY > 0 {
			m.mapCameraY -= 16 // Move up by one screen
		}
		if p8.Btnp(p8.DOWN) && m.mapCameraY < 320-16 {
			m.mapCameraY += 16 // Move down by one screen
		}

		// Handle sprite placement
		mx, my := p8.Mouse()

		// Place sprite(s) on left click if within bounds
		if p8.Btn(p8.MouseLeft) && mx >= 10 && mx < 138 && my >= 10 && my < 138 {
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
					if targetX >= 0 && targetX < 128 && targetY >= 0 && targetY < 128 {
						// Calculate the correct sprite index based on position in grid
						spriteOffset := dy*32 + dx // 32 is the spritesheet width
						p8.Mset(targetX, targetY, baseSprite + spriteOffset)
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
			} else if p8.Btn(p8.MouseRight) { // Right mouse button
				// Update both the visible drawing grid and the selected sprite
				setSquareColor(row, col, 0)                                     // Reset color in the visible grid
				spritesheet[sprRow][sprCol][spritePixelRow][spritePixelCol] = 0 // Reset color in the sprite
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
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
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
		if hoverX >= 0 && hoverX < 16 && hoverY >= 0 && hoverY < 16 {
			// Draw hover highlight for each sprite in the grid
			for dy := 0; dy < gridHeight; dy++ {
				for dx := 0; dx < gridWidth; dx++ {
					if int(hoverX)+dx < 16 && int(hoverY)+dy < 16 {
						// Draw hover highlight
						p8.Rect(float64(viewportX+(int(hoverX)+dx)*8), float64(viewportY+(int(hoverY)+dy)*8),
							float64(viewportX+(int(hoverX)+dx+1)*8-1), float64(viewportY+(int(hoverY)+dy+1)*8-1), 7)
					}
				}
			}
		}

		// Draw a border around the current screen (128x128 pixels)
		p8.Rect(float64(viewportX), float64(viewportY), float64(viewportX+128), float64(viewportY+128), 7) // White border

		// Reset camera for UI elements
		p8.Camera()
		return
	}

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
		p8.Print(coordText, gridStartX, gridStartY-10, 7)
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
		spritesheetStartX, spritesheetEndY+4, 7)

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
	// Initialize all sprites with transparent color (0)
	for sprRow := range spriteSheetRows {
		for sprCol := range spriteSheetCols {
			for row := range 8 {
				for col := range 8 {
					spritesheet[sprRow][sprCol][row][col] = 0
				}
			}
		}
	}

	// Try to load spritesheet.json if it exists
	loadSpritesheet()
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
		p8.Print(strconv.Itoa(i), checkboxX+1, checkboxY+checkboxSize+2, 7)
	}

	// Draw label
	p8.Print("flags", x, y-10, 7)
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

// drawSavePopup draws a popup notification when the spritesheet is saved
func (g *myGame) drawSavePopup() {
	// Calculate the center of the screen
	centerX := width * unit / 2
	centerY := height * unit / 2

	// Calculate popup dimensions
	popupWidth := 80
	popupHeight := 30
	popupX := centerX - popupWidth/2
	popupY := centerY - popupHeight/2

	// Calculate fade effect based on timer (fade in and out)
	fadeAlpha := 1.0
	if g.popupTimer < 15 {
		// Fade out during the last 15 frames
		fadeAlpha = float64(g.popupTimer) / 15.0
	} else if g.popupTimer > 45 {
		// Fade in during the first 15 frames
		fadeAlpha = float64(60-g.popupTimer) / 15.0
	}

	// Use fadeAlpha to determine popup visibility
	// For PICO-8 style, we'll adjust the size of the popup based on the fade value
	popupScaledWidth := int(float64(popupWidth) * fadeAlpha)
	popupScaledHeight := int(float64(popupHeight) * fadeAlpha)

	// Recalculate popup position to keep it centered
	popupX = centerX - popupScaledWidth/2
	popupY = centerY - popupScaledHeight/2

	// Draw popup background with size based on fade value
	p8.Rectfill(popupX, popupY, popupX+popupScaledWidth, popupY+popupScaledHeight, 0) // Black background
	p8.Rect(popupX, popupY, popupX+popupScaledWidth, popupY+popupScaledHeight, 7)     // White border

	// Draw text
	message := "SAVED!"
	// Calculate text position to center it in the scaled popup
	textX := popupX + popupScaledWidth/2 - len(message)*2
	textY := popupY + popupScaledHeight/2 - 2

	// Scale text color based on fade alpha for a better effect
	textColor := 7   // Default white text
	shadowColor := 1 // Default shadow color

	// Draw text with a shadow for better visibility
	p8.Print(message, textX+1, textY+1, shadowColor) // Shadow
	p8.Print(message, textX, textY, textColor)       // White text
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

	// Copy the selected sprites to the drawing canvas
	// If gridSize is less than 1, default to 1 to ensure at least one sprite is copied
	effectiveGridSize := g.gridSize
	if effectiveGridSize < 1 {
		effectiveGridSize = 1
	}

	for gridRow := 0; gridRow < effectiveGridSize; gridRow++ {
		for gridCol := 0; gridCol < effectiveGridSize; gridCol++ {
			// Calculate the sprite position
			sprRow := baseRow + gridRow
			sprCol := baseCol + gridCol

			// Make sure we don't go out of bounds
			if sprRow >= 0 && sprRow < spriteSheetRows && sprCol >= 0 && sprCol < spriteSheetCols {
				// Copy this sprite's pixels to the appropriate section of the drawing canvas
				for r := range 8 {
					for c := range 8 {
						// Calculate the position in the drawing canvas
						drawRow := gridRow*8 + r
						drawCol := gridCol*8 + c

						// Copy the pixel
						squareColors[drawRow][drawCol] = spritesheet[sprRow][sprCol][r][c]
					}
				}
			}
		}
	}
}

func main() {
	settings := p8.NewSettings()
	settings.ScreenWidth = width * unit
	settings.ScreenHeight = height * unit
	settings.ScaleFactor = 5
	p8.InsertGame(&myGame{})
	p8.PlayGameWith(settings)
}
