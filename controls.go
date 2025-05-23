// Package pigo8 package provides a set of functions to handle input for the PICO-8 fantasy console.
package pigo8

import (
	"os/exec"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// PICO-8 Button Index Constants
const (
	// Directional buttons (keyboard and gamepad)
	LEFT = iota
	RIGHT
	UP
	DOWN
	// Face buttons (keyboard and gamepad)
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
	// Gamepad-specific buttons (for direct mapping)
	joyUp
	joyDown
	joyLeft
	joyRight
	joyA
	joyB
	joyX
	joyY
	// Shoulder buttons and triggers
	l1 // Left shoulder button
	r1 // Right shoulder button
	l2 // Left trigger (analog)
	r2 // Right trigger (analog)
	// Stick clicks
	l3 // Left stick click
	r3 // Right stick click
	// Additional Steam Deck back buttons
	l4 // Back left button 1 (Steam Deck)
	r4 // Back right button 1 (Steam Deck)
	l5 // Back left button 2 (Steam Deck)
	r5 // Back right button 2 (Steam Deck)
	// Alias for pause button (same as START)
	PAUSE = START
)

// isSteamDeck checks if the game is running on a Steam Deck by checking the hostname
func isSteamDeck() bool {
	// Execute uname --nodename command to get the hostname
	cmd := exec.Command("uname", "--nodename")
	output, err := cmd.Output()
	if err != nil {
		// Command failed, likely not a Linux system or uname not available
		return false
	}

	// Trim whitespace and check if the output is exactly "steamdeck"
	hostname := strings.TrimSpace(string(output))
	return hostname == "steamdeck"
}

// pico8ButtonToStandard maps PICO-8 button indices to Ebitengine Standard Gamepad Buttons.
var pico8ButtonToStandard map[int]ebiten.StandardGamepadButton

// initButtonMappings initializes the button and axis mappings based on the current platform
func initButtonMappings() {
	// Initialize with common button mappings
	pico8ButtonToStandard = map[int]ebiten.StandardGamepadButton{
		// Face buttons (Xbox-style: A=bottom, B=right, X=left, Y=top)
		O: ebiten.StandardGamepadButtonRightLeft,   // Left button (X on Xbox, Square on PlayStation)
		X: ebiten.StandardGamepadButtonRightBottom, // Bottom button (A on Xbox, X on PlayStation)

		// Menu buttons
		START:  ebiten.StandardGamepadButtonCenterRight, // Start button
		SELECT: ebiten.StandardGamepadButtonCenterLeft,  // Select/Back button

		// Gamepad-specific buttons
		joyUp:    ebiten.StandardGamepadButtonLeftTop,
		joyDown:  ebiten.StandardGamepadButtonLeftBottom,
		joyLeft:  ebiten.StandardGamepadButtonLeftLeft,
		joyRight: ebiten.StandardGamepadButtonLeftRight,
		joyA:     ebiten.StandardGamepadButtonRightBottom, // A button (bottom face button)
		joyB:     ebiten.StandardGamepadButtonRightRight,  // B button (right face button)
		joyX:     ebiten.StandardGamepadButtonRightLeft,   // X button (left face button)
		joyY:     ebiten.StandardGamepadButtonRightTop,    // Y button (top face button)
	}

	// Set platform-specific overrides if needed
	if isSteamDeck() {
		// Steam Deck specific button mappings
		pico8ButtonToStandard = map[int]ebiten.StandardGamepadButton{
			// Face buttons (A,B,X,Y) - Steam Deck uses XBox layout
			X: ebiten.StandardGamepadButtonRightLeft,   // X button (Left)
			O: ebiten.StandardGamepadButtonRightBottom, // A button (Bottom)

			joyA: ebiten.StandardGamepadButtonRightBottom, // A button (bottom)
			joyB: ebiten.StandardGamepadButtonRightRight,  // B button (right)
			joyX: ebiten.StandardGamepadButtonRightLeft,   // X button (left)
			joyY: ebiten.StandardGamepadButtonRightTop,    // Y button (top)

			// D-pad directions
			UP:       ebiten.StandardGamepadButtonLeftTop,
			DOWN:     ebiten.StandardGamepadButtonLeftBottom,
			LEFT:     ebiten.StandardGamepadButtonLeftLeft,
			RIGHT:    ebiten.StandardGamepadButtonLeftRight,
			joyUp:    ebiten.StandardGamepadButtonLeftTop,
			joyDown:  ebiten.StandardGamepadButtonLeftBottom,
			joyLeft:  ebiten.StandardGamepadButtonLeftLeft,
			joyRight: ebiten.StandardGamepadButtonLeftRight,

			// Shoulder buttons
			l1: ebiten.StandardGamepadButtonFrontTopLeft,     // L1
			r1: ebiten.StandardGamepadButtonFrontTopRight,    // R1
			l2: ebiten.StandardGamepadButtonFrontBottomLeft,  // L2 (also analog)
			r2: ebiten.StandardGamepadButtonFrontBottomRight, // R2 (also analog)

			// Stick clicks
			l3: ebiten.StandardGamepadButtonLeftStick,  // Left stick click
			r3: ebiten.StandardGamepadButtonRightStick, // Right stick click

			// Menu buttons
			START:  ebiten.StandardGamepadButtonCenterRight, // Menu button (right)
			SELECT: ebiten.StandardGamepadButtonCenterLeft,  // View button (left)

			// Steam/Quick Access button is not mappable through standard gamepad API
		}

		// Map Steam Deck touchpad clicks
		// These are mapped to mouse buttons for compatibility
		pico8ButtonToStandard[MouseLeft] = ebiten.StandardGamepadButtonFrontBottomLeft   // Map to L2
		pico8ButtonToStandard[MouseRight] = ebiten.StandardGamepadButtonFrontBottomRight // Map to R2

		// Map back buttons (L4/L5, R4/R5 on Steam Deck)
		// These are mapped to L1/R1 for now since they're not standard
		pico8ButtonToStandard[l4] = ebiten.StandardGamepadButtonFrontTopLeft  // L4 -> L1
		pico8ButtonToStandard[r4] = ebiten.StandardGamepadButtonFrontTopRight // R4 -> R1
		pico8ButtonToStandard[l5] = ebiten.StandardGamepadButtonLeftStick     // L5 -> Left stick click
		pico8ButtonToStandard[r5] = ebiten.StandardGamepadButtonRightStick    // R5 -> Right stick click

		// Note: For full Steam Deck back button support, you might want to use
		// SDL's game controller API directly or a Steam Input wrapper
	}
}

// init initializes the button mappings when the package is imported
func init() {
	initButtonMappings()
}

// pico8ButtonToKeyboardP0 maps PICO-8 button indices to default keyboard keys for Player 0.
// Updated for better Steam Deck keyboard/on-screen keyboard support
var pico8ButtonToKeyboardP0 = map[int]ebiten.Key{
	// Arrow keys for direction
	LEFT:  ebiten.KeyLeft,
	RIGHT: ebiten.KeyRight,
	UP:    ebiten.KeyUp,
	DOWN:  ebiten.KeyDown,

	// Face buttons (mapped to common game keys)
	O: ebiten.KeyZ, // PICO-8 O button ('Z' key)
	X: ebiten.KeyX, // PICO-8 X button ('X' key)

	// Menu buttons
	START:  ebiten.KeyEnter, // Start/Confirm
	SELECT: ebiten.KeyTab,   // Select/Back

	// Additional Steam Deck specific mappings
	// These are useful for Steam Deck's on-screen keyboard
	// You can add more mappings as needed
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

// isMouseButton checks if the given buttonIndex corresponds to a mouse button or wheel.
func isMouseButton(buttonIndex int) bool {
	switch buttonIndex {
	case MouseLeft, MouseRight, MouseMiddle, MouseWheelUp, MouseWheelDown:
		return true
	default:
		return false
	}
}

// handleMouseInput checks if the specified PICO-8 mouse button/wheel is currently active.
func handleMouseInput(buttonIndex int) bool {
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
	return false
}

// handleKeyboardInput checks if the specified PICO-8 button is pressed on the keyboard for Player 0.
func handleKeyboardInput(buttonIndex int) bool {
	if key, ok := pico8ButtonToKeyboardP0[buttonIndex]; ok {
		return ebiten.IsKeyPressed(key)
	}
	return false
}

// isDirectionButton checks if the buttonIndex corresponds to a directional PICO-8 button.
func isDirectionButton(buttonIndex int) bool {
	switch buttonIndex {
	case LEFT, RIGHT, UP, DOWN, joyLeft, joyRight, joyUp, joyDown:
		return true
	default:
		return false
	}
}

// handleGamepadDirectionalInput checks for directional inputs (D-pad and analog stick) on the gamepad.
func handleGamepadDirectionalInput(buttonIndex int, gamepadID ebiten.GamepadID) bool {
	axisThreshold := 0.5
	switch buttonIndex {
	case LEFT, joyLeft:
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			if ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftLeft) ||
				ebiten.StandardGamepadAxisValue(gamepadID, ebiten.StandardGamepadAxisLeftStickHorizontal) < -axisThreshold {
				return true
			}
		}
		return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftLeft)) ||
			ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickHorizontal)) < -axisThreshold
	case RIGHT, joyRight:
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			if ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftRight) ||
				ebiten.StandardGamepadAxisValue(gamepadID, ebiten.StandardGamepadAxisLeftStickHorizontal) > axisThreshold {
				return true
			}
		}
		return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftRight)) ||
			ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickHorizontal)) > axisThreshold
	case UP, joyUp:
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			if ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftTop) ||
				ebiten.StandardGamepadAxisValue(gamepadID, ebiten.StandardGamepadAxisLeftStickVertical) < -axisThreshold {
				return true
			}
		}
		return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftTop)) ||
			ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickVertical)) < -axisThreshold
	case DOWN, joyDown:
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			if ebiten.IsStandardGamepadButtonPressed(gamepadID, ebiten.StandardGamepadButtonLeftBottom) ||
				ebiten.StandardGamepadAxisValue(gamepadID, ebiten.StandardGamepadAxisLeftStickVertical) > axisThreshold {
				return true
			}
		}
		return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftBottom)) ||
			ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickVertical)) > axisThreshold
	}
	return false
}

// handleGamepadStandardButtonInput checks for standard PICO-8 button presses on the gamepad.
func handleGamepadStandardButtonInput(buttonIndex int, gamepadID ebiten.GamepadID) bool {
	if standardButton, ok := pico8ButtonToStandard[buttonIndex]; ok {
		if ebiten.IsStandardGamepadLayoutAvailable(gamepadID) {
			return ebiten.IsStandardGamepadButtonPressed(gamepadID, standardButton)
		}
		switch buttonIndex {
		case LEFT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftLeft))
		case RIGHT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftRight))
		case UP:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftTop))
		case DOWN:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftBottom))
		case O:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton1)
		case X:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton0)
		case START:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton9)
		case SELECT:
			return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton8)
		default:
			if btn, found := pico8ButtonToStandard[buttonIndex]; found {
				return ebiten.IsGamepadButtonPressed(gamepadID, ebiten.GamepadButton(btn))
			}
		}
	}
	return false
}

// Btn checks if a specific PICO-8 button is currently held down via gamepad, keyboard (Player 0 only), mouse, or gamepad axes.
// Mimics the PICO-8 btn() function behavior (returns true while held).
//
// buttonIndex: The PICO-8 button index (0-15).
// playerIndex: Optional PICO-8 player index (0-7). Defaults to 0 (player 1) if omitted.
func Btn(buttonIndex int, playerIndex ...int) bool {
	pIdx := 0 // Default to player 0
	if len(playerIndex) > 0 {
		pIdx = playerIndex[0]
	}

	// Check mouse input first
	if isMouseButton(buttonIndex) {
		return handleMouseInput(buttonIndex)
	}

	// Check keyboard input (player 0 only)
	if pIdx == 0 && handleKeyboardInput(buttonIndex) {
		return true
	}

	// Check gamepad input
	ids := ebiten.AppendGamepadIDs(nil)
	if pIdx < 0 || pIdx >= len(ids) {
		return false // No gamepad connected for this player index
	}
	gamepadID := ids[pIdx]

	// Check directional inputs first
	if isDirectionButton(buttonIndex) {
		return handleGamepadDirectionalInput(buttonIndex, gamepadID)
	}

	// Then check standard button mappings
	return handleGamepadStandardButtonInput(buttonIndex, gamepadID)
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
		// Directional buttons (D-pad and left stick)
		case LEFT, joyLeft:
			// Check D-pad left or left stick left
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftLeft)) ||
				ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickHorizontal)) < -0.5
		case RIGHT, joyRight:
			// Check D-pad right or left stick right
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftRight)) ||
				ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickHorizontal)) > 0.5
		case UP, joyUp:
			// Check D-pad up or left stick up
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftTop)) ||
				ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickVertical)) < -0.5
		case DOWN, joyDown:
			// Check D-pad down or left stick down
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton(ebiten.StandardGamepadButtonLeftBottom)) ||
				ebiten.GamepadAxisValue(gamepadID, int(ebiten.StandardGamepadAxisLeftStickVertical)) > 0.5
		// Face buttons
		case O, joyB: // B button (right face button)
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton1)
		case X, joyA: // A button (bottom face button)
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton0)
		case joyX: // X button (left face button)
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton2)
		case joyY: // Y button (top face button)
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton3)
		// Menu buttons
		case START:
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton9)
		case SELECT:
			return inpututil.IsGamepadButtonJustPressed(gamepadID, ebiten.GamepadButton8)
		}
	}

	return false
}
