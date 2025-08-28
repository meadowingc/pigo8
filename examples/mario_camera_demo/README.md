# Mario Camera System Demo

This example demonstrates the new PIGO8 camera system using a Super Mario Bros-style platformer. It shows how the camera smoothly follows the player with dead zones, smoothing, and proper map scrolling.

## 🎮 What This Demo Shows

### **1. Smooth Camera Following**
- **Before**: The old system required manual camera calculations and had no built-in smoothing
- **Now**: Just call `Camera(targetX, targetY)` and the camera smoothly follows with lerp interpolation

### **2. Dead Zone System**
- **Horizontal Dead Zone**: 32 pixels - Mario can move left/right without camera following
- **Vertical Dead Zone**: 16 pixels - Smaller for better vertical tracking in platformers
- **Result**: No more jittery camera movement during small player adjustments

### **3. Look-Ahead Camera**
- **LookAheadX**: 24 pixels - Camera looks ahead in movement direction
- **LookAheadY**: 0 pixels - No vertical look-ahead (good for platformers)
- **Result**: Player sees where they're going, not just where they've been

### **4. Automatic Map Scrolling**
- **Map Boundaries**: 1024x240 pixel world
- **Camera Clamping**: Automatically prevents camera from going beyond map edges
- **Result**: Smooth scrolling without manual boundary checks

## 🎯 Key Camera Features Demonstrated

### **Camera Setup**
```go
SetCameraOptions(CameraFollowOptions{
    Lerp:       0.08,  // Smooth camera movement
    DeadZoneW:  32.0,  // Horizontal dead zone
    DeadZoneH:  16.0,  // Vertical dead zone  
    LookAheadX: 24.0,  // Look ahead in movement direction
    LookAheadY: 0.0,   // No vertical look-ahead
    ClampToMap: true,  // Prevent camera from going beyond map
    MapWidth:   1024.0, // World width
    MapHeight:  240.0,  // World height
})
```

### **Camera Update**
```go
func updateCamera() {
    // Calculate target to center Mario on screen
    screenCenterX := float64(GetScreenWidth()) / 2
    screenCenterY := float64(GetScreenHeight()) / 2
    
    targetCameraX := player.position.X - screenCenterX
    targetCameraY := player.position.Y - screenCenterY
    
    // The new system handles all the smoothing and dead zone logic!
    Camera(targetCameraX, targetCameraY)
}
```

### **Automatic Drawing**
```go
func (m *Game) Draw() {
    Cls(7)
    
    // Camera automatically offsets all drawing operations
    Map()         // Level scrolls with camera
    player.draw() // Mario stays centered
}
```

## 🎮 Controls

- **Arrow Keys / A/D**: Move left/right
- **X**: Jump
- **O**: Hold to run
- **Movement**: Smooth physics with acceleration and friction
- **Jumping**: Variable jump height based on movement speed

## 🚀 How to Run

```bash
cd examples/mario_camera_demo
go run main.go
```

## 🔧 Camera Settings Explained

### **Lerp (0.08)**
- **What it does**: Controls how smoothly the camera follows the target
- **Value range**: 0.0 (no movement) to 1.0 (instant movement)
- **Why 0.08**: Smooth enough to feel good, fast enough to keep up with Mario

### **Dead Zone (32x16)**
- **What it does**: Creates a "comfort zone" where small movements don't move the camera
- **Horizontal (32)**: Mario can move 32 pixels left/right before camera follows
- **Vertical (16)**: Smaller for better vertical tracking in platformers

### **Look Ahead (24x0)**
- **What it does**: Moves camera ahead of player in movement direction
- **Horizontal (24)**: Shows what's coming up ahead
- **Vertical (0)**: No vertical look-ahead (good for platformers)

### **Map Clamping (1024x240)**
- **What it does**: Prevents camera from showing empty space beyond map boundaries
- **Width (1024)**: Large world for Mario to explore
- **Height (240)**: Standard Mario height

## 🎯 Comparison: Old vs New System

### **Old System (Manual)**
```go
// Had to manually calculate camera position
var cameraX float64 = 0

func updateCamera() {
    // Complex manual calculations
    regionMin := cameraX + (float64(GetScreenWidth())-StaticRegionWidth)/2 - StaticRegionForwardOffset/2
    regionMax := cameraX + (float64(GetScreenHeight())+StaticRegionWidth)/2 + StaticRegionForwardOffset/2
    
    if player.position.X+16 > regionMax {
        cameraX += (player.position.X + 16 - regionMax)
    }
    if player.position.X < regionMin {
        cameraX -= (regionMin - player.position.X)
    }
    
    // Manual camera setting
    Camera(cameraX, 0)
}
```

### **New System (Automatic)**
```go
func updateCamera() {
    // Simple target calculation
    targetCameraX := player.position.X - GetScreenWidth()/2
    targetCameraY := player.position.Y - GetScreenHeight()/2
    
    // Everything else is automatic!
    Camera(targetCameraX, targetCameraY)
}
```

## 🎨 What You'll See

1. **Smooth Movement**: Mario moves with realistic physics (acceleration, friction, skidding)
2. **Smart Camera**: Camera follows Mario smoothly with dead zones and look-ahead
3. **Large World**: 1024x240 pixel map that scrolls smoothly as Mario moves
4. **Professional Feel**: No jittery camera, smooth scrolling, proper boundaries

This demo shows how the new camera system makes professional-quality camera behavior automatic, letting you focus on game mechanics instead of camera math!
