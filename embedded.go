package pigo8

import (
	"embed"
)

// defaultResources contains the default embedded resources for PIGO8
//
//go:embed resources/default_spritesheet.json resources/default_map.json
var defaultResources embed.FS

// defaultSpritesheetPath is the path to the default spritesheet in the embedded resources
const defaultSpritesheetPath = "resources/default_spritesheet.json"

// defaultMapPath is the path to the default map in the embedded resources
const defaultMapPath = "resources/default_map.json"
