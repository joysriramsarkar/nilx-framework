// Package animation provides timing curves, spring physics, and interpolators for NilX UI.
package animation

import (
	"math"
)

// EasingFunc maps a normalized time progress t in [0.0, 1.0] to an output progress.
type EasingFunc func(t float64) float64

// Linear interpolation: f(t) = t
func Linear(t float64) float64 {
	return t
}

// EaseInQuad: quadratic acceleration
func EaseInQuad(t float64) float64 {
	return t * t
}

// EaseOutQuad: quadratic deceleration
func EaseOutQuad(t float64) float64 {
	return t * (2 - t)
}

// EaseInOutCubic: smooth acceleration and deceleration
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	p := 2*t - 2
	return 0.5*p*p*p + 1
}

// Spring computes a damped harmonic oscillation at progress t.
func Spring(t float64, damping, frequency float64) float64 {
	return 1.0 - math.Exp(-damping*t)*math.Cos(frequency*t)
}

// Interpolate maps progress t between start and end values.
func Interpolate(start, end, t float64, ease EasingFunc) float64 {
	if ease == nil {
		ease = Linear
	}
	progress := ease(t)
	return start + (end-start)*progress
}
