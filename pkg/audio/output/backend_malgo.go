//go:build cgo

// ABOUTME: Malgo backend adapter for the neutral backend interface
// ABOUTME: Exposes miniaudio enumeration/probing as backend.Backend

package output

import (
	"strings"

	"github.com/Sendspin/sendspin-go/pkg/audio/output/backend"
)

// MalgoBackend exposes the miniaudio backend through the neutral
// backend.Backend interface (device enumeration + capability probe).
// Playback itself stays on NewMalgo (output.Output).
type MalgoBackend struct{}

// List enumerates miniaudio playback devices in backend-neutral form.
// The opaque ID is miniaudio's device id; Name is what Open and Probe
// accept.
func (MalgoBackend) List() ([]backend.Device, error) {
	devices, err := ListPlaybackDevices()
	if err != nil {
		return nil, err
	}
	out := make([]backend.Device, 0, len(devices))
	for _, d := range devices {
		out = append(out, backend.Device{
			Name:      d.Name,
			ID:        strings.TrimRight(string(d.ID[:]), "\x00"),
			IsDefault: d.IsDefault,
		})
	}
	return out, nil
}

// Probe reports the rate/bit-depth ceiling for a named device.
// MaxChannels is 0 (unknown): miniaudio reports formats, not channel
// bounds, so callers must not treat 0 as a limit.
func (MalgoBackend) Probe(device string) (backend.Capabilities, error) {
	rate, bits, err := QueryDeviceCapabilities(device)
	if err != nil {
		return backend.Capabilities{}, err
	}
	return backend.Capabilities{MaxSampleRate: rate, MaxBitDepth: bits}, nil
}
