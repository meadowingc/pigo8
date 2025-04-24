package pigo8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSettings(t *testing.T) {
	settings := NewSettings()
	assert.NotNil(t, settings)
	// Verify default values match those specified in NewSettings
	assert.Equal(t, 4, settings.ScaleFactor) // Default is 4
	assert.Equal(t, "PIGO-8 Game", settings.WindowTitle)
	assert.Equal(t, 30, settings.TargetFPS)
}

// --- Add tests for PlayGameWith, InsertGame etc. if needed ---
// (Though these often require more integration-style testing)
