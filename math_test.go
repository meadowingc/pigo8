package pigo8

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlr(t *testing.T) {
	t.Run("Positive Floats", func(t *testing.T) {
		assert.Equal(t, 1, Flr(1.99), "Flr(1.99) should be 1")
		assert.Equal(t, 1, Flr(1.01), "Flr(1.01) should be 1")
		assert.Equal(t, 1, Flr(1.0), "Flr(1.0) should be 1")
		assert.Equal(t, 0, Flr(0.99), "Flr(0.99) should be 0")
		assert.Equal(t, 123, Flr(123.456), "Flr(123.456) should be 123")
	})

	t.Run("Negative Floats", func(t *testing.T) {
		assert.Equal(t, -6, Flr(-5.3), "Flr(-5.3) should be -6")
		assert.Equal(t, -6, Flr(-5.99), "Flr(-5.99) should be -6")
		assert.Equal(t, -5, Flr(-5.0), "Flr(-5.0) should be -5")
		assert.Equal(t, -1, Flr(-0.01), "Flr(-0.01) should be -1")
		assert.Equal(t, -1, Flr(-0.99), "Flr(-0.99) should be -1")
	})

	t.Run("Zero", func(t *testing.T) {
		assert.Equal(t, 0, Flr(0.0), "Flr(0.0) should be 0")
		assert.Equal(t, 0, Flr(0), "Flr(0) should be 0")
	})

	t.Run("Positive Integers", func(t *testing.T) {
		assert.Equal(t, 10, Flr(10), "Flr(10) should be 10")
		assert.Equal(t, 1, Flr(1), "Flr(1) should be 1")
		assert.Equal(t, 0, Flr(0), "Flr(0) should be 0")
	})

	t.Run("Negative Integers", func(t *testing.T) {
		assert.Equal(t, -2, Flr(-2), "Flr(-2) should be -2")
		assert.Equal(t, -1, Flr(-1), "Flr(-1) should be -1")
	})

	t.Run("Different Numeric Types", func(t *testing.T) {
		assert.Equal(t, 3, Flr[int32](3), "int32")
		assert.Equal(t, -4, Flr[int64](-4), "int64")
		assert.Equal(t, 5, Flr[uint](5), "uint") // uint also satisfies constraints.Integer
		assert.Equal(t, 6, Flr[float32](6.7), "float32 (positive)")
		assert.Equal(t, -7, Flr[float32](-6.1), "float32 (negative)")
	})

	t.Run("Edge Cases (Max/Min Floats)", func(t *testing.T) {
		// Note: Conversion to int might overflow for extreme float64 values,
		// but Flr itself should handle the floor operation correctly.
		// We test within reasonable int range.
		assert.Equal(t, int(math.Floor(float64(math.MaxInt32)+0.5)), Flr(float64(math.MaxInt32)+0.5))
		assert.Equal(t, int(math.Floor(float64(math.MinInt32)-0.5)), Flr(float64(math.MinInt32)-0.5))
	})
}

func TestRnd(t *testing.T) {
	t.Run("Positive Float Limit", func(t *testing.T) {
		limit := 5.7
		expectedMax := 5           // flr(5.7) = 5
		for i := 0; i < 500; i++ { // Run more times to increase chance of hitting bounds
			val := Rnd(limit)
			assert.IsType(t, int(0), val, "Rnd should return an int")
			assert.GreaterOrEqual(t, val, 0, "Rnd(%v) should be >= 0", limit)
			assert.LessOrEqual(t, val, expectedMax, "Rnd(%v) should be <= %v", limit, expectedMax)
		}
	})

	t.Run("Positive Integer Limit", func(t *testing.T) {
		limit := 10
		expectedMax := 10 // flr(10) = 10
		for i := 0; i < 500; i++ {
			val := Rnd(limit)
			assert.IsType(t, int(0), val, "Rnd should return an int")
			assert.GreaterOrEqual(t, val, 0, "Rnd(%v) should be >= 0", limit)
			assert.Less(t, val, expectedMax, "Rnd(%v) should be < %v", limit, expectedMax)
		}
	})

	t.Run("Zero Limit", func(t *testing.T) {
		assert.Equal(t, 0, Rnd(0.0), "Rnd(0.0) should be 0")
		assert.Equal(t, 0, Rnd(0), "Rnd(0) should be 0")
	})

	t.Run("Negative Limits", func(t *testing.T) {
		assert.Equal(t, 0, Rnd(-10.5), "Rnd(-10.5) should be 0")
		assert.Equal(t, 0, Rnd(-1), "Rnd(-1) should be 0")
	})

	t.Run("Different Numeric Types for Limit", func(t *testing.T) {
		limitInt32 := int32(20)
		expectedMaxInt32 := 20
		val1 := Rnd(limitInt32)
		assert.IsType(t, int(0), val1, "Rnd(int32) should return int")
		assert.GreaterOrEqual(t, val1, 0)
		assert.Less(t, val1, expectedMaxInt32)

		limitFloat32 := float32(3.5)
		expectedMaxFloat32 := 3 // flr(3.5) = 3
		val2 := Rnd(limitFloat32)
		assert.IsType(t, int(0), val2, "Rnd(float32) should return int")
		assert.GreaterOrEqual(t, val2, 0)
		assert.LessOrEqual(t, val2, expectedMaxFloat32)

		limitUint := uint(5)
		expectedMaxUint := 5
		val3 := Rnd(limitUint)
		assert.IsType(t, int(0), val3, "Rnd(uint) should return int")
		assert.GreaterOrEqual(t, val3, 0)
		assert.Less(t, val3, expectedMaxUint)
	})
}
