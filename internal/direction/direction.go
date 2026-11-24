// Package direction defines the movement directions shared by the server and the clients.
package direction

import (
	"slices"
	"strings"
)

// Direction is a movement direction, as used in exits and in the MOVE command.
type Direction string

// Directions supported by the protocol.
const (
	North Direction = "north"
	East  Direction = "east"
	South Direction = "south"
	West  Direction = "west"
)

// Opposite returns the opposite Direction and whether d is a known direction.
func (d Direction) Opposite() (Direction, bool) {
	switch d {
	case North:
		return South, true
	case East:
		return West, true
	case South:
		return North, true
	case West:
		return East, true
	default:
		return "", false
	}
}

// Delta returns the grid offset of d, with y growing downward (North is (0, -1)).
// It returns (0, 0) for an unknown direction.
func (d Direction) Delta() (dx, dy int) {
	switch d {
	case North:
		return 0, -1
	case East:
		return 1, 0
	case South:
		return 0, 1
	case West:
		return -1, 0
	default:
		return 0, 0
	}
}

// All returns all directions in clockwise order, starting from North.
func All() []Direction {
	return []Direction{North, East, South, West}
}

// Parse converts s (case-insensitive) into a Direction and whether s is a known direction.
func Parse(s string) (Direction, bool) {
	d := Direction(strings.ToLower(s))

	if slices.Contains(All(), d) {
		return d, true
	}
	return "", false
}
