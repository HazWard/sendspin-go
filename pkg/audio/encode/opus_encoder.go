// ABOUTME: Opus encoder (pure-Go pion/opus, RFC 6716)
// ABOUTME: 48 kHz 20 ms frames, tuned like AppAudio at 128 kbit/s/ch

package encode

import (
	"encoding/binary"
	"fmt"

	"github.com/pion/opus"
)

// frameSamples is the one frame length the pion encoder takes:
// exactly one 20 ms frame at 48 kHz.
const pionFrameSamples = 960

// maxPacketBytes bounds a single Opus packet on the wire.
const maxPacketBytes = 1276

// OpusEncoder encodes 16-bit PCM to Opus packets using a pure-Go
// RFC 6716 implementation. Constraints mirror the encoder slice:
// 48 kHz input (the server resamples to 48 kHz for Opus already)
// and single 20 ms frames.
type OpusEncoder struct {
	encoder    *opus.Encoder
	channels   int
	outBuf     []byte
	bitrate    int
	sampleRate int
}

// NewOpusEncoder creates an OpusEncoder tuned like the cgo backend:
// full-bandwidth music at 128 kbit/s per channel.
func NewOpusEncoder(sampleRate, channels, frameSize int) (*OpusEncoder, error) {
	if sampleRate != 48000 {
		return nil, fmt.Errorf("pion opus encoder supports 48 kHz only, got %d", sampleRate)
	}
	if channels < 1 || channels > 2 {
		return nil, fmt.Errorf("pion opus encoder supports 1-2 channels, got %d", channels)
	}
	if frameSize != pionFrameSamples {
		return nil, fmt.Errorf("pion opus encoder takes 20 ms frames (%d samples), got %d", pionFrameSamples, frameSize)
	}

	enc, err := opus.NewEncoder(
		opus.WithSampleRate(sampleRate),
		opus.WithChannels(channels),
		opus.WithBitrate(128000*channels),
		opus.WithApplication(opus.ApplicationAudio),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder: %w", err)
	}

	return &OpusEncoder{
		encoder:    enc,
		channels:   channels,
		outBuf:     make([]byte, maxPacketBytes),
		bitrate:    128000 * channels,
		sampleRate: sampleRate,
	}, nil
}

// Encode encodes one 20 ms frame of interleaved int16 PCM.
func (e *OpusEncoder) Encode(pcm []int16) ([]byte, error) {
	if len(pcm) != pionFrameSamples*e.channels {
		return nil, fmt.Errorf("opus encode needs %d samples, got %d", pionFrameSamples*e.channels, len(pcm))
	}

	in := make([]byte, len(pcm)*2)
	for i, s := range pcm {
		binary.LittleEndian.PutUint16(in[i*2:], uint16(s))
	}

	n, err := e.encoder.Encode(in, e.outBuf)
	if err != nil {
		return nil, fmt.Errorf("opus encode failed: %w", err)
	}

	packet := make([]byte, n)
	copy(packet, e.outBuf[:n])
	return packet, nil
}

// Close releases encoder resources (no-op for the pure-Go encoder).
func (e *OpusEncoder) Close() error {
	return nil
}
