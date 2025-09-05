package pigo8

// Virtual (touch / on-screen) button support.
// We store up to 32 virtual buttons in a bitfield (LOW bits).
// 0: LEFT
// 1: RIGHT
// 2: UP
// 3: DOWN
// 4: O
// 5: X
// 6: ButtonStart (Pause)
// Additional bits are reserved for future expansion.

const maxVirtualButtons = 32

var virtualButtonBits uint32

// setVirtualButtonsBitfield replaces the entire virtual button bitfield.
func setVirtualButtonsBitfield(bits uint32) {
	virtualButtonBits = bits
}

// getVirtualButton returns true if the virtual button at index is active.
func getVirtualButton(index int) bool {
	if index < 0 || index >= maxVirtualButtons {
		return false
	}
	return (virtualButtonBits>>index)&1 == 1
}
