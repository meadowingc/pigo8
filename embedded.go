package pigo8

import (
	"embed"
)

// DefaultResources contains the default embedded resources for PIGO8
//
//go:embed resources/default_spritesheet.json resources/default_map.json
var DefaultResources embed.FS

// DefaultSpritesheetPath is the path to the default spritesheet in the embedded resources
const DefaultSpritesheetPath = "resources/default_spritesheet.json"

// DefaultMapPath is the path to the default map in the embedded resources
const DefaultMapPath = "resources/default_map.json"

// DefaultAudioPathPrefix is the prefix for default audio files in the embedded resources
const DefaultAudioPathPrefix = "resources/default_audio"
