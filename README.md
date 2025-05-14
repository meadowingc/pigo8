# pigo8

[![License][License-Image]][License-Url]
[![CI][CI-Image]][CI-URL]
[![Go Report Card](https://goreportcard.com/badge/drpaneas/pigo8)](https://goreportcard.com/report/drpaneas/pigo8)
[![Publish docs][Doc-Image]][Doc-URL]
[![Dependabot Updates][Dependabot-Image]][Dependabot-URL]

![pigo8_logo](pigo8.png)

PIGO8 is a Go library (package) that reimagines the creative spirit of the PICO-8 fantasy console, allowing developers to build charming, retro-style games entirely in Go.
Built on top of [Ebitengine], PIGO8 gives you a pixel-perfect environment with a familiar and fun API inspired by PICO-8, without needing to learn Lua and start your arrays counting from 1.

While PIGO8 is inspired by PICO-8, it does not contain or distribute any part of the original PICO-8 engine or its proprietary assets.
It is a clean-room reimplementation for educational and creative purposes.

## Features

* Go-first API: Design and develop retro games in idiomatic Go code.
* PICO-8 Style: Recreates the look and feel of the PICO-8 API and game loop.
* Sprite & Map Support: Import `.p8` files and reuse your graphics and map data.
* Backed by Ebiten: Harness cross-platform power with minimal setup.
* Nostalgic Aesthetic: Build games that feel right at home on a fantasy console.
* Editor: Extremely basic minimal editor for sprite and map creation.
* Extras: Custom palletes, any resolution and helpful new functions (e.g. for collision detection).
* Sound System: Audio player plays any wav file.
* Advanced Palette: Dynamic color palette management with alpha transparency support.
* Fade System: Frame-by-frame fade transitions for smooth scene changes and effects.
* Enhanced Sprite Editing: Multi-sprite editing capabilities with different grid sizes.

### Basic Game Structure

```go
package main

import p8 "github.com/drpaneas/pigo8"

// Game represents our game state
type Game struct {
    // Your game state variables here
}

// Init initializes the game
func (g *Game) Init() {
    // Initialization code
}

// Update updates the game state
func (g *Game) Update() {
    // Update game logic
}

// Draw draws the game
func (g *Game) Draw() {
    // Drawing code
}
```

```go
func main() {
    // Create an instance of the game
    game := &Game{}
    
    // Insert the game into PIGO8
    p8.InsertGame(game)

    // Run the game with the configured settings
    p8.Play()
}
```

or with custom settings:

```go
func main() {
    // Create an instance of the game
    game := &Game{}
    
    // Insert the game into PIGO8
    p8.InsertGame(game)
    
    // Configure settings
    settings := p8.NewSettings()
    settings.WindowTitle = "My PIGO8 Game"
    settings.ScaleFactor = 4
    settings.TargetFPS = 60
    settings.ScreenWidth = 160
    settings.ScreenHeight = 144
    
    // Run the game with the configured settings
    p8.PlayGameWith(settings)
}
```

**Note**: *If you have any of the `spritesheet.json`, `map.json`, `palette.json` or any `music*.wav` files in the same directory as your game, PIGO8 can automatically load them. To do that, you need to put at the top of your `main.go` the following line:*

```go
//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
```

And run `go generate` to generate the embedded files.

## Installation

To add PIGO8 to your Go project:

```bash
go get github.com/drpaneas/pigo8
```

## Documentation

Check out the `examples` directory in the repository for complete, working examples of PIGO8 games and demos.

Comprehensive documentation can be found on the [PIGO8 Wiki](https://drpaneas.github.io/pigo8/).

### Feature Examples

* Basic Game: [examples/basic](https://github.com/drpaneas/pigo8/tree/main/examples/basic)
* Custom Resolution: [examples/customResolution](https://github.com/drpaneas/pigo8/tree/main/examples/customResolution)
* Map Layers: [examples/map_layers](https://github.com/drpaneas/pigo8/tree/main/examples/map_layers)
* Palette Management: [examples/palette](https://github.com/drpaneas/pigo8/tree/main/examples/palette)
* Transparency: [examples/transparency](https://github.com/drpaneas/pigo8/tree/main/examples/transparency)
* Fade System: [examples/fade](https://github.com/drpaneas/pigo8/tree/main/examples/fade)

## Contributing

Contributions are welcome! Whether it's fixing bugs, adding features, or improving docs—pull requests and issues are encouraged.
If you are interested in contributing to PIGO8, read about our [Contributing guide](./CONTRIBUTING.md)

## Using .p8 Files

If you have existing assets made in PICO-8, you can load them by using [parsepico].
This will read your `*.p8` cartridge and extract `spritesheet.json` and `map.json`.
By placing those two JSON files into the same directory with your PIGO8 game, they will automatically get picked up by the library.

**Note**: You must own a legitimate copy of PICO-8 to access its files.
PIGO8 does not redistribute any proprietary data or code from Lexaloffle.

## License

### PICO-8

PIGO8 is released under the [MIT License](LICENSE).
It is not affiliated with or endorsed by [Lexaloffle] Games.

[PICO-8] is a trademark of [Lexaloffle] Games.
This project is a community-made, open-source interpretation and requires that users own [PICO-8] to use any `.p8` assets.

### Font Attribution

The font used in this project, [PICO-8 wide reversed], was created by [RhythmLynx] using [FontStruct], and is licensed under the [Creative Commons CC0 1.0 Public Domain Dedication](https://creativecommons.org/publicdomain/zero/1.0/).

More info and source: PICO-8 Font on GitHub and FontStruct page.

[Ebitengine]: https://ebitengine.org/
[License-Url]: https://mit-license.org/
[License-Image]: https://img.shields.io/badge/License-MIT-blue.svg
[CI-URL]: https://github.com/drpaneas/pigo8/actions/workflows/ci.yml
[CI-Image]: https://github.com/drpaneas/pigo8/actions/workflows/ci.yml/badge.svg
[Dependabot-URL]: https://github.com/drpaneas/pigo8/actions/workflows/dependabot/dependabot-updates
[Dependabot-Image]: https://github.com/drpaneas/pigo8/actions/workflows/dependabot/dependabot-updates/badge.svg
[Doc-URL]: https://github.com/drpaneas/pigo8/actions/workflows/mdbook.yml
[Doc-Image]: https://github.com/drpaneas/pigo8/actions/workflows/mdbook.yml/badge.svg
[parsepico]: https://github.com/drpaneas/parsepico
[Lexaloffle]: https://www.lexaloffle.com/
[PICO-8]: https://www.lexaloffle.com/pico-8.php
[PICO-8 wide reversed]: https://fontstruct.com/fontstructions/show/1345445/pico-8-wide-reversed
[RhythmLynx]: https://fontstruct.com/fontstructors/1292279/rhythmlynx
[FontStruct]: https://fontstruct.com/
