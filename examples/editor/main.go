// Package main basic sprite editor
//
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
package main

import (
	"image/color"
	"strconv"
	"time"

	p8 "github.com/drpaneas/pigo8"
)

type myGame struct {
	currentColor  int             // Current selected color from palette
	currentSprite int             // Current selected sprite from spritesheet (0-255)
	spriteFlags   [24][32][8]bool // Flags for each sprite [row][col][flag0-7]
	hoverX        int             // X coordinate of the pixel being hovered over (-1 if none)
	hoverY        int             // Y coordinate of the pixel being hovered over (-1 if none)
	gridSize      int             // Size of the working grid (1=8x8, 2=16x16, 4=32x32, 8=64x64)
	lastWheelTime int64           // Last time the mouse wheel was scrolled (for debouncing)
}

func (m *myGame) Init() {
	initSquareColors()
	initSpritesheet()
	m.currentColor = 8  // Default to color 8 (usually red in PICO-8 palette)
	m.currentSprite = 0 // Default to first sprite
	m.hoverX = -1       // No hover initially
	m.hoverY = -1       // No hover initially
	m.gridSize = 1      // Start with 8x8 grid (1 sprite)
	m.lastWheelTime = 0 // Initialize wheel time

	// Initialize sprite flags to false
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			for flag := range 8 {
				m.spriteFlags[row][col][flag] = false
			}
		}
	}

	// Initialize the first sprite with the drawing canvas
	for row := range 8 {
		for col := range 8 {
			squareColors[row][col] = 0 // Start with empty canvas
		}
	}
}

func (m *myGame) Update() {
	// Get mouse position
	mx, my := p8.Mouse()

	// Calculate draw grid boundaries for reference
	gridStartX := 10
	gridStartY := 10
	gridEndY := gridStartY + 8*12 - 2

	// Calculate spritesheet grid boundaries
	spritesheetStartX := 110 // Position spritesheet closer to the left side
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
				currentState := m.spriteFlags[baseRow][baseCol][i]
				// Toggle to the opposite state
				newState := !currentState

				// Apply the flag change to all selected sprites based on grid size
				for r := 0; r < m.gridSize; r++ {
					for c := 0; c < m.gridSize; c++ {
						// Calculate the sprite position
						sprRow := baseRow + r
						sprCol := baseCol + c

						// Make sure we don't go out of bounds
						if sprRow < spriteSheetRows && sprCol < spriteSheetCols {
							// Set the flag to the new state
							m.spriteFlags[sprRow][sprCol][i] = newState
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
			m.currentSprite = spriteIndex

			// Load the selected sprite into the drawing grid
			for r := range 8 {
				for c := range 8 {
					squareColors[r][c] = spritesheet[sprRow][sprCol][r][c]
				}
			}

			// The flags will be automatically updated since we're using the currentSprite index
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

	// Calculate drawing grid boundaries
	gridStartX := 10
	gridStartY := 10
	gridEndX := gridStartX + 8*12 - 2 // 8 columns * 12 pixels - 2 for border alignment
	gridEndY := gridStartY + 8*12 - 2 // 8 rows * 12 pixels - 2 for border alignment

	// Calculate spritesheet grid boundaries
	spritesheetStartX := 110 // Position spritesheet closer to the left side
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
}

var (
	width           = 47 // Increased to accommodate the larger spritesheet
	height          = 27 // Increased to accommodate the taller spritesheet
	unit            = 8
	spriteCellSize  = 8  // Size of each sprite cell in the spritesheet
	spriteSheetCols = 32 // Number of columns in the spritesheet
	spriteSheetRows = 24 // Number of rows in the spritesheet
)

var squareColors [64][64]int      // Up to 64x64 grid to store square colors
var spritesheet [24][32][8][8]int // 24x32 grid of 8x8 sprites

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
		for r := 0; r < g.gridSize; r++ {
			for c := 0; c < g.gridSize; c++ {
				// Calculate the sprite position
				sprRow := baseRow + r
				sprCol := baseCol + c

				// Make sure we don't go out of bounds
				if sprRow < spriteSheetRows && sprCol < spriteSheetCols {
					// Check flag state
					if g.spriteFlags[sprRow][sprCol][i] {
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
	for gridRow := 0; gridRow < g.gridSize; gridRow++ {
		for gridCol := 0; gridCol < g.gridSize; gridCol++ {
			// Calculate the sprite position
			sprRow := baseRow + gridRow
			sprCol := baseCol + gridCol

			// Make sure we don't go out of bounds
			if sprRow < spriteSheetRows && sprCol < spriteSheetCols {
				// Copy this sprite's pixels to the appropriate section of the drawing canvas
				for r := 0; r < 8; r++ {
					for c := 0; c < 8; c++ {
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
