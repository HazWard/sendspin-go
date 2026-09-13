// ABOUTME: Tests for the shared backend.MatchDevice selection logic
// ABOUTME: Ported from the malgo device matcher tests; runs without cgo

package backend

import (
	"strings"
	"testing"
)

// newDevice builds a Device with a unique sentinel ID so tests can
// assert the correct entry was returned.
func newDevice(name string, isDefault bool, id string) Device {
	return Device{Name: name, IsDefault: isDefault, ID: id}
}

func TestMatchDevice_EmptyRequest(t *testing.T) {
	tests := []struct {
		name     string
		devices  []Device
		wantNil  bool
		wantName string
		wantID   string
	}{
		{
			name:    "empty catalog returns nil",
			devices: nil,
			wantNil: true,
		},
		{
			name: "prefers the device flagged IsDefault",
			devices: []Device{
				newDevice("First", false, "id-01"),
				newDevice("DefaultSink", true, "id-02"),
				newDevice("Third", false, "id-03"),
			},
			wantName: "DefaultSink",
			wantID:   "id-02",
		},
		{
			name: "falls back to first device when none flagged default",
			devices: []Device{
				newDevice("Alpha", false, "id-0a"),
				newDevice("Beta", false, "id-0b"),
			},
			wantName: "Alpha",
			wantID:   "id-0a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MatchDevice(tt.devices, "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected a device, got nil")
			}
			if got.Name != tt.wantName {
				t.Errorf("name = %q, want %q", got.Name, tt.wantName)
			}
			if got.ID != tt.wantID {
				t.Errorf("id = %q, want %q (wrong slice element returned)", got.ID, tt.wantID)
			}
		})
	}
}

func TestMatchDevice_ExactNameMatch(t *testing.T) {
	devices := []Device{
		newDevice("HDA Intel PCH: ALC257 Analog", true, "id-10"),
		newDevice("HDMI 0", false, "id-11"),
		newDevice("USB Audio Device", false, "id-12"),
	}

	got, err := MatchDevice(devices, "USB Audio Device")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.Name != "USB Audio Device" {
		t.Errorf("got %+v, want USB Audio Device", got)
	}
	if got.ID != "id-12" {
		t.Errorf("wrong device matched: id = %q, want id-12", got.ID)
	}
}

func TestMatchDevice_NoMatchListsAvailable(t *testing.T) {
	devices := []Device{
		newDevice("Charlie", false, "id-01"),
		newDevice("Alpha", true, "id-02"),
		newDevice("Bravo", false, "id-03"),
	}

	got, err := MatchDevice(devices, "DoesNotExist")
	if got != nil {
		t.Errorf("expected nil device, got %+v", got)
	}
	if err == nil {
		t.Fatal("expected an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, `"DoesNotExist"`) {
		t.Errorf("error should name the missing device: %q", msg)
	}
	// Available names must be listed, sorted, so users can copy/paste the right one.
	for _, want := range []string{"Alpha", "Bravo", "Charlie"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error should list %q; got %q", want, msg)
		}
	}
	alphaIdx := strings.Index(msg, "Alpha")
	bravoIdx := strings.Index(msg, "Bravo")
	charlieIdx := strings.Index(msg, "Charlie")
	if !(alphaIdx < bravoIdx && bravoIdx < charlieIdx) {
		t.Errorf("available names should be sorted alphabetically; got %q", msg)
	}
}

func TestMatchDevice_NoMatchEmptyCatalogGivesDistinctError(t *testing.T) {
	got, err := MatchDevice(nil, "Anything")
	if got != nil {
		t.Errorf("expected nil device, got %+v", got)
	}
	if err == nil {
		t.Fatal("expected an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no playback devices available") {
		t.Errorf("error should distinguish empty-catalog case: %q", msg)
	}
}

// TestMatchDevice_ShortNameMatch covers miniaudio's Linux/ALSA naming where
// device.name is "<card-short>, <stream-description>" — users typing just
// the short prefix should match unambiguously when only one device has that
// prefix. Reproduces the HiFiBerry case from the field bug.
func TestMatchDevice_ShortNameMatch(t *testing.T) {
	devices := []Device{
		newDevice("Default Audio Device", true, "id-01"),
		newDevice("vc4-hdmi-0, MAI PCM i2s-hifi-0", false, "id-02"),
		newDevice("vc4-hdmi-1, MAI PCM i2s-hifi-0", false, "id-03"),
		newDevice("PDP Audio Device, USB Audio", false, "id-04"),
	}

	got, err := MatchDevice(devices, "vc4-hdmi-0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != "id-02" {
		t.Errorf("short-name %q should resolve to id-02; got %+v", "vc4-hdmi-0", got)
	}

	got, err = MatchDevice(devices, "PDP Audio Device")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != "id-04" {
		t.Errorf("short-name %q should resolve to id-04; got %+v", "PDP Audio Device", got)
	}
}

// TestMatchDevice_ExactNameWinsOverShortName guards the precedence: if a
// device's full name happens to equal someone else's short prefix, the
// exact match takes priority over the short-name search.
func TestMatchDevice_ExactNameWinsOverShortName(t *testing.T) {
	devices := []Device{
		newDevice("Foo, long description", false, "id-01"),
		newDevice("Foo", false, "id-02"),
	}

	got, err := MatchDevice(devices, "Foo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil || got.ID != "id-02" {
		t.Errorf("exact match should win: expected id-02; got %+v", got)
	}
}

// TestMatchDevice_ShortNameAmbiguousReturnsError covers the two-HiFiBerry
// case: the same short prefix matches multiple devices. We must not silently
// pick one.
func TestMatchDevice_ShortNameAmbiguousReturnsError(t *testing.T) {
	devices := []Device{
		newDevice("HiFiBerry, card 0", false, "id-01"),
		newDevice("HiFiBerry, card 1", false, "id-02"),
	}

	got, err := MatchDevice(devices, "HiFiBerry")
	if got != nil {
		t.Errorf("expected nil on ambiguous short-name match, got %+v", got)
	}
	if err == nil {
		t.Fatal("expected an error on ambiguous short-name match")
	}
	msg := err.Error()
	if !strings.Contains(msg, "ambiguous") {
		t.Errorf("error should mention ambiguity: %q", msg)
	}
	if !strings.Contains(msg, `"HiFiBerry, card 0"`) || !strings.Contains(msg, `"HiFiBerry, card 1"`) {
		t.Errorf("ambiguity error should list both candidates quoted with %%q: %q", msg)
	}
}

// TestMatchDevice_NoMatchQuotesNames ensures names with embedded commas are
// distinguishable from the list separator in the error output.
func TestMatchDevice_NoMatchQuotesNames(t *testing.T) {
	devices := []Device{
		newDevice("vc4-hdmi-0, MAI PCM i2s-hifi-0", false, "id-01"),
	}

	_, err := MatchDevice(devices, "nonexistent")
	if err == nil {
		t.Fatal("expected an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, `"vc4-hdmi-0, MAI PCM i2s-hifi-0"`) {
		t.Errorf("name should appear quoted in error so embedded comma is unambiguous: %q", msg)
	}
}
