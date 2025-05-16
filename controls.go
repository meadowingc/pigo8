// Package pigo8 package provides a set of functions to handle input for the PICO-8 fantasy console.
package pigo8

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PICO-8 Button Index Constants
const (
	LEFT = iota
	RIGHT
	UP
	DOWN
	O // Often mapped to A/Cross on standard controllers
	X // Often mapped to B/Circle on standard controllers
	START
	SELECT
	// Mouse buttons
	MouseLeft
	MouseRight
	MouseMiddle // Mouse wheel press
	MouseWheelUp
	MouseWheelDown
)

// pico8ButtonToStandard maps PICO-8 button indices to Ebitengine Standard Gamepad Buttons.
// Mapped specifically for Xbox controller layout.
var pico8ButtonToStandard = map[int]ebiten.StandardGamepadButton{
	// D-Pad (on Xbox, these are the actual D-pad buttons, not the left stick)
	LEFT:  ebiten.StandardGamepadButtonLeftLeft,
	RIGHT: ebiten.StandardGamepadButtonLeftRight,
	UP:    ebiten.StandardGamepadButtonLeftTop,
	DOWN:  ebiten.StandardGamepadButtonLeftBottom,

	// Face buttons (A/B/X/Y on Xbox)
	O: ebiten.StandardGamepadButtonRightLeft,   // X button on Xbox (left face button)
	X: ebiten.StandardGamepadButtonRightBottom, // A button on Xbox (bottom face button)

	// Start/Select (Menu/View on Xbox)
	START:  ebiten.StandardGamepadButtonCenterRight, // Start/Menu button
	SELECT: ebiten.StandardGamepadButtonCenterLeft,  // Select/View button
}

// pico8ButtonToKeyboardP0 maps PICO-8 button indices to default keyboard keys for Player 0.
var pico8ButtonToKeyboardP0 = map[int]ebiten.Key{
	LEFT:   ebiten.KeyLeft,
	RIGHT:  ebiten.KeyRight,
	UP:     ebiten.KeyUp,
	DOWN:   ebiten.KeyDown,
	O:      ebiten.KeyZ,          // PICO-8 O button ('Z' key)
	X:      ebiten.KeyX,          // PICO-8 X button ('X' key)
	START:  ebiten.KeyEnter,      // Often mapped to Enter/Return
	SELECT: ebiten.KeyShiftRight, // Often mapped to Shift (Right)
}

// connectedGamepadIDs stores the currently connected gamepad IDs.
// Use a map for efficient add/remove operations.
var connectedGamepadIDs = make(map[ebiten.GamepadID]struct{})

// gamepadIDsBuf is a temporary buffer reused by UpdateConnectedGamepads.
var gamepadIDsBuf []ebiten.GamepadID

// updateConnectedGamepads refreshes the list of connected gamepad IDs.
// Call this function once per frame in your game's Update method.
func updateConnectedGamepads() {
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

// // getGamepadID retrieves the Ebitengine GamepadID for a given PICO-8 player index (0-7).
// // It uses the map of connected gamepads updated by UpdateConnectedGamepads.
// // Returns the ID and true if found, otherwise returns 0 and false.
// func getGamepadID(playerIndex int) (ebiten.GamepadID, bool) {
// 	// Extract the current IDs from the map into a slice
// 	currentIDs := make([]ebiten.GamepadID, 0, len(connectedGamepadIDs))
// 	for id := range connectedGamepadIDs {
// 		currentIDs = append(currentIDs, id)
// 	}

// 	// Sort the IDs numerically to ensure consistent player mapping
// 	sort.Slice(currentIDs, func(i, j int) bool {
// 		return currentIDs[i] < currentIDs[j]
// 	})

// 	if playerIndex < 0 || playerIndex >= len(currentIDs) {
// 		// Invalid player index or no gamepad connected for this index
// 		return 0, false
// 	}

// 	// Return the ID corresponding to the player index from the sorted list
// 	return currentIDs[playerIndex], true
// }

// Btn checks if a specific PICO-8 button is currently held down via gamepad, keyboard (Player 0 only), or mouse.
// Mimics the PICO-8 btn() function behavior (returns true while held).
//
// buttonIndex: The PICO-8 button index (0-15).
// playerIndex: Optional PICO-8 player index (0-7). Defaults to 0 (player 1) if omitted.
func Btn(buttonIndex int, playerIndex ...int) bool {
	pIdx := 0 // Default to player 0
	if len(playerIndex) > 0 {
		pIdx = playerIndex[0]
	}

	// --- Mouse Check ---
	switch buttonIndex {
	case MouseLeft:
		return ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	case MouseRight:
		return ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	case MouseMiddle:
		return ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle)
	case MouseWheelUp:
		_, wheelY := ebiten.Wheel()
		return wheelY < 0
	case MouseWheelDown:
		_, wheelY := ebiten.Wheel()
		return wheelY > 0
	}

	// --- Keyboard Check (Player 0 Only) ---
	if pIdx == 0 {
		if key, ok := pico8ButtonToKeyboardP0[buttonIndex]; ok {
			if ebiten.IsKeyPressed(key) {
				return true
			}
		}
	}

	// --- Gamepad Check ---
	// Get the first connected gamepad for this player (simplified version)
	ids := ebiten.AppendGamepadIDs(nil)
	if pIdx < 0 || pIdx >= len(ids) {
		return false
	}
	gamepadID := ids[pIdx]

	// Check if we have a standard mapping for this button
	if standardButton, ok := pico8ButtonToStandard[buttonIndex]; ok {
		// First try standard gamepad layout
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			return ebiten.IsStandardGamepadButtonPressed(gamepadID, standardButton)
		}

		// Fallback to direct button mapping for non-standard gamepads
		// This is a simple fallback and might need adjustment
		switch buttonIndex {
		case LEFT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftLeft))
		case RIGHT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftRight))
		case UP:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftTop))
		case DOWN:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftBottom))
		case O: // B button on Xbox (right face button)
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton1) // B button
		case X: // A button on Xbox (bottom face button)
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton0) // A button
		case START:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton9)
		case SELECT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton8)
		}
	}

	return false
}

// Note: For "just pressed" behavior similar to PICO-8's btnp(), you would use
// inpututil functions.

// Btnp checks if a specific PICO-8 button was just pressed via gamepad, keyboard (Player 0 only), or mouse.
// Mimics the PICO-8 btnp() function behavior (without auto-repeat).
// It returns true only on the single frame the button transitions from up to down.
//
// buttonIndex: The PICO-8 button index (0-15).
// playerIndex: Optional PICO-8 player index (0-7). Defaults to 0 (player 1) if omitted.
//
//	Keyboard input is only checked for playerIndex 0.
//	Mouse input is available for all player indices.
//
// Usage:
//
//	Btnp(buttonIndex)
//	Btnp(buttonIndex, playerIndex)
//
// Example:
//
//	// Check if the 'X' button/key was just pressed for player 0
//	if Btnp(X) {
//		// Jump action
//	}
//
//	// Check if the right mouse button was just pressed
//	if Btnp(MOUSE_RIGHT) {
//		// Handle right click
//	}
//
//	// Check if the 'Start' button (gamepad only) was just pressed for player 1
//	if Btnp(START, 1) {
//		// Pause game for player 1
//	}
func Btnp(buttonIndex int, playerIndex ...int) bool {
	pIdx := 0 // Default to player 0
	if len(playerIndex) > 0 {
		pIdx = playerIndex[0]
	}

	// --- Mouse Check ---
	switch buttonIndex {
	case MouseLeft:
		return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	case MouseRight:
		return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
	case MouseMiddle:
		return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle)
	case MouseWheelUp:
		_, wheelY := ebiten.Wheel()
		return wheelY < 0
	case MouseWheelDown:
		_, wheelY := ebiten.Wheel()
		return wheelY > 0
	}

	// --- Keyboard Check (Player 0 Only) ---
	if pIdx == 0 {
		if key, ok := pico8ButtonToKeyboardP0[buttonIndex]; ok {
			if inpututil.IsKeyJustPressed(key) {
				return true
			}
		}
	}

	// --- Gamepad Check ---
	// Get the first connected gamepad for this player (simplified version)
	ids := ebiten.AppendGamepadIDs(nil)
	if pIdx < 0 || pIdx >= len(ids) {
		return false
	}
	gamepadID := ids[pIdx]

	// Check if we have a standard mapping for this button
	if standardButton, ok := pico8ButtonToStandard[buttonIndex]; ok {
		// First try standard gamepad layout
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			return inpututil.IsStandardGamepadButtonJustPressed(gamepadID, standardButton)
		}

		// Fallback to direct button mapping for non-standard gamepads
		switch buttonIndex {
		case LEFT:
			return ebiten.GamepadAxisValue(gamepadID, 0) < -0.5 && inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftLeft))
		case RIGHT:
			return ebiten.GamepadAxisValue(gamepadID, 0) > 0.5 && ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftRight))
		case UP:
			return ebiten.GamepadAxisValue(gamepadID, 1) < -0.5 && ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftTop))
		case DOWN:
			return ebiten.GamepadAxisValue(gamepadID, 1) > 0.5 && ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftBottom))
		case O: // B button on Xbox (right face button)
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton1) // B button
		case X: // A button on Xbox (bottom face button)
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton0) // A button
		case START:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton9)
		case SELECT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton8)
		}
	}

	return false
}
