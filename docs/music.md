# Using Music in PIGO8

PIGO8 provides audio playback capabilities that allow you to incorporate music from PICO-8 into your game.

## Exporting Music from PICO-8

To use music in your PIGO8 application, you first need to export the audio files from PICO-8:

1. In PICO-8, create your music using the sound editor and music tracker
2. Export each music pattern using the `export` command in the PICO-8 console:

   ```
   export music0.wav
   export music1.wav
   export music2.wav
   ... etc ...
   ```

   or simply do:

   ```
   export music%d.wav
   ```

   Notice this will export all music patterns from 0 to 63, regardless if you have actually created them or not.
   That means you will get a lot of empty files, but don't worry, the `embedgen` tool will only embed valid music files.

PICO-8 will save these files in its current working directory.
You'll need to copy these files to your PIGO8 project directory, placing them next to your `main.go` file.

## Setting Up Your PIGO8 Project

To use music in your PIGO8 project:

1. Copy the exported `.wav` files to your project directory
2. Add this specific `go:generate` directive to your `main.go` file without changing it:

   ```go
   //go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .
   ```

3. Run `go generate` in your project directory to embed the music files

The `embedgen` tool will automatically detect and embed valid music files. It analyzes each WAV file to ensure it contains actual audio data and is not silent or corrupted.

## Playing Music in Your Game

PIGO8 provides the `Music()` function to play audio files:

```go
// Play music track 0, meaning music0.wav
p8.Music(0)

// Play music track 3 exclusively (stops any currently playing music)
p8.Music(3, true)

// Stop all music
p8.Music(-1)
```

### Function Parameters

- `n` (int): The music track number to play (0-63), or -1 to stop all music
- `exclusive` (bool, optional): If true, stops any currently playing music before playing the new track

## Example Usage

Here's a simple example of using music in a PIGO8 game:

```go
package main

//go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .

import (
    p8 "github.com/drpaneas/pigo8"
)

type Game struct {
    // game state
}

func (g *Game) Init() {
    // Play background music when the game starts
    p8.Music(0)
}

func (g *Game) Update() {
    // Play different music when a key is pressed
    if p8.Btn(p8.ButtonUp) {
        p8.Music(1)
    }
    
    // Play exclusive music (stops other tracks) when DOWN is pressed
    if p8.Btn(p8.ButtonDown) {
        p8.Music(2, true)
    }
    
    // Stop all music when LEFT+RIGHT are pressed together
    if p8.Btn(p8.ButtonLeft) && p8.Btn(p8.ButtonRight) {
        p8.Music(-1)
    }
}
```

## Audio File Validation

The PIGO8 embedgen tool performs validation on audio files to ensure they contain actual audio data:

1. Checks for valid WAV file format (RIFF and WAVE markers)
2. Verifies the file has a non-zero data chunk
3. Analyzes a sample of the audio data to ensure it's not silent

Only valid audio files will be included in your application.
