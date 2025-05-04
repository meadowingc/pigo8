package pigo8

import (
	"golang.org/x/exp/constraints"
)

// CameraNumber covers ints, floats for camera coordinates
type CameraNumber interface {
	constraints.Integer | constraints.Float
}

// Camera state
var (
	// cameraX is the current camera X offset
	cameraX float64
	// cameraY is the current camera Y offset
	cameraY float64
)

// Camera sets the camera position, offsetting all subsequent drawing operations.
// If called with no arguments, resets the camera to (0,0).
// This function mimics PICO-8's camera(x, y) function.
//
// The camera offsets all drawing operations (Shapes, Print, Sprites, and Maps).
//
// Args:
//   - x: (optional) horizontal offset amount
//   - y: (optional) vertical offset amount
//
// Example:
//
//	// Set camera to position (64, 32)
//	Camera(64, 32)
//
//	// Reset camera to (0, 0)
//	Camera()
//
//	// Lock UI elements in place
//	Camera() // Reset camera
//	Print("SCORE: 1000", 2, 2) // Draw UI
//	Camera(playerX-64, playerY-64) // Set camera to follow player
//	Map() // Draw scrolling map
func Camera(args ...interface{}) {
	// Reset camera if no arguments
	if len(args) == 0 {
		cameraX = 0
		cameraY = 0
		return
	}

	// Handle one argument (x only)
	if len(args) == 1 {
		if x, ok := convertToFloat64(args[0]); ok {
			cameraX = x
		}
		return
	}

	// Handle two arguments (x and y)
	if len(args) >= 2 {
		if x, ok := convertToFloat64(args[0]); ok {
			cameraX = x
		}
		if y, ok := convertToFloat64(args[1]); ok {
			cameraY = y
		}
	}
}

// CameraG is the generic version of Camera that accepts any number type for coordinates.
// The x and y coordinates can be any integer or float type (e.g., int, float64)
// due to the use of generics [X CameraNumber, Y CameraNumber]. They are converted internally
// to float64 for camera calculations.
func CameraG[X CameraNumber, Y CameraNumber](x X, y Y) {
	// Set camera position
	cameraX = float64(x)
	cameraY = float64(y)
}

// convertToFloat64 attempts to convert a value to float64
func convertToFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	default:
		return 0, false
	}
}

// ApplyCameraOffset applies the current camera offset to the given coordinates
// and returns the transformed coordinates
func ApplyCameraOffset(x, y float64) (float64, float64) {
	return x - cameraX, y - cameraY
}
