package pigo8

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
)

// mapCell holds a single cell's data from a PICO-8 map export.
type mapCell struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Sprite int `json:"sprite"`
}

// MapData represents the structure of a PICO-8 map JSON export.
type MapData struct {
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	Name        string    `json:"name"`
	Cells       []mapCell `json:"cells"`
}

var (
	currentMap    *MapData
	spriteInfoMap map[int]*SpriteInfo

	// Memory monitoring
	lastMemoryUsage uint64
	memoryMutex     sync.Mutex
)

// logMemory logs the current memory usage with a label
func logMemory(label string, forceLog bool) {
	memoryMutex.Lock()
	defer memoryMutex.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert to MB for comparison and display
	currentUsageMB := float64(m.Alloc) / 1024 / 1024

	// If this is the first check, just store the value
	if lastMemoryUsage == 0 {
		lastMemoryUsage = m.Alloc
		log.Printf("PICO-8 Memory: %.2f MB", currentUsageMB)
		return
	}

	// Calculate memory difference
	diffBytes := int64(m.Alloc) - int64(lastMemoryUsage)
	diffMB := float64(diffBytes) / 1024 / 1024

	// Log if forced or if memory changed significantly
	if forceLog || diffMB > 1.0 || diffMB < -1.0 {
		log.Printf("PICO-8 Memory (%s): %.2f MB (change: %.2f MB)", label, currentUsageMB, diffMB)
		lastMemoryUsage = m.Alloc
	}
}

// loadMap reads and unmarshals "map.json" from the current directory.
func loadMap() (*MapData, error) {
	const mapFilename = "map.json"

	// Log memory before loading map
	logMemory("before map load", false)

	data, err := os.ReadFile(mapFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: ensure '%s' exists in your game's base directory; this file is required for map functionality", err, mapFilename)
		}
		log.Printf("Error reading %s: %v", mapFilename, err)
		return nil, fmt.Errorf("error reading %s: %w", mapFilename, err)
	}

	// Log memory after reading file
	logMemory("after reading map file", false)

	var m MapData
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("error unmarshalling %s: %w", mapFilename, err)
	}

	// Always log when map is loaded, regardless of memory change
	fileSize := float64(len(data)) / 1024
	log.Printf("Map: %dx%d tiles, %d cells (%.1f KB)", m.Width, m.Height, len(m.Cells), fileSize)
	logMemory("after map load", true)

	return &m, nil
}

// Map draws a rectangular region of the PICO-8 map to the screen.
// Optional args: [mx, my, sx, sy, w, h, layers]
//   - mx, my: map tile coordinates in tiles (defaults 0,0)
//   - sx, sy: screen pixel coordinates to draw at (defaults 0,0)
//   - w, h: dimensions in tiles (defaults 16x16)
//   - layers: bitfield to filter sprites by their flags (0 = draw all)
func Map(args ...any) {
	// Default map coordinates
	mx, my := 0, 0

	// If arguments are provided, extract map coordinates
	if len(args) >= 1 {
		if mxVal, ok := args[0].(int); ok {
			mx = mxVal
		} else if mxVal, ok := args[0].(float64); ok {
			mx = int(mxVal)
		}
	}

	if len(args) >= 2 {
		if myVal, ok := args[1].(int); ok {
			my = myVal
		} else if myVal, ok := args[1].(float64); ok {
			my = int(myVal)
		}
	}

	// Pass remaining arguments to the generic implementation
	var remainingArgs []any
	if len(args) > 2 {
		remainingArgs = args[2:]
	}

	// Call the generic implementation
	MapG(mx, my, remainingArgs...)
}

// MapG is the generic version of Map that accepts any number type for coordinates.
// The mx and my coordinates can be any integer or float type (e.g., int, float64)
// due to the use of generics [MX Number, MY Number]. They are converted internally
// to integers for map calculations.
//
// Optional args: [sx, sy, w, h, layers]
//   - sx, sy: screen pixel coordinates to draw at (defaults 0,0)
//   - w, h: dimensions in tiles (defaults 16x16)
//   - layers: bitfield to filter sprites by their flags (0 = draw all)
//
// Usage:
//
//	MapG(0, 0) // Draw map at (0,0) with default size
//	MapG(mx, my) // Draw map at specified coordinates
//	MapG(mx, my, sx, sy) // Draw map at specified coordinates with screen offset
//	MapG(mx, my, sx, sy, w, h) // Draw map with custom dimensions
//	MapG(mx, my, sx, sy, w, h, layers) // Draw map with layer filtering
func MapG[MX Number, MY Number](mx MX, my MY, args ...any) {
	if !ensureMapResources() {
		return
	}

	// Convert generic mx, my to required types
	mapX := int(mx)
	mapY := int(my)

	// Parse arguments
	sx, sy, wTiles, hTiles, layers := parseMapArgs(args)
	if wTiles <= 0 || hTiles <= 0 {
		return
	}

	// Draw the map region
	drawMapRegion(mapX, mapY, sx, sy, wTiles, hTiles, layers)
}

// ensureMapResources ensures all required resources for map rendering are loaded
// Returns false if resources couldn't be loaded or screen isn't ready
func ensureMapResources() bool {
	if currentScreen == nil {
		log.Println("Warning: Map() called before screen was ready.")
		return false
	}

	// Lazy-load map data
	if currentMap == nil {
		m, err := loadMap()
		if err != nil {
			log.Fatalf("Fatal: Failed to load required map for Map(): %v", err)
		}
		currentMap = m
	}

	// Lazy-load sprite info for layer filtering
	if currentSprites == nil {
		sprites, err := loadSpritesheet()
		if err != nil {
			log.Fatalf("Fatal: Failed to load required spritesheet for Map(): %v", err)
		}
		currentSprites = sprites
	}

	// Build lookup map from sprite ID to SpriteInfo
	if spriteInfoMap == nil {
		spriteInfoMap = make(map[int]*SpriteInfo, len(currentSprites))
		for i := range currentSprites {
			info := &currentSprites[i]
			spriteInfoMap[info.ID] = info
		}
	}

	return true
}

// parseMapArgs parses the optional arguments for the Map functions
// Returns screen x, screen y, width in tiles, height in tiles, and layers bitfield
func parseMapArgs(args []any) (sx, sy, wTiles, hTiles, layers int) {
	// Default parameters
	sx, sy = 0, 0
	wTiles = LogicalWidth / 8
	hTiles = LogicalHeight / 8
	layers = 0

	// Process optional arguments
	if len(args) >= 1 {
		if sxVal, ok := args[0].(int); ok {
			sx = sxVal
		} else if sxVal, ok := args[0].(float64); ok {
			sx = int(sxVal)
		}
	}
	if len(args) >= 2 {
		if syVal, ok := args[1].(int); ok {
			sy = syVal
		} else if syVal, ok := args[1].(float64); ok {
			sy = int(syVal)
		}
	}
	if len(args) >= 3 {
		if wVal, ok := args[2].(int); ok {
			wTiles = wVal
		} else if wVal, ok := args[2].(float64); ok {
			wTiles = int(wVal)
		}
	}
	if len(args) >= 4 {
		if hVal, ok := args[3].(int); ok {
			hTiles = hVal
		} else if hVal, ok := args[3].(float64); ok {
			hTiles = int(hVal)
		}
	}
	if len(args) >= 5 {
		if layerVal, ok := args[4].(int); ok {
			layers = layerVal
		} else if layerVal, ok := args[4].(float64); ok {
			layers = int(layerVal)
		}
	}
	if len(args) > 5 {
		log.Printf("Warning: Map() called with too many arguments (%d), expected max 5 ([sx,sy,w,h,layers]).", len(args))
	}

	return sx, sy, wTiles, hTiles, layers
}

// drawMapRegion draws a region of the map to the screen
func drawMapRegion(mapX, mapY, sx, sy, wTiles, hTiles, layers int) {
	// Determine region bounds in map coordinates
	xMin, xMax := mapX, mapX+wTiles
	yMin, yMax := mapY, mapY+hTiles

	for _, cell := range currentMap.Cells {
		// Skip cells outside the visible region
		if cell.X < xMin || cell.X >= xMax || cell.Y < yMin || cell.Y >= yMax {
			continue
		}

		// Get sprite info
		info, ok := spriteInfoMap[cell.Sprite]
		if !ok {
			continue
		}

		// Filter by layers bitfield
		if layers != 0 && (info.Flags.Bitfield&layers) == 0 {
			continue
		}

		// Calculate screen position (8 pixels per tile)
		dx := sx + (cell.X-mapX)*8
		dy := sy + (cell.Y-mapY)*8

		// Draw sprite
		Spr(cell.Sprite, dx, dy)
	}
}
