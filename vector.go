package pigo8

import (
	"fmt"
	"math"
)

// Vector2D represents a 2D vector with X and Y components
type Vector2D struct {
	X, Y float64
}

// NewVector2D creates a new Vector2D with the specified components
func NewVector2D(x, y float64) Vector2D {
	return Vector2D{X: x, Y: y}
}

// Add returns a new vector that is the sum of this vector and another vector
func (v Vector2D) Add(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X + other.X,
		Y: v.Y + other.Y,
	}
}

// Sub returns a new vector that is the difference of this vector and another vector
func (v Vector2D) Sub(other Vector2D) Vector2D {
	return Vector2D{
		X: v.X - other.X,
		Y: v.Y - other.Y,
	}
}

// Scale returns a new vector that is this vector scaled by a factor
func (v Vector2D) Scale(factor float64) Vector2D {
	return Vector2D{
		X: v.X * factor,
		Y: v.Y * factor,
	}
}

// Magnitude returns the length of this vector
func (v Vector2D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

// Normalize returns a new vector in the same direction but with a length of 1
// If the vector has zero length, it returns a zero vector
func (v Vector2D) Normalize() Vector2D {
	mag := v.Magnitude()
	if mag == 0 {
		return Vector2D{0, 0}
	}
	return Vector2D{
		X: v.X / mag,
		Y: v.Y / mag,
	}
}

// Distance returns the distance between this vector and another vector
func (v Vector2D) Distance(other Vector2D) float64 {
	return v.Sub(other).Magnitude()
}

// Dot returns the dot product of this vector and another vector
func (v Vector2D) Dot(other Vector2D) float64 {
	return v.X*other.X + v.Y*other.Y
}

// AngleBetween returns the angle in radians between this vector and another vector
func (v Vector2D) AngleBetween(other Vector2D) float64 {
	dot := v.Dot(other)
	mag1 := v.Magnitude()
	mag2 := other.Magnitude()

	// Avoid division by zero
	if mag1 == 0 || mag2 == 0 {
		return 0
	}

	// Clamp to avoid floating point errors
	cosTheta := dot / (mag1 * mag2)
	if cosTheta > 1 {
		cosTheta = 1
	} else if cosTheta < -1 {
		cosTheta = -1
	}

	return math.Acos(cosTheta)
}

// ToInt returns a new vector with the components rounded to integers
func (v Vector2D) ToInt() (int, int) {
	return int(math.Round(v.X)), int(math.Round(v.Y))
}

// String returns a string representation of the vector
func (v Vector2D) String() string {
	return fmt.Sprintf("Vector2D(%.2f, %.2f)", v.X, v.Y)
}

// ZeroVector returns a vector with both components set to zero
func ZeroVector() Vector2D {
	return Vector2D{0, 0}
}

// DirectionVector returns a unit vector in the specified cardinal direction
func DirectionVector(direction int) Vector2D {
	switch direction {
	case ButtonUp:
		return Vector2D{0, -1}
	case ButtonDown:
		return Vector2D{0, 1}
	case ButtonLeft:
		return Vector2D{-1, 0}
	case ButtonRight:
		return Vector2D{1, 0}
	default:
		return Vector2D{0, 0}
	}
}
