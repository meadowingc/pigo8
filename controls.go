// Package pigo8 package provides a set of functions to handle input for the PICO-8 fantasy console.
package pigo8

import (
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PICO-8 Button Index Constants
const (
	ButtonLeft = iota
	ButtonRight
	ButtonUp
	ButtonDown
	ButtonO // Often mapped to A/Cross on standard controllers
	ButtonX // Often mapped to B/Circle on standard controllers
	ButtonStart
	ButtonSelect
)

// pico8ButtonToStandard maps PICO-8 button indices to Ebitengine Standard Gamepad Buttons.
// This assumes a standard layout mapping common on many controllers.
var pico8ButtonToStandard = map[int]ebiten.StandardGamepadButton{
	ButtonLeft:   ebiten.StandardGamepadButtonLeftLeft,
	ButtonRight:  ebiten.StandardGamepadButtonLeftRight,
	ButtonUp:     ebiten.StandardGamepadButtonLeftTop,
	ButtonDown:   ebiten.StandardGamepadButtonLeftBottom,
	ButtonO:      ebiten.StandardGamepadButtonRightBottom, // A / Cross
	ButtonX:      ebiten.StandardGamepadButtonRightLeft,   // B / Circle
	ButtonStart:  ebiten.StandardGamepadButtonCenterRight, // Start / Options
	ButtonSelect: ebiten.StandardGamepadButtonCenterLeft,  // Select / Back / Share
}

// pico8ButtonToKeyboardP0 maps PICO-8 button indices to default keyboard keys for Player 0.
var pico8ButtonToKeyboardP0 = map[int]ebiten.Key{
	ButtonLeft:   ebiten.KeyLeft,
	ButtonRight:  ebiten.KeyRight,
	ButtonUp:     ebiten.KeyUp,
	ButtonDown:   ebiten.KeyDown,
	ButtonO:      ebiten.KeyZ,          // PICO-8 O button ('Z' key)
	ButtonX:      ebiten.KeyX,          // PICO-8 X button ('X' key)
	ButtonStart:  ebiten.KeyEnter,      // Often mapped to Enter/Return
	ButtonSelect: ebiten.KeyShiftRight, // Often mapped to Shift (Right)
}

// connectedGamepadIDs stores the currently connected gamepad IDs.
// Use a map for efficient add/remove operations.
var connectedGamepadIDs = make(map[ebiten.GamepadID]struct{})

// gamepadIDsBuf is a temporary buffer reused by UpdateConnectedGamepads.
var gamepadIDsBuf []ebiten.GamepadID

// UpdateConnectedGamepads refreshes the list of connected gamepad IDs.
// Call this function once per frame in your game's Update method.
func UpdateConnectedGamepads() {
	// Check for newly connected gamepads
	gamepadIDsBuf = inpututil.AppendJustConnectedGamepadIDs(gamepadIDsBuf[:0])
	for _, id := range gamepadIDsBuf {
		connectedGamepadIDs[id] = struct{}{}
	}

	// Check for disconnected gamepads
	for id := range connectedGamepadIDs {
		if inpututil.IsGamepadJustDisconnected(id) {
			delete(connectedGamepadIDs, id)
		}
	}
}

// getGamepadID retrieves the Ebitengine GamepadID for a given PICO-8 player index (0-7).
// It uses the map of connected gamepads updated by UpdateConnectedGamepads.
// Returns the ID and true if found, otherwise returns 0 and false.
func getGamepadID(playerIndex int) (ebiten.GamepadID, bool) {
	// Extract the current IDs from the map into a slice
	currentIDs := make([]ebiten.GamepadID, 0, len(connectedGamepadIDs))
	for id := range connectedGamepadIDs {
		currentIDs = append(currentIDs, id)
	}

	// Sort the IDs numerically to ensure consistent player mapping
	sort.Slice(currentIDs, func(i, j int) bool {
		return currentIDs[i] < currentIDs[j]
	})

	if playerIndex < 0 || playerIndex >= len(currentIDs) {
		// Invalid player index or no gamepad connected for this index
		return 0, false
	}

	// Return the ID corresponding to the player index from the sorted list
	return currentIDs[playerIndex], true
}

// Btn checks if a specific PICO-8 button is currently held down via gamepad OR keyboard (Player 0 only).
// Mimics the PICO-8 btn() function behavior (returns true while held).
//
// buttonIndex: The PICO-8 button index (0-7 or constants like PICO8_BUTTON_LEFT).
// playerIndex: Optional PICO-8 player index (0-7). Defaults to 0 (player 1) if omitted.
//
//	Keyboard input is only checked for playerIndex 0.
//
// Usage:
//
//	Btn(buttonIndex)
//	Btn(buttonIndex, playerIndex)
//
// Example:
//
//	// Check if the left button/arrow key is held for player 0
//	if Btn(PICO8_BUTTON_LEFT) {
//		// Move left
//	}
//
//	// Check if the 'O' button (gamepad only) is held for player 1 (index 1)
//	if Btn(PICO8_BUTTON_O, 1) {
//		// Player 1 action
//	}
func Btn(buttonIndex int, playerIndex ...int) bool {
	pIdx := 0 // Default to player 0
	if len(playerIndex) > 0 {
		pIdx = playerIndex[0]
	}

	// --- Keyboard Check (Player 0 Only) ---
	keyboardPressed := false
	if pIdx == 0 {
		if key, ok := pico8ButtonToKeyboardP0[buttonIndex]; ok {
			keyboardPressed = ebiten.IsKeyPressed(key)
		}
	}

	// --- Gamepad Check ---
	gamepadPressed := false
	// Validate PICO-8 player index range (0-7) for gamepad
	if pIdx >= 0 && pIdx <= 7 {
		// Get the corresponding Ebitengine GamepadID for the player index
		gamepadID, ok := getGamepadID(pIdx)
		// Check if the standard button mapping exists
		standardButton, mappingExists := pico8ButtonToStandard[buttonIndex]

		// Only proceed if gamepad connected, button is mapped, and layout/button available
		if ok && mappingExists &&
			ebiten.IsStandardGamepadLayoutAvailable(gamepadID) &&
			ebiten.IsStandardGamepadButtonAvailable(gamepadID, standardButton) {
			gamepadPressed = ebiten.IsStandardGamepadButtonPressed(gamepadID, standardButton)
		}
	}

	return keyboardPressed || gamepadPressed
}

// Note: For "just pressed" behavior similar to PICO-8's btnp(), you would use
// inpututil functions.

// Btnp checks if a specific PICO-8 button was just pressed via gamepad OR keyboard (Player 0 only).
// Mimics the PICO-8 btnp() function behavior (without auto-repeat).
// It returns true only on the single frame the button transitions from up to down.
//
// buttonIndex: The PICO-8 button index (0-7 or constants like PICO8_BUTTON_LEFT).
// playerIndex: Optional PICO-8 player index (0-7). Defaults to 0 (player 1) if omitted.
//
//	Keyboard input is only checked for playerIndex 0.
//
// Usage:
//
//	Btnp(buttonIndex)
//	Btnp(buttonIndex, playerIndex)
//
// Example:
//
//	// Check if the 'X' button/key was just pressed for player 0
//	if Btnp(PICO8_BUTTON_X) {
//		// Jump action
//	}
//
//	// Check if the 'Start' button (gamepad only) was just pressed for player 1
//	if Btnp(PICO8_BUTTON_START, 1) {
//		// Pause game for player 1
//	}
func Btnp(buttonIndex int, playerIndex ...int) bool {
	pIdx := 0 // Default to player 0
	if len(playerIndex) > 0 {
		pIdx = playerIndex[0]
	}

	// --- Keyboard Check (Player 0 Only) ---
	keyboardJustPressed := false
	if pIdx == 0 {
		if key, ok := pico8ButtonToKeyboardP0[buttonIndex]; ok {
			keyboardJustPressed = inpututil.IsKeyJustPressed(key)
		}
	}

	// --- Gamepad Check ---
	gamepadJustPressed := false
	// Validate PICO-8 player index range (0-7) for gamepad
	if pIdx >= 0 && pIdx <= 7 {
		// Get the corresponding Ebitengine GamepadID
		gamepadID, ok := getGamepadID(pIdx)
		// Map to standard button
		standardButton, mappingExists := pico8ButtonToStandard[buttonIndex]

		// Only proceed if gamepad connected, button is mapped, and layout/button available
		if ok && mappingExists &&
			ebiten.IsStandardGamepadLayoutAvailable(gamepadID) &&
			ebiten.IsStandardGamepadButtonAvailable(gamepadID, standardButton) {
			gamepadJustPressed = inpututil.IsStandardGamepadButtonJustPressed(gamepadID, standardButton)
		}
	}

	return keyboardJustPressed || gamepadJustPressed
}
