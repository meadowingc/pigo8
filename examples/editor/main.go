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

const (
	// Sprite dimensions
	spriteSize = 8 // Size of each sprite in pixels

	// Grid sizes
	defaultGridSize = 1 // 8x8 grid (1 sprite)
	mediumGridSize  = 2 // 16x16 grid (4 sprites)
	largeGridSize   = 4 // 32x32 grid (16 sprites)

	// Map dimensions
	mapWidth  = 320
	mapHeight = 320

	// Default colors
	defaultColor     = 1 // Default color (usually red in PICO-8 palette)
	transparentColor = 0 // Transparent color

	// UI constants
	paletteColumns = 8 // Number of columns in the palette display
	numFlags       = 8 // Number of sprite flags

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
	mapCameraX int                      // Camera X position in the map (in sprites)
	mapCameraY int                      // Camera Y position in the map (in sprites)
	mapData    [mapWidth][mapHeight]int // The map data - stores sprite indices
}

type mapData struct {
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

// forEachSpritePixel iterates over every pixel in every sprite in the spritesheet
// and calls the provided function with the current coordinates
func forEachSpritePixel(fn func(row, col, r, c int)) {
	for row := 0; row < spriteSheetRows; row++ {
		for col := 0; col < spriteSheetCols; col++ {
			for r := 0; r < 8; r++ {
				for c := 0; c < 8; c++ {
					fn(row, col, r, c)
				}
			}
		}
	}
}

// forEachSelectedSprite iterates over each sprite in the current selection grid
// and calls the provided function with the sprite's row and column coordinates
func (g *myGame) forEachSelectedSprite(fn func(row, col int)) {
	baseRow := g.currentSprite / spriteSheetCols
	baseCol := g.currentSprite % spriteSheetCols
	size := g.gridSize
	if size < 1 {
		size = 1
	}
	for r := 0; r < size; r++ {
		for c := 0; c < size; c++ {
			sprRow := baseRow + r
			sprCol := baseCol + c
			// Make sure we don't go out of bounds
			if sprRow >= 0 && sprRow < spriteSheetRows && sprCol >= 0 && sprCol < spriteSheetCols {
				fn(sprRow, sprCol)
			}
		}
	}
}

func (g *myGame) Init() {
	initSquareColors()

	// Initialize sprite flags to false
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			for flag := range numFlags {
				spriteFlags[row][col][flag] = false
			}
		}
	}

	// Initialize spritesheet (will also load from file if available)
	initSpritesheet()

	// Try to load map data from map.json
	if err := g.loadMapData(); err != nil {
		fmt.Println("No map.json found, starting with empty map")
		// Everytime you get out of map mode, save the map
		err := g.saveMapData()
		if err != nil {
			fmt.Println("Error saving map:", err)
			os.Exit(1)
		}
		fmt.Println("Map saved to map.json")
	}

	if err := g.loadMapData(); err != nil {
		fmt.Println("Could not create map.json")
		os.Exit(1)
	}

	g.currentColor = defaultColor // Default color (usually red in PICO-8 palette)
	g.currentSprite = 1           // Default to first non-transparent sprite (sprite 0 is reserved)
	g.hoverX = -1                 // No hover initially
	g.hoverY = -1                 // No hover initially
	g.gridSize = defaultGridSize  // Start with 8x8 grid (1 sprite)

	// Ensure grid size is never less than 1
	if g.gridSize < defaultGridSize {
		g.gridSize = defaultGridSize
	}
	g.lastWheelTime = 0 // Initialize wheel time

	// Initialize the drawing canvas with the default sprite (1)
	g.updateDrawingCanvas()
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

// convertSpriteToData converts a sprite at the given row and column to PIGO8's spriteData format
func convertSpriteToData(row, col int) spriteData {
	// Calculate sprite index
	spriteIndex := row*spriteSheetCols + col

	// Create a new sprite
	sprite := spriteData{
		ID:     spriteIndex,
		X:      col * spriteSize,
		Y:      row * spriteSize,
		Width:  spriteSize,
		Height: spriteSize,
		Used:   true, // Mark all sprites as used
		Flags: p8.FlagsData{
			Bitfield:   0, // Will be calculated below
			Individual: make([]bool, 8),
		},
		Pixels: make([][]int, 8),
	}

	// Fill in the flags
	bitfield := 0
	for i := 0; i < numFlags; i++ {
		flagValue := spriteFlags[row][col][i]
		sprite.Flags.Individual[i] = flagValue
		if flagValue {
			bitfield |= 1 << i
		}
	}
	sprite.Flags.Bitfield = bitfield

	// Initialize pixel data
	for r := 0; r < spriteSize; r++ {
		sprite.Pixels[r] = make([]int, spriteSize)
		for c := 0; c < spriteSize; c++ {
			sprite.Pixels[r][c] = spritesheet[row][col][r][c]
		}
	}

	return sprite
}

// applySpriteData applies PIGO8's spriteData format to the spritesheet at the given position
func applySpriteData(sprite spriteData) {
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

// loadSpritesheet loads the spritesheet from spritesheet.json if it exists
func loadSpritesheet() error {
	// Check if spritesheet.json exists
	data, err := os.ReadFile("spritesheet.json")
	if err != nil {
		// File doesn't exist or can't be read, just use the default empty spritesheet
		fmt.Println("No spritesheet.json found, using empty spritesheet")
		return err
	}

	// Parse the JSON data
	var sheet spriteSheetData
	err = json.Unmarshal(data, &sheet)
	if err != nil {
		return fmt.Errorf("error parsing spritesheet.json: %w", err)
	}

	// Load the sprites into the spritesheet
	for _, sprite := range sheet.Sprites {
		applySpriteData(sprite)
	}

	fmt.Println("Loaded spritesheet from spritesheet.json")
	return nil
}

// saveSpritesheet saves the current spritesheet to a JSON file
func saveSpritesheet() error {
	// Create the spritesheet structure following the PIGO8 format
	sheet := spriteSheetData{
		SpriteSheetColumns: spriteSheetCols,
		SpriteSheetRows:    spriteSheetRows,
		SpriteSheetWidth:   spriteSheetCols * spriteSize, // Each sprite is 8x8 pixels
		SpriteSheetHeight:  spriteSheetRows * spriteSize, // Each sprite is 8x8 pixels
		Sprites:            make([]spriteData, 0, spriteSheetRows*spriteSheetCols),
	}

	// Convert all sprites
	for row := 0; row < spriteSheetRows; row++ {
		for col := 0; col < spriteSheetCols; col++ {
			sheet.Sprites = append(sheet.Sprites, convertSpriteToData(row, col))
		}
	}

	return saveJSONToFile("spritesheet.json", sheet)
}

func (g *myGame) Update() {
	// Check if 'X' button is pressed to switch to map mode
	if p8.Btnp(p8.X) {
		g.mapMode = !g.mapMode

		if g.mapMode {
			// Everytie you get into map mode, save the spritesheet
			err := saveSpritesheet()
			if err != nil {
				fmt.Println("Error saving spritesheet:", err)
				os.Exit(1)
			}
			// fmt.Println("Spritesheet saved to spritesheet.json")
		} else {
			// Everytime you get out of map mode, save the map
			err := g.saveMapData()
			if err != nil {
				fmt.Println("Error saving map:", err)
				os.Exit(1)
			}
			// fmt.Println("Map saved to map.json")
		}
	}

	// Handle map mode controls
	if g.mapMode {

		// Move camera with arrow keys (full screen = 16 sprites = 128 pixels)
		visibleSpritesX := mapViewWidth / unit
		if p8.Btnp(p8.LEFT) && g.mapCameraX > 0 {
			g.mapCameraX -= visibleSpritesX // Move left by one screen
		}
		if p8.Btnp(p8.RIGHT) && g.mapCameraX < mapWidth-visibleSpritesX {
			g.mapCameraX += visibleSpritesX // Move right by one screen
		}
		visibleSpritesY := mapViewHeight / unit
		if p8.Btnp(p8.UP) && g.mapCameraY > 0 {
			g.mapCameraY -= visibleSpritesY // Move up by one screen
		}
		if p8.Btnp(p8.DOWN) && g.mapCameraY < mapHeight-visibleSpritesY {
			g.mapCameraY += visibleSpritesY // Move down by one screen
		}

		// Handle sprite placement
		mx, my := p8.Mouse()

		// Check if mouse is within map bounds
		if mx >= 10 && mx < 10+mapViewWidth && my >= 10 && my < 10+mapViewHeight {
			// Calculate map coordinates from mouse position
			mapX := g.mapCameraX + (mx-10)/8
			mapY := g.mapCameraY + (my-10)/8

			// Handle right click to erase (set to sprite 0)
			if p8.Btn(p8.MouseRight) {
				// Check if target position is within map bounds
				if mapX >= 0 && mapX < 320 && mapY >= 0 && mapY < 320 {
					p8.Mset(mapX, mapY, 0)    // Set to sprite 0 (empty/transparent)
					g.mapData[mapY][mapX] = 0 // Update internal map data
				}
				return
			}

			// Place sprite on left click
			// Place sprite(s) on left click
			if p8.Btn(p8.MouseLeft) {
				// Calculate grid dimensions based on current grid size
				gridWidth := 1
				gridHeight := 1
				switch g.gridSize {
				case 2: // 16x16
					gridWidth, gridHeight = 2, 2
				case 4: // 32x32
					gridWidth, gridHeight = 4, 4
				}

				// Calculate map coordinates from mouse position
				mapX := g.mapCameraX + (mx-10)/8
				mapY := g.mapCameraY + (my-10)/8

				// Get the base sprite (top-left of selection)
				baseSprite := g.currentSprite

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
							g.mapData[targetY][targetX] = spriteIndex // Update internal map data
						}
					}
				}
			}
		}

		return // Skip rest of update when in map mode
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
				baseSprite := g.currentSprite
				baseRow := baseSprite / spriteSheetCols
				baseCol := baseSprite % spriteSheetCols

				// Get the current state of the flag from the base sprite
				currentState := spriteFlags[baseRow][baseCol][i]
				// Toggle to the opposite state
				newState := !currentState

				// Apply the flag change to all selected sprites based on grid size
				// If gridSize is less than 1, default to 1 to ensure at least one sprite is affected
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
	gridSize := 8 * g.gridSize  // Actual pixel dimensions (8, 16, 32, or 64)
	cellSize := 96 / gridSize   // 96 is the total space (8*12) divided by number of cells
	cellSize = max(1, cellSize) // Ensure cell size is at least 1 pixel

	// Calculate row and column based on cell size
	row := (my - gridStartY) / cellSize
	col := (mx - gridStartX) / cellSize

	// Check if mouse is over the drawing grid
	if row >= 0 && row < gridSize && col >= 0 && col < gridSize {
		// Calculate the base sprite (top-left of the selection)
		baseSprite := g.currentSprite
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
		g.hoverX = sprCol*8 + spritePixelCol
		g.hoverY = sprRow*8 + spritePixelRow

		// Make sure we don't go out of bounds
		if sprRow < spriteSheetRows && sprCol < spriteSheetCols {
			// Handle mouse clicks for drawing
			if p8.Btn(p8.MouseLeft) { // Left mouse button
				// Update both the visible drawing grid and the selected sprite
				setSquareColor(row, col, g.currentColor)                                     // Set color in the visible grid
				spritesheet[sprRow][sprCol][spritePixelRow][spritePixelCol] = g.currentColor // Set color in the sprite
				// Update the sprite in PIGO8
				p8.Sset(sprCol*8+spritePixelCol, sprRow*8+spritePixelRow, g.currentColor)
				// Update any map tiles using this sprite
				spriteIndex := sprRow*spriteSheetCols + sprCol
				updateMapSprites(spriteIndex)
				// Update the drawing canvas to reflect changes
				g.updateDrawingCanvas()
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
				g.updateDrawingCanvas()
			}
		}
	} else {
		// Not hovering over the grid
		g.hoverX = -1
		g.hoverY = -1
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
				g.currentSprite = spriteIndex
			}

			// Update the entire drawing canvas to reflect the new sprite selection
			// This ensures all pixels (including transparent ones) are properly refreshed
			g.updateDrawingCanvas()
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
				g.currentColor = colorIndex
			}
		}
	}

	// Handle mouse wheel scrolling to adjust grid size with debouncing
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	// Only process wheel events if enough time has passed (500ms delay)
	if currentTime-g.lastWheelTime > 500 {
		if p8.Btnp(p8.MouseWheelUp) {
			// Increase grid size: 1 (8x8) -> 2 (16x16) -> 4 (32x32)
			// Limit to 4 (32x32) as the maximum grid size
			if g.gridSize < 4 {
				// Ensure grid size is at least 1 before multiplying
				if g.gridSize < 1 {
					g.gridSize = 1
				}
				g.gridSize *= 2
				// Update the drawing canvas to show the selected sprites
				g.updateDrawingCanvas()
				// Update the last wheel time
				g.lastWheelTime = currentTime
			}
		} else if p8.Btnp(p8.MouseWheelDown) {
			// Decrease grid size: 4 (32x32) -> 2 (16x16) -> 1 (8x8)
			if g.gridSize > 1 {
				g.gridSize /= 2
				// Ensure grid size is never less than 1
				if g.gridSize < 1 {
					g.gridSize = 1
				}
				// Update the drawing canvas to show the selected sprites
				g.updateDrawingCanvas()
				// Update the last wheel time
				g.lastWheelTime = currentTime
			}
		}
	}
}

func (g *myGame) Draw() {
	p8.ClsRGBA(color.RGBA{25, 25, 25, 255})

	if g.mapMode {
		g.drawMapMode()
		return
	}

	g.updateDrawingCanvas()
	g.drawEditorCanvas()
	g.drawSpritesheetPanel()
	g.drawSelectionAndPalette()
}

// drawMapMode draws everything when in map‐editing mode
func (g *myGame) drawMapMode() {
	const (
		viewX = 10
		viewY = 10
		unit8 = 8
	)
	// 1) map viewport tiles
	g.drawMapTiles(viewX, viewY)

	// 2) hover highlight on map
	mx, my := p8.Mouse()
	g.drawMapHover(viewX, viewY, mx, my)

	// 3) border and UI text
	p8.Rect(viewX, viewY,
		viewX+mapViewWidth, viewY+mapViewHeight, 7)
	p8.Camera() // reset for text
	g.printMapInfo(viewX, viewY, mx, my)
}

func (g *myGame) drawMapTiles(vx, vy int) {
	cols := mapViewWidth / unit
	rows := mapViewHeight / unit
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			spr := p8.Mget(g.mapCameraX+x, g.mapCameraY+y)
			p8.Spr(spr, float64(vx+x*8), float64(vy+y*8))
		}
	}
}

// drawMapHover draws the hover highlight on the map
func (g *myGame) drawMapHover(vx, vy, mx, my int) {
	cols := mapViewWidth / unit
	rows := mapViewHeight / unit
	// determine multi‐sprite grid
	w, h := 1, 1
	if g.gridSize >= 2 {
		w, h = g.gridSize, g.gridSize
	}
	gx, gy := (mx-vx)/8, (my-vy)/8
	if gx < 0 || gx >= cols || gy < 0 || gy >= rows {
		return
	}
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			x, y := gx+dx, gy+dy
			if x < cols && y < rows {
				p8.Rect(
					float64(vx+x*8),
					float64(vy+y*8),
					float64(vx+(x+1)*8-1),
					float64(vy+(y+1)*8-1),
					7,
				)
			}
		}
	}
}

// printMapInfo prints the map info
func (g *myGame) printMapInfo(vx, vy, mx, my int) {
	// Screen coords
	p8.Color(7)
	sx := g.mapCameraX / (mapViewWidth / unit)
	sy := g.mapCameraY / (mapViewHeight / unit)
	textY := vy + mapViewHeight + 10
	p8.Print(fmt.Sprintf("Screen: %d,%d", sx, sy), vx, textY, 7)

	// Mouse in map space
	if mx < vx || mx >= vx+mapViewWidth ||
		my < vy || my >= vy+mapViewHeight {
		return
	}
	mxMap := g.mapCameraX + (mx-vx)/8
	myMap := g.mapCameraY + (my-vy)/8
	p8.Print(fmt.Sprintf("Map: %d,%d", mxMap, myMap),
		vx+90, textY, 7)

	if mxMap >= 0 && mxMap < 128 && myMap >= 0 && myMap < 128 {
		spr := p8.Mget(mxMap, myMap)
		p8.Print(fmt.Sprintf("Sprite: %d", spr),
			vx+180, textY, 7)
	}
}

// drawEditorCanvas draws the non‐map “editor” canvas
func (g *myGame) drawEditorCanvas() {
	const startX, startY = 10, 10
	endX := startX + 8*12 - 2
	endY := startY + 8*12 - 2

	gridPx := 8 * g.gridSize
	cell := max(1, 96/gridPx)
	// draw each cell
	for row := 0; row < gridPx; row++ {
		for col := 0; col < gridPx; col++ {
			color := getSquareColor(row, col)
			x := startX + col*cell
			y := startY + row*cell
			if cell > 1 {
				p8.Rectfill(x, y, x+cell-1, y+cell-1, color)
			} else {
				p8.Pset(x, y, color)
			}
		}
	}
	// hover text
	if g.hoverX >= 0 && g.hoverY >= 0 {
		p8.Print(
			fmt.Sprintf("pixel: (%d,%d)", g.hoverX, g.hoverY),
			startX, startY-10, 7,
		)
	}
	p8.Rect(startX-1, startY-1, endX+1, endY+1, 7)
}

// drawSpritesheetPanel draws the spritesheet area and label
func (g *myGame) drawSpritesheetPanel() {
	sx, sy := spritesheetStartX, 10
	ex := sx + spriteSheetCols*spriteCellSize
	ey := sy + spriteSheetRows*spriteCellSize

	// draw each sprite tile
	for r := 0; r < spriteSheetRows; r++ {
		for c := 0; c < spriteSheetCols; c++ {
			baseX := sx + c*spriteCellSize
			baseY := sy + r*spriteCellSize
			for py := 0; py < 8; py++ {
				for px := 0; px < 8; px++ {
					col := spritesheet[r][c][py][px]
					if col == 0 {
						p8.Pset(baseX+px, baseY+py, 0)
					} else {
						p8.Pset(baseX+px, baseY+py, col)
					}
				}
			}
		}
	}

	// draw selection border
	if g.gridSize >= 1 {
		g.drawSelectionBorder(sx, sy)
	}
	p8.Rect(sx-1, sy-1, ex+1, ey+1, 7)

	sizeText := map[int]string{1: "8x8", 2: "16x16", 4: "32x32"}[g.gridSize]
	p8.Print(
		fmt.Sprintf("spritesheet - sprite: %d - grid: %s",
			g.currentSprite, sizeText),
		sx, ey+4, 7,
	)
}

// drawSelectionBorder highlights the multi‐cell selection in the spritesheet
func (g *myGame) drawSelectionBorder(sx, sy int) {
	cols, rows := spriteSheetCols, spriteSheetRows
	base := g.currentSprite
	br := base / cols
	bc := base % cols

	x1 := sx + bc*spriteCellSize - 1
	y1 := sy + br*spriteCellSize - 1
	x2 := x1 + g.gridSize*spriteCellSize + 1
	y2 := y1 + g.gridSize*spriteCellSize + 1

	maxX := sx + cols*spriteCellSize
	maxY := sy + rows*spriteCellSize
	if x2 > maxX {
		x2 = maxX
	}
	if y2 > maxY {
		y2 = maxY
	}

	p8.Rect(x1, y1, x2, y2, 7)
}

// drawSelectionAndPalette draws the selection and palette
func (g *myGame) drawSelectionAndPalette() {
	// spacing constants
	const offset = 15
	const paletteOffset = 40
	g.drawCheckboxes(10, 10+8*12-2+offset)
	g.drawPalette(10, 10+8*12-2+paletteOffset)
}

var (
	width           = 52  // Increased to accommodate the larger spritesheet and more space
	height          = 27  // Increased to accommodate the taller spritesheet
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
	if err := loadSpritesheet(); err != nil {
		// Only initialize with transparent colors if loading failed
		forEachSpritePixel(func(row, col, r, c int) {
			// Initialize with transparent color (0)
			spritesheet[row][col][r][c] = 0
		})
	}

	// Initialize the spritesheet in PIGO8 with our current data
	forEachSpritePixel(func(row, col, r, c int) {
		// Calculate the absolute pixel position
		px := col*8 + c
		py := row*8 + r
		// Set the pixel color in PIGO8
		p8.Sset(px, py, spritesheet[row][col][r][c])
	})
}

func initPico8Spritesheet() error {
	// Check if spritesheet.json already exists
	if _, err := os.Stat("spritesheet.json"); err == nil {
		// File exists, no need to create it
		return nil
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("error checking spritesheet.json: %w", err)
	}

	// Create a temporary spritesheet.json file that PIGO8 can load
	createTempSpritesheet()

	// Now initialize all sprites with our data
	forEachSpritePixel(func(row, col, r, c int) {
		// Calculate the absolute pixel position
		px := col*8 + c
		py := row*8 + r
		// Set the pixel color in PIGO8
		p8.Sset(px, py, spritesheet[row][col][r][c])
	})

	return nil
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

	// Create all sprites first
	for row := range spriteSheetRows {
		for col := range spriteSheetCols {
			spriteIndex := row*spriteSheetCols + col

			// Create a sprite with basic data
			sprite := spriteData{
				ID:     spriteIndex,
				X:      col * 8,
				Y:      row * 8,
				Width:  8,
				Height: 8,
				Used:   true,
				Flags: p8.FlagsData{
					Bitfield:   0,
					Individual: make([]bool, 8),
				},
				Pixels: make([][]int, 8),
			}

			// Initialize pixel arrays
			for r := range 8 {
				sprite.Pixels[r] = make([]int, 8)
			}

			// Add to sprites array
			sprites[spriteIndex] = sprite
		}
	}

	// Fill in pixel data
	forEachSpritePixel(func(row, col, r, c int) {
		spriteIndex := row*spriteSheetCols + col
		sprites[spriteIndex].Pixels[r][c] = spritesheet[row][col][r][c]
	})

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

	// Draw 8 checkboxes in a row
	for i := range 8 {
		checkboxX := x + i*checkboxSize*3/2 // Space them out a bit
		checkboxY := y

		// Draw checkbox outline
		p8.Rect(checkboxX, checkboxY, checkboxX+checkboxSize-1, checkboxY+checkboxSize-1, 7)

		// Check the flag state across all selected sprites
		allTrue := true
		allFalse := true

		// Check flag state for all selected sprites
		g.forEachSelectedSprite(func(sprRow, sprCol int) {
			if spriteFlags[sprRow][sprCol][i] {
				allFalse = false // At least one is true
			} else {
				allTrue = false // At least one is false
			}
		})

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
func (g *myGame) updateDrawingCanvas() {
	// Clear the drawing canvas
	for row := range 64 {
		for col := range 64 {
			squareColors[row][col] = 0
		}
	}

	// Copy the sprites to the drawing canvas
	g.forEachSelectedSprite(func(srcRow, srcCol int) {
		// Calculate relative position from base sprite
		baseRow := g.currentSprite / spriteSheetCols
		baseCol := g.currentSprite % spriteSheetCols
		r := srcRow - baseRow
		c := srcCol - baseCol

		// Copy the sprite to the drawing canvas
		for pixelRow := 0; pixelRow < 8; pixelRow++ {
			for pixelCol := 0; pixelCol < 8; pixelCol++ {
				// Calculate destination position in drawing canvas
				dstRow := r*8 + pixelRow
				dstCol := c*8 + pixelCol

				// Copy the pixel
				squareColors[dstRow][dstCol] = spritesheet[srcRow][srcCol][pixelRow][pixelCol]
			}
		}
	})
}

// saveJSONToFile saves any data structure to a JSON file with proper indentation
func saveJSONToFile(filename string, data interface{}) error {
	// Convert to JSON with indentation
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing %s: %w", filename, err)
	}

	return nil
}

// loadJSONFromFile loads a JSON file into the provided data structure
func loadJSONFromFile(filename string, data any) error {
	// Read the file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading %s: %w", filename, err)
	}

	// Parse the JSON data
	if err := json.Unmarshal(jsonData, data); err != nil {
		return fmt.Errorf("error parsing %s: %w", filename, err)
	}

	return nil
}

// convertMapToData converts the game's map data to the PIGO8 MapData format
func (g *myGame) convertMapToData() mapData {
	mapData := mapData{
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
			sprite := g.mapData[y][x]
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

	return mapData
}

// applyMapData applies the PIGO8 MapData format to the game's map
func (g *myGame) applyMapData(mapData mapData) {
	// Initialize map with zeros
	for y := range g.mapData {
		for x := range g.mapData[y] {
			g.mapData[y][x] = 0
		}
	}

	// Load the cells into our map data
	for _, cell := range mapData.Cells {
		// Make sure coordinates are within bounds
		if cell.X >= 0 && cell.X < 320 && cell.Y >= 0 && cell.Y < 320 {
			g.mapData[cell.Y][cell.X] = cell.Sprite
			// Also update the PIGO8 map
			p8.Mset(cell.X, cell.Y, cell.Sprite)
		}
	}

	fmt.Printf("Loaded map: %dx%d with %d cells\n", mapData.Width, mapData.Height, len(mapData.Cells))
}

// saveMapData saves the current map to map.json
func (g *myGame) saveMapData() error {
	mapData := g.convertMapToData()
	return saveJSONToFile("map.json", mapData)
}

// loadMapData loads the map from map.json if it exists
func (g *myGame) loadMapData() error {
	var mapData mapData
	if err := loadJSONFromFile("map.json", &mapData); err != nil {
		return err
	}

	g.applyMapData(mapData)
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

	// Initialize spritesheet if it doesn't exist
	if err := initPico8Spritesheet(); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing spritesheet: %v\n", err)
		os.Exit(1)
	}

	settings := p8.NewSettings()
	settings.ScreenWidth = width * unit
	settings.ScreenHeight = height * unit
	settings.ScaleFactor = 5
	p8.InsertGame(&myGame{})
	p8.PlayGameWith(settings)
}
