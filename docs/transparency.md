# PIGO8 Transparency Management

PIGO8 uses binary transparency, where each color in the palette is either fully visible or fully transparent. By default, only color 0 (black) is transparent. You can change which colors are transparent to create various visual effects.

## Function

### Palt

```go
func Palt(args ...interface{})
```

PICO-8 style function for setting color transparency. This function has multiple usage patterns:

1. `Palt()` - Reset all transparency settings to default (only color 0 is transparent)
2. `Palt(colorIndex, transparent)` - Set a specific color's transparency

**Parameters:**

- `colorIndex`: Index of the color in the palette.
- `transparent`: Whether the color should be transparent (`true`) or opaque (`false`).

**Example:**

```go
// Make color 0 opaque and color 1 transparent
pigo8.Palt(0, false)
pigo8.Palt(1, true)

// Reset to default (only color 0 is transparent)
pigo8.Palt()
```

## How Transparency Works

When drawing sprites, pixels with transparent colors are not drawn, allowing the background to show through. This is useful for:

1. Creating sprites with irregular shapes
2. Layering sprites on top of each other
3. Creating special effects

The transparency is checked in drawing functions like `Spr()`, `Sspr()`, and `Pset()`.

## Advanced Usage Examples

### Creating a Sprite Mask

```go
// Draw a background
pigo8.Rectfill(0, 0, 127, 127, 12) // Fill screen with blue

// Set red (color 8) as transparent
pigo8.Palt(8, true)

// Draw a sprite where all red pixels will be transparent
pigo8.Spr(0, 64, 64, 1, 1)

// Reset transparency to default
pigo8.Palt()
```

### Multiple Transparent Colors

```go
// Make both black (0) and white (7) transparent
pigo8.Palt(0, true)
pigo8.Palt(7, true)

// Draw sprite with both black and white areas transparent
pigo8.Spr(1, 10, 10, 1, 1)

// Reset to default
pigo8.Palt()
```

### Swapping Transparent Colors

```go
// Make black (0) opaque and blue (12) transparent
pigo8.Palt(0, false)
pigo8.Palt(12, true)

// Draw sprites with this new transparency setting
pigo8.Spr(2, 20, 20, 1, 1)

// Reset to default
pigo8.Palt()
```

### Creating a Cutout Effect

```go
// Draw a colorful background
for y := 0; y < 128; y += 8 {
    for x := 0; x < 128; x += 8 {
        pigo8.Rectfill(x, y, x+7, y+7, (x+y) % pigo8.GetPaletteSize())
    }
}

// Make white (7) transparent
pigo8.Palt(7, true)

// Draw a sprite with white areas that act as "windows" to the background
pigo8.Spr(3, 32, 32, 2, 2)
```

## Best Practices

1. **Reset transparency** when you're done with special effects to avoid unexpected behavior.
2. **Be consistent** with your transparency usage to avoid confusion.
3. **Document your transparency choices** in your code with comments.
4. **Consider performance** - using many transparent colors can make your code harder to understand.
5. **Test thoroughly** - transparency effects can sometimes be subtle and may not work as expected on all backgrounds.

## Transparency and Custom Palettes

When you use `SetPalette()` to change the palette, the transparency array is automatically resized to match. The function preserves existing transparency settings for colors that still exist in the new palette and ensures that color 0 is transparent by default.

```go
// Create a custom palette
customPalette := []color.Color{
    color.RGBA{0, 0, 0, 255},     // Black
    color.RGBA{255, 0, 0, 255},   // Red
    color.RGBA{0, 255, 0, 255},   // Green
    color.RGBA{0, 0, 255, 255},   // Blue
}

// Set the palette - color 0 will be transparent by default
pigo8.SetPalette(customPalette)

// Make red (color 1) transparent as well
pigo8.Palt(1, true)
```

## Related Functions

- `SetPalette()` - Replace the entire palette (see [palette-management.md](palette-management.md))
- `GetPaletteSize()` - Get the number of colors in the palette (see [palette-management.md](palette-management.md))
