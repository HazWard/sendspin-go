//go:build !cgo

// ABOUTME: Static-build stub for the malgo audio backend
// ABOUTME: miniaudio needs cgo; static builds inject their own Output

package output

// Malgo is a static-build placeholder. Only NewMalgo's signature is
// provided so callers compile; it returns nil. A static player MUST
// inject a real backend via PlayerConfig.Output — using the default
// (nil) output will fail loudly at stream start instead of playing
// silence.
type Malgo struct{}

// NewMalgo always returns nil in static builds: there is no miniaudio
// without cgo. Inject an Output implementation instead.
func NewMalgo(deviceName string) Output {
	return nil
}
