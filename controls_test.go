package pigo8

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: Unit testing Btn and Btnp fully is difficult due to dependencies
// on Ebitengine's input handling and connected gamepads. These tests
// primarily focus on playerIndex validation.

func TestBtnPlayerIndexValidation(t *testing.T) {
	// We don't need to mock actual gamepad state for this validation
	// The functions should return false early if playerIndex is invalid.

	button := ButtonO // Arbitrary button

	t.Run("Valid player indices (0-7)", func(t *testing.T) {
		for i := 0; i <= 7; i++ {
			// We expect false because no gamepad is connected/mocked,
			// but the function shouldn't fail due to the index itself.
			assert.False(t, Btn(button, i), "Expected false for valid player index %d (no gamepad)", i)
		}
	})

	t.Run("Invalid player index (< 0)", func(t *testing.T) {
		assert.False(t, Btn(button, -1), "Expected false for player index -1")
	})

	t.Run("Invalid player index (> 7)", func(t *testing.T) {
		assert.False(t, Btn(button, 8), "Expected false for player index 8")
		assert.False(t, Btn(button, 99), "Expected false for player index 99")
	})

	t.Run("Default player index (0)", func(t *testing.T) {
		assert.False(t, Btn(button), "Expected false for default player index 0 (no gamepad)")
	})
}

func TestBtnpPlayerIndexValidation(t *testing.T) {
	// Similar logic to TestBtnPlayerIndexValidation
	button := ButtonX // Arbitrary button

	t.Run("Valid player indices (0-7)", func(t *testing.T) {
		for i := 0; i <= 7; i++ {
			assert.False(t, Btnp(button, i), "Expected false for valid player index %d (no gamepad)", i)
		}
	})

	t.Run("Invalid player index (< 0)", func(t *testing.T) {
		assert.False(t, Btnp(button, -1), "Expected false for player index -1")
	})

	t.Run("Invalid player index (> 7)", func(t *testing.T) {
		assert.False(t, Btnp(button, 8), "Expected false for player index 8")
		assert.False(t, Btnp(button, 99), "Expected false for player index 99")
	})

	t.Run("Default player index (0)", func(t *testing.T) {
		assert.False(t, Btnp(button), "Expected false for default player index 0 (no gamepad)")
	})
}
