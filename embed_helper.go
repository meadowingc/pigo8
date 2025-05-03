package pigo8

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
)

// EmbeddedResources represents a set of embedded resources for a PIGO8 game
type EmbeddedResources struct {
	FS              fs.FS
	SpritesheetPath string
	MapPath         string
}

// CustomResources holds application-specific embedded resources
var CustomResources *EmbeddedResources

// autoDetectResourcesAttempted tracks whether we've already tried to auto-detect resources
var autoDetectResourcesAttempted bool

// RegisterEmbeddedResources allows applications to register their own embedded resources
// This should be called before any PIGO8 functions that might need these resources
func RegisterEmbeddedResources(resources fs.FS, spritesheetPath, mapPath string) {
	CustomResources = &EmbeddedResources{
		FS:              resources,
		SpritesheetPath: spritesheetPath,
		MapPath:         mapPath,
	}
	log.Printf("Registered custom embedded resources: spritesheet=%s, map=%s", spritesheetPath, mapPath)
}

// tryLoadEmbeddedFile attempts to load a file from embedded resources
// It first tries custom resources, then falls back to default resources
// Returns the file content and a boolean indicating success
func tryLoadEmbeddedFile(defaultPath string, isMap bool) ([]byte, bool) {
	// Auto-detect resources if not already done
	if !autoDetectResourcesAttempted {
		autoDetectResources()
	}

	// First try custom resources if registered
	if CustomResources != nil {
		var path string
		if isMap {
			path = CustomResources.MapPath
		} else {
			path = CustomResources.SpritesheetPath
		}

		data, err := fs.ReadFile(CustomResources.FS, path)
		if err == nil {
			fileType := "spritesheet"
			if isMap {
				fileType = "map"
			}
			log.Printf("Using custom embedded %s file: %s", fileType, path)
			return data, true
		}
	}

	// Fall back to default resources
	data, err := DefaultResources.ReadFile(defaultPath)
	if err != nil {
		return nil, false
	}

	fileType := "spritesheet"
	if isMap {
		fileType = "map"
	}
	log.Printf("Using default embedded %s file: %s", fileType, defaultPath)
	return data, true
}

// autoDetectResources attempts to find common resource locations
// and automatically register them if found
func autoDetectResources() {
	// Only try once
	autoDetectResourcesAttempted = true

	// Don't override if already registered
	if CustomResources != nil {
		return
	}

	// Get information about the current build
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	// Check if we're running in a Go module
	hasModule := false
	for _, setting := range buildInfo.Settings {
		if setting.Key == "GOMOD" && setting.Value != "" {
			hasModule = true
			break
		}
	}

	// Log module detection status
	if hasModule {
		log.Printf("Running in a Go module, checking for resources")
	}

	// Check common resource locations
	resourceDirs := []string{
		".", // Current directory
		"assets",
		"resources",
		"data",
		"static",
	}

	// Try each location
	for _, dir := range resourceDirs {
		mapPath := filepath.Join(dir, "map.json")
		spritesheetPath := filepath.Join(dir, "spritesheet.json")

		mapExists := fileExists(mapPath)
		spritesheetExists := fileExists(spritesheetPath)

		// If either resource exists, try to create an embedded FS
		if mapExists || spritesheetExists {
			log.Printf("Found resources in directory: %s", dir)
			// We can't create an embed.FS at runtime, but we can note the locations
			// for future file operations
			break
		}
	}
}

// fileExists checks if a file exists and is not a directory
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil && !info.IsDir()
}

// tryLoadEmbeddedMap attempts to load a map from embedded resources
func tryLoadEmbeddedMap() ([]byte, error) {
	data, ok := tryLoadEmbeddedFile(DefaultMapPath, true)
	if !ok {
		return nil, fmt.Errorf("failed to load embedded map file")
	}
	return data, nil
}

// tryLoadEmbeddedSpritesheet attempts to load a spritesheet from embedded resources
func tryLoadEmbeddedSpritesheet() ([]byte, error) {
	data, ok := tryLoadEmbeddedFile(DefaultSpritesheetPath, false)
	if !ok {
		return nil, fmt.Errorf("failed to load embedded spritesheet file")
	}
	return data, nil
}
