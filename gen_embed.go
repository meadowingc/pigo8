//go:build exec

package main

import (
	"fmt"
	"os"
)

func main() {
	// Check if map.json and spritesheet.json exist in the current directory
	mapExists := fileExists("map.json")
	spritesheetExists := fileExists("spritesheet.json")

	if !mapExists && !spritesheetExists {
		fmt.Println("Warning: Neither map.json nor spritesheet.json found in current directory.")
		return
	}

	// Generate the embed.go file
	content := `package main

import (
	"embed"
	
	p8 "github.com/drpaneas/pigo8"
)

// Embed the game-specific resources
//
`
	// Only include files that exist
	embedDirective := "//go:embed"
	if mapExists {
		embedDirective += " map.json"
	}
	if spritesheetExists {
		embedDirective += " spritesheet.json"
	}
	content += embedDirective + "\n"
	content += `var resources embed.FS

func init() {
	// Register the embedded resources with PIGO8
	p8.RegisterEmbeddedResources(resources, `

	// Add the correct paths based on what exists
	content += `"`
	if spritesheetExists {
		content += "spritesheet.json"
	}
	content += `", "`
	if mapExists {
		content += "map.json"
	}
	content += `")`

	content += `
}
`
	// Write the file
	err := os.WriteFile("embed.go", []byte(content), 0644)
	if err != nil {
		fmt.Printf("Error generating embed.go: %v\n", err)
		return
	}

	fmt.Println("Generated embed.go for PIGO8 resources")
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
