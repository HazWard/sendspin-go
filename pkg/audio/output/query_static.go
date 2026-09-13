//go:build !cgo

// ABOUTME: Static-build fallback for device capability probing
// ABOUTME: Used when the malgo backend is excluded (CGO_ENABLED=0)

package output

// QueryDeviceCapabilities reports a conservative universal baseline in
// static builds: 48 kHz / 16-bit plays on effectively every ALSA
// device. Callers that need more (hi-res hardware) must set explicit
// caps instead of relying on auto-probe — in the headless player this
// is what --max-sample-rate / --max-bit-depth are for.
func QueryDeviceCapabilities(deviceName string) (maxSampleRate, maxBitDepth int, err error) {
	return 48000, 16, nil
}
