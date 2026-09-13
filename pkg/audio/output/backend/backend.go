// ABOUTME: Backend-neutral audio device vocabulary and capabilities
// ABOUTME: Device/Capabilities/Backend plus the shared device matcher

// Package backend describes what an audio playback backend offers and
// how the rest of the application consumes it, independent of any
// particular implementation (miniaudio/malgo, raw ALSA, ...).
//
// Today malgo is the only in-tree backend and the mapping is:
//
//   - Backend.List: device enumeration, used by the player CLI
//     (--list-audio-devices) and the TUI device picker.
//   - Backend.Probe: capability probing for the player auto-probe
//     (what the server may be offered in client/hello).
//   - Playback itself stays behind each backend's own constructor
//     (output.NewMalgo, alsa.NewSink) returning output.Output.
//
// MatchDevice holds the shared selection policy (blank means default,
// exact name wins, unambiguous short prefix accepted, everything else
// is a loud error) so every backend and UI resolves names the same way.
package backend

import (
	"fmt"
	"sort"
	"strings"
)

// Device is a backend-neutral playback endpoint.
type Device struct {
	// Name is human-readable and what users pass to --audio-device.
	Name string
	// ID is backend-opaque (miniaudio device id, ALSA hw:C,D, ...).
	ID string
	// IsDefault marks the backend's preferred default, if any.
	IsDefault bool
}

// Capabilities bounds what a device can do. Zero values mean unknown.
type Capabilities struct {
	MaxSampleRate int
	MaxBitDepth   int
	MaxChannels   int
}

// Backend is the capability surface of a playback backend: enumerate
// devices and probe what they support.
type Backend interface {
	List() ([]Device, error)
	Probe(device string) (Capabilities, error)
}

// MatchDevice picks a Device from a list based on a requested name.
//
// Empty requested name -> the device with IsDefault set, else the first in
// the list, else nil if the list is empty.
//
// Non-empty requested name -> exact Name match first, then short-name match
// (the text before the first ", "). Miniaudio's Linux/ALSA backend builds
// device names from snd_device_name_hint's DESC field, which follows a
// "<card-short>, <stream-description>" convention, so users naturally try
// just the short part. If the short-name match is ambiguous, we error out
// instead of picking one silently.
//
// Fail-loud on no-match: the error lists every available device name, each
// quoted with %q so embedded commas are distinguishable from the list
// separator. Silent fallback to default is the behavior this feature
// exists to correct.
func MatchDevice(devices []Device, requested string) (*Device, error) {
	if requested == "" {
		if len(devices) == 0 {
			return nil, nil
		}
		for i := range devices {
			if devices[i].IsDefault {
				return &devices[i], nil
			}
		}
		return &devices[0], nil
	}
	for i := range devices {
		if devices[i].Name == requested {
			return &devices[i], nil
		}
	}
	var shortMatches []int
	for i, d := range devices {
		if idx := strings.Index(d.Name, ", "); idx > 0 && d.Name[:idx] == requested {
			shortMatches = append(shortMatches, i)
		}
	}
	if len(shortMatches) == 1 {
		return &devices[shortMatches[0]], nil
	}
	if len(devices) == 0 {
		return nil, fmt.Errorf("audio device %q not found (no playback devices available)", requested)
	}
	quoted := make([]string, len(devices))
	for i, d := range devices {
		quoted[i] = fmt.Sprintf("%q", d.Name)
	}
	sort.Strings(quoted)
	if len(shortMatches) > 1 {
		return nil, fmt.Errorf("audio device %q is ambiguous (matches %d devices by short name); use the full quoted name. Available: %s", requested, len(shortMatches), strings.Join(quoted, ", "))
	}
	return nil, fmt.Errorf("audio device %q not found; available: %s", requested, strings.Join(quoted, ", "))
}
