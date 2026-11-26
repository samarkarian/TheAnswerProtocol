package direction

import "testing"

// known lists every valid direction.
var known = All()

func TestParse(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   Direction
		wantOK bool
	}{
		// Directions valides
		{"north", "north", North, true},
		{"south", "south", South, true},
		{"east", "east", East, true},
		{"west", "west", West, true},

		// Insensible à la casse, résultat toujours en minuscules
		{"capitalized", "North", North, true},
		{"upper case", "EAST", East, true},
		{"mixed case", "wEsT", West, true},

		// Invalides
		{"empty", "", "", false},
		{"typo", "nort", "", false},
		{"abbreviation", "n", "", false},
		{"unsupported direction", "up", "", false},
		{"leading space", " north", "", false},
		{"trailing space", "north ", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Parse(tt.in)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("Parse(%q) = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestOpposite(t *testing.T) {
	tests := []struct {
		in     Direction
		want   Direction
		wantOK bool
	}{
		{North, South, true},
		{South, North, true},
		{East, West, true},
		{West, East, true},
		{"up", "", false},
		{"North", "", false}, // pas normalisée : Opposite ne fait pas le travail de Parse
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.in), func(t *testing.T) {
			got, ok := tt.in.Opposite()
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("%q.Opposite() = (%q, %v), want (%q, %v)", tt.in, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// Applying Opposite twice must give back the starting direction.
func TestOppositeTwice(t *testing.T) {
	for _, d := range known {
		opp, ok := d.Opposite()
		if !ok {
			t.Fatalf("%q.Opposite() is not ok", d)
		}
		if opp == d {
			t.Errorf("%q.Opposite() = %q, a direction cannot be its own opposite", d, opp)
		}
		back, ok := opp.Opposite()
		if !ok || back != d {
			t.Errorf("%q.Opposite().Opposite() = (%q, %v), want (%q, true)", d, back, ok, d)
		}
	}
}

// Parse must accept the exact value of every constant, unchanged.
func TestParseConstants(t *testing.T) {
	for _, d := range known {
		got, ok := Parse(string(d))
		if !ok || got != d {
			t.Errorf("Parse(%q) = (%q, %v), want (%q, true)", d, got, ok, d)
		}
	}
}
