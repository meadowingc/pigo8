# PIGO8 Palette Management

## Functions

### SetPalette

```go
func SetPalette(newPalette []color.Color)
```

Replaces the current color palette with a new one. This also resizes the transparency array to match the new palette size, setting only the first color (index 0) as transparent by default.

**Parameters:**

- `newPalette`: Slice of `color.Color` values to use as the new palette.

**Example:**

```go
// Create a 4-color grayscale palette
grayscale := []color.Color{
    color.RGBA{0, 0, 0, 255},       // Black
    color.RGBA{85, 85, 85, 255},    // Dark Gray
    color.RGBA{170, 170, 170, 255}, // Light Gray
    color.RGBA{255, 255, 255, 255}, // White
}
pigo8.SetPalette(grayscale)
```

### GetPaletteSize

```go
func GetPaletteSize() int
```

Returns the current number of colors in the palette.

**Returns:**

- The number of colors in the current palette.

**Example:**

```go
// Get the current palette size
size := pigo8.GetPaletteSize()
pigo8.Print(fmt.Sprintf("Palette has %d colors", size), 10, 10, 7)
```

### GetPaletteColor

```go
func GetPaletteColor(colorIndex int) color.Color
```

Returns the `color.Color` at the specified index in the palette. Returns `nil` if the index is out of range.

**Parameters:**

- `colorIndex`: Index of the color to retrieve.

**Returns:**

- The color at the specified index, or `nil` if the index is out of range.

**Example:**

```go
// Get the color at index 3
color3 := pigo8.GetPaletteColor(3)
```

### SetPaletteColor

```go
func SetPaletteColor(colorIndex int, newColor color.Color)
```

Replaces a single color in the palette at the specified index. If the index is out of range, the function does nothing.

**Parameters:**

- `colorIndex`: Index of the color to replace.
- `newColor`: The new `color.Color` to use.

**Example:**

```go
// Change color 7 (white) to a light blue
pigo8.SetPaletteColor(7, color.RGBA{200, 220, 255, 255})
```

## Advanced Usage Examples

### Creating a Custom Palette

```go
// Create a custom palette with 8 colors
customPalette := []color.Color{
    color.RGBA{0, 0, 0, 255},       // Black
    color.RGBA{29, 43, 83, 255},    // Dark Blue
    color.RGBA{126, 37, 83, 255},   // Dark Purple
    color.RGBA{0, 135, 81, 255},    // Dark Green
    color.RGBA{171, 82, 54, 255},   // Brown
    color.RGBA{95, 87, 79, 255},    // Dark Gray
    color.RGBA{194, 195, 199, 255}, // Light Gray
    color.RGBA{255, 241, 232, 255}, // White
}

// Set the palette
pigo8.SetPalette(customPalette)
```

### Creating a Palette Programmatically

```go
// Create a rainbow palette with 12 colors
rainbowPalette := make([]color.Color, 12)

for i := 0; i < 12; i++ {
    // Convert hue (0-360) to RGB
    hue := float64(i) * 30.0 // 12 colors * 30 degrees = 360 degrees
    
    // Simple HSV to RGB conversion (simplified for this example)
    h := hue / 60.0
    sector := int(math.Floor(h))
    f := h - float64(sector)
    
    p := uint8(255 * 0.0)
    q := uint8(255 * (1.0 - f))
    t := uint8(255 * f)
    v := uint8(255)
    
    var r, g, b uint8
    
    switch sector {
    case 0:
        r, g, b = v, t, p
    case 1:
        r, g, b = q, v, p
    case 2:
        r, g, b = p, v, t
    case 3:
        r, g, b = p, q, v
    case 4:
        r, g, b = t, p, v
    default:
        r, g, b = v, p, q
    }
    
    rainbowPalette[i] = color.RGBA{r, g, b, 255}
}

// Set the palette
pigo8.SetPalette(rainbowPalette)
```

### Cycling Colors for Animation

```go
// In your game's Update() function:
func (g *Game) Update() {
    // Every 10 frames, cycle the colors
    if g.frameCount % 10 == 0 {
        g.cycleColors()
    }
    g.frameCount++
}

// Function to cycle colors
func (g *Game) cycleColors() {
    // Save the first color
    firstColor := pigo8.GetPaletteColor(0)
    
    // Shift all colors down by one
    for i := 0; i < pigo8.GetPaletteSize()-1; i++ {
        pigo8.SetPaletteColor(i, pigo8.GetPaletteColor(i+1))
    }
    
    // Put the first color at the end
    pigo8.SetPaletteColor(pigo8.GetPaletteSize()-1, firstColor)
}
```

### Day/Night Cycle Effect

```go
// Create a night-time version of the current palette
func createNightPalette() []color.Color {
    size := pigo8.GetPaletteSize()
    nightPalette := make([]color.Color, size)
    
    for i := 0; i < size; i++ {
        baseColor := pigo8.GetPaletteColor(i)
        r, g, b, _ := baseColor.RGBA()
        
        // Convert from color.Color's 16-bit per channel to 8-bit per channel
        r8 := uint8(r >> 8)
        g8 := uint8(g >> 8)
        b8 := uint8(b >> 8)
        
        // Make darker and blue-tinted (night effect)
        r8 = uint8(float64(r8) * 0.5)
        g8 = uint8(float64(g8) * 0.5)
        b8 = uint8(float64(b8) * 0.7) // Less reduction for blue = blue tint
        
        nightPalette[i] = color.RGBA{r8, g8, b8, 255}
    }
    
    return nightPalette
}
```

## Best Practices

1. **Save the original palette** before making changes if you need to restore it later.
2. **Check palette size** before accessing colors to avoid out-of-range errors.
3. **Use meaningful colors** for game elements - consider color blindness and accessibility.
4. **Be consistent** with your color usage throughout your game.
5. **Use transparency carefully** to create layering effects without overcomplicating your game.

## Related Functions

- `SetTransparency()` - Controls which colors are transparent (see [transparency.md](transparency.md))
- `Palt()` - PICO-8 style function for setting transparency (see [transparency.md](transparency.md))
