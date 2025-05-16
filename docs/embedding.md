# Resource Embedding in PIGO8

## Quick Start Guide

**To create a portable PIGO8 game that works anywhere:**

1. Add this line at the top of your main.go file:

   ```go
   //go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
   ```

2. Run these commands before distributing your game:

   ```bash
   go generate
   go build
   ```

That's it! Your game binary will now include all necessary resources and work correctly even when moved to a different directory.

## What This Does

PIGO8 uses the following resource files for games:

- `map.json` - Contains the game map data
- `spritesheet.json` - Contains sprite definitions and pixel data
- `palette.hex` - Contains custom color palette definitions

The `go generate` command automatically creates an `embed.go` file that embeds these resources into your binary, making your game fully portable.

## How Resource Loading Works

PIGO8 uses a smart resource loading system with the following priority order:

1. Files in the current directory (highest priority)
2. Common subdirectories: `assets/`, `resources/`, `data/`, `static/`
3. Embedded resources registered via `RegisterEmbeddedResources`
4. Default embedded resources in the PIGO8 library (lowest priority)

This approach gives you the best of both worlds:

- During development: Edit local files for quick iteration
- For distribution: Embed resources for portability

## Detailed Usage Guide

### Automatic Embedding with go:generate (Recommended)

PIGO8 provides a tool that automatically generates the necessary embedding code for your game. This is the recommended approach for distributing your game.

1. Add this line at the top of your main.go file:

   ```go
   //go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
   ```

2. Run the generate command to create the embed.go file:

   ```bash
   go generate
   ```

3. Build your game normally:

   ```bash
   go build
   ```

The generated `embed.go` file will embed your map.json, spritesheet.json, and palette.hex files into the binary. Your game will now work correctly even when moved to a different directory.

### Manual Embedding (Alternative)

If you prefer to manually control the embedding process, you can create an `embed.go` file in your project:

```go
package main

import (
 "embed"
 
 p8 "github.com/drpaneas/pigo8"
)

// Embed the game-specific resources
//
//go:embed map.json spritesheet.json palette.hex
var resources embed.FS

func init() {
 // Register the embedded resources with PIGO8
 // Audio will be automatically initialized if audio files are present
 p8.RegisterEmbeddedResources(resources, "spritesheet.json", "map.json", "palette.hex")
}
```

Adjust the `go:embed` directive to include only the files you have. For example, if you only have a palette.hex file, your embed.go would look like:

```go
package main

import (
 "embed"
 
 p8 "github.com/drpaneas/pigo8"
)

// Embed the game-specific resources
//
//go:embed palette.hex
var resources embed.FS

func init() {
 // Register the embedded resources with PIGO8
 // Audio will be automatically initialized if audio files are present
 p8.RegisterEmbeddedResources(resources, "", "", "palette.hex")
}
```

## Custom Color Palettes with palette.hex

PIGO8 now supports custom color palettes through a `palette.hex` file. This allows you to use color palettes from sites like [Lospec](https://lospec.com/palette-list) in your games.

### Creating a palette.hex File

1. Visit [Lospec Palette List](https://lospec.com/palette-list) and find a palette you like
2. Download the palette in HEX format
3. Save it as `palette.hex` in your game directory

Each line in the palette.hex file should contain a single hex color code, for example:

```
c60021
e70000
e76121
e7a263
e7c384
```

### How Palette Loading Works

When a palette.hex file is loaded:

1. The first color (index 0) is automatically set to be fully transparent (rgba(0, 0, 0, 0))
2. All colors from the palette.hex file are shifted up by one index
3. The palette can be used like any other PIGO8 palette

### Example Usage

See the `examples/palette_hex` directory for a complete example of loading and using a custom palette.

## Resource Loading Priority

PIGO8 uses the following priority order when looking for resources:

1. Files in the current directory (highest priority)
2. Custom embedded resources registered via `RegisterEmbeddedResources`
3. Default embedded resources in the PIGO8 library (lowest priority)

This allows you to:

- Develop with local files for quick iteration
- Distribute with embedded resources for portability
- Always have fallback resources from the library
