//go:build !js || !wasm

package pigo8

// syncVirtualFromJS is a no-op stub on non-wasm targets.
// The real implementation is provided in wasm_virtual_input.go for js/wasm builds.
func syncVirtualFromJS() {}
