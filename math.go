package pigo8

import (
	"math"
	"math/rand"
)

// Flr rounds the given number down and returns the nearest integer (whole number).
// It mimics the behavior of PICO-8's `flr()` function.
//
// Due to the use of generics [T Number], the input `a` can be any standard integer
// or float type (e.g., int, float64, int32, float32).
// The function always returns an int.
//
// Args:
//   - a: The number (integer or float) to round down.
//
// Returns:
//   - int: The nearest whole integer less than or equal to `a`.
//
// Example:
//
//	val1 := Flr(1.99)   // val1 will be 1
//	val2 := Flr(-5.3)  // val2 will be -6
//	val3 := Flr(10)    // val3 will be 10
//	val4 := Flr(-2)   // val4 will be -2
func Flr[T Number](a T) int {
	// Convert the generic number to float64 for math.Floor
	floatVal := float64(a)
	// Apply floor operation
	floorVal := math.Floor(floatVal)
	// Convert the result to int
	return int(floorVal)
}

// Rnd returns a random integer between 0 (inclusive) and the integer part of the
// given upper bound `a` (exclusive).
// It mimics the behavior of PICO-8's `flr(rnd(a))`.
//
// Due to the use of generics [T Number], the input `a` can be any standard integer
// or float type (e.g., int, float64, int32, float32).
// The function always returns an int.
//
// If `a` is zero or negative, Rnd returns 0.
// If `a` is positive, the result is in the range [0, floor(a)).
//
// Note: This uses Go's standard `math/rand` package. Unlike PICO-8's default `rnd()`,
// the sequence is not deterministic across program runs unless the global random
// source is explicitly seeded using `rand.Seed()`.
//
// Args:
//   - a: The upper exclusive bound (any Number type) for the random number.
//
// Returns:
//   - int: A random integer in the range [0, floor(a)).
//
// Example:
//
//	val1 := Rnd(5)     // val1 will be an int: 0, 1, 2, 3, or 4
//	val2 := Rnd(5.9)   // val2 will be an int: 0, 1, 2, 3, or 4 (floor(5.9) = 5)
//	val3 := Rnd(1.1)   // val3 will be an int: 0 or 1 (floor(1.1) = 1)
//	val4 := Rnd(1)     // val4 will be 0 (floor(1) = 1)
//	val5 := Rnd(0.5)   // val5 will be 0 (floor(0.5) = 0)
//	val6 := Rnd(0)     // val6 will be 0
//	val7 := Rnd(-10)   // val7 will be 0
func Rnd[T Number](a T) int {
	limit := float64(a)

	if limit <= 0 {
		return 0
	}

	// rand.Float64() returns a float64 in [0.0, 1.0)
	// Multiplying by limit gives a float64 in [0.0, limit)
	// Applying Floor and converting to int gives an integer in [0, floor(limit))
	return int(math.Floor(rand.Float64() * limit))
}

// Sqrt returns the square root of the given number.
// It mimics the behavior of PICO-8's `sqrt()` function.
//
// Due to the use of generics [T Number], the input `a` can be any standard integer
// or float type (e.g., int, float64, int32, float32).
// The function always returns a float64.
//
// If `a` is negative, Sqrt returns 0 (unlike Go's math.Sqrt which returns NaN).
//
// Args:
//   - a: The number to calculate the square root of.
//
// Returns:
//   - float64: The square root of `a`, or 0 if `a` is negative.
//
// Example:
//
//	val1 := Sqrt(16)    // val1 will be 4.0
//	val2 := Sqrt(2)     // val2 will be approximately 1.4142
//	val3 := Sqrt(0)     // val3 will be 0.0
//	val4 := Sqrt(-4)    // val4 will be 0.0 (PICO-8 compatibility)
func Sqrt[T Number](a T) float64 {
	value := float64(a)
	
	// Handle negative values by returning 0 (PICO-8 compatibility)
	if value < 0 {
		return 0.0
	}
	
	return math.Sqrt(value)
}
