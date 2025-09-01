# PICO-8 Camera Examples

This directory contains comprehensive examples demonstrating the camera functionality that matches the [PICO-8 Camera Guide](https://nerdyteachers.com/PICO-8/Guide/CAMERA) 1:1.

## Examples Overview

- **Example 1:** Basic drawing with no camera offset
- **Example 2:** Camera offset affecting previously drawn elements (retroactive)
- **Example 3:** Using two cameras to lock UI elements in place

### 1. `camera_example1/` - Basic Camera (No Offset)
**Run:** `cd camera_example1 && go run main.go`

Demonstrates the first example from the PICO-8 guide:
```go
p8.Cls()                      // Clear screen
p8.Rectfill(0, 0, 127, 127, 2) // Dark purple background
p8.Rect(0, 0, 127, 127, 8)     // Red outline
p8.Print("camera(0,0)", 2, 2)  // Label text
```

This shows basic drawing operations with the default camera position (0,0) - no offset applied.

### 2. `camera_example2/` - Retroactive Camera Effect
**Run:** `cd camera_example2 && go run main.go`

Demonstrates how camera offset affects **previously drawn elements**:
```go
p8.Cls()
p8.Rectfill(0, 0, 127, 127, 2) // Draw background first
p8.Rect(0, 0, 127, 127, 8)     // Draw outline first
p8.Print("camera(0,0)", 2, 2)  // Draw text first

p8.Camera(63, 63)               // NOW set camera offset
p8.Rect(63, 63, 127+63, 127+63, 11) // Draw new elements
p8.Print("camera(63,63)", 136, 182)
```

**Key Point:** The camera offset affects the previously drawn background, outline, and text even though they were drawn before `Camera(63, 63)` was called.

### 3. `camera_example3/` - Locked UI Elements
**Run:** `cd camera_example3 && go run main.go`

Demonstrates using **two camera calls** to create locked UI overlays:
```go
p8.Cls()
p8.Camera()                     // First camera call - LOCKS following elements
p8.Rectfill(0, 0, 127, 127, 2)  // These elements get locked in place
p8.Rect(0, 0, 127, 127, 8)      
p8.Print("camera(0,0)", 2, 2)   

p8.Camera(63, 63)               // Second camera call - only affects new elements
p8.Rect(63, 63, 190, 190, 11)   // These elements are offset
p8.Print("camera(63,63)", 136, 182)
```

**Key Point:** The first set of elements (background, outline, text) are **locked in position** by the first `Camera()` call and are NOT affected by the second `Camera(63, 63)` call.

## Camera Behavior Summary

### Core Functionality
- `Camera()` - Resets camera to (0,0)
- `Camera(x, y)` - Sets camera offset to (x, y)
- Camera affects ALL drawing operations: Shapes, Sprites, Text, Maps, Pixels

### Key Behaviors
1. **Retroactive Effect:** Camera offset affects previously drawn elements unless locked
2. **Locking Elements:** Calling `Camera()` locks previously drawn elements in place
3. **Coordinate Transformation:** Camera subtracts offset from drawing coordinates (`x - cameraX, y - cameraY`)
4. **UI Overlay Pattern:** Use two camera calls to create fixed UI overlays that don't move with the game world

### Common Use Cases
```go
// Game world with camera following player
Camera(playerX - 64, playerY - 64)
Map() // Draw scrolling world

// Fixed UI overlay
Camera() // Reset and lock UI elements
Print("SCORE: " + score, 2, 2) // UI doesn't move with world
Print("HEALTH: " + health, 2, 10)
```

## Implementation Notes

The PIGO-8 camera implementation is a faithful 1:1 reproduction of PICO-8's camera system:
- Identical function signatures and behavior
- Same coordinate transformation math
- Same retroactive and locking behavior
- Compatible with all drawing functions (shapes, sprites, text, maps)

All examples match the behavior described in the [NerdyTeachers PICO-8 Camera Guide](https://nerdyteachers.com/PICO-8/Guide/CAMERA) exactly.
