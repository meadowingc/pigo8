//go:build js && wasm

package pigo8

import "syscall/js"

// syncVirtualFromJS mirrors window.pigo8BtnBits (uint32) into internal virtual button state.
// Bits: 0 LEFT, 1 RIGHT, 2 UP, 3 DOWN, 4 O, 5 X, 6 ButtonStart (Pause)
func syncVirtualFromJS() {
	v := js.Global().Get("pigo8BtnBits")
	setVirtualButtonsBitfield(uint32(v.Int()))
}
