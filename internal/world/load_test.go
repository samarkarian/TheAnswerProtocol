package world

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// writeWorld writes content to a temporary world file and returns its path.
func writeWorld(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "world.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadValid(t *testing.T) {
	w, err := Load(filepath.Join("testdata", "valid.json"))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	if w.Spawn != "loc.hall" {
		t.Errorf("Spawn = %q, want %q", w.Spawn, "loc.hall")
	}
	if len(w.Rooms) != 2 {
		t.Fatalf("len(Rooms) = %d, want 2", len(w.Rooms))
	}

	hall, ok := w.byID["loc.hall"]
	if !ok {
		t.Fatal(`byID has no "loc.hall"`)
	}
	if hall.Name != "Hall" || hall.Description != "The entrance of the campus." {
		t.Errorf("hall = %+v, want name and description from the file", hall)
	}
	if got := hall.Exits["north"]; got != "loc.bocal" {
		t.Errorf(`hall.Exits["north"] = %q, want "loc.bocal"`, got)
	}
	if len(w.byID) != len(w.Rooms) {
		t.Errorf("len(byID) = %d, want %d", len(w.byID), len(w.Rooms))
	}

	// The index must point into w.Rooms, not to copies.
	if hall != &w.Rooms[0] {
		t.Error(`byID["loc.hall"] does not point to w.Rooms[0]`)
	}
}

func TestLoadMissingFile(t *testing.T) {
	w, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Load() error = %v, want it to match fs.ErrNotExist", err)
	}
	if w != nil {
		t.Errorf("Load() returned a World on error: %+v", w)
	}
}

func TestLoadErrors(t *testing.T) {
	const room = `{"id": "loc.hall", "name": "Hall", "description": "x", "exits": {}}`

	tests := []struct {
		name    string
		content string
		wantErr error // nil: any error is accepted, we only check there is one
	}{
		{"empty file", ``, nil},
		{"invalid json", `{"spawn": "loc.hall",`, nil},
		{"unknown top-level field", `{"spawn": "loc.hall", "rooms": [], "romos": []}`, nil},
		{"unknown room field", `{"spawn": "loc.hall", "rooms": [{"id": "loc.hall", "exist": {}}]}`, nil},
		{"exits is not an object", `{"spawn": "loc.hall", "rooms": [{"id": "loc.hall", "exits": "north"}]}`, nil},
		{"rooms is not an array", `{"spawn": "loc.hall", "rooms": {}}`, nil},
		{"duplicate room id", `{"spawn": "loc.hall", "rooms": [` + room + `, ` + room + `]}`, ErrDuplicateRoom},
		{"trailing data", `{"spawn": "loc.hall", "rooms": []} {"spawn": "loc.other"}`, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := Load(writeWorld(t, tt.content))
			if err == nil {
				t.Fatal("Load() error = nil, want an error")
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("Load() error = %v, want it to match %v", err, tt.wantErr)
			}
			if w != nil {
				t.Errorf("Load() returned a World on error: %+v", w)
			}
		})
	}
}

// Load only reads the file: checking that the world makes sense is Validate's job.
func TestLoadDoesNotValidate(t *testing.T) {
	_, err := Load(writeWorld(t, `{"spawn": "loc.nowhere", "rooms": []}`))
	if err != nil {
		t.Errorf("Load() error = %v, want nil (validation belongs to Validate)", err)
	}
}
