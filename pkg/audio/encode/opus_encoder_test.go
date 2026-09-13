// ABOUTME: Tests for the Opus encoder
// ABOUTME: Encode/decode roundtrip plus constructor validation

package encode

import (
	"math"
	"testing"

	"github.com/Sendspin/sendspin-go/pkg/audio"
	"github.com/Sendspin/sendspin-go/pkg/audio/decode"
)

func TestNewPionOpusEncoder(t *testing.T) {
	enc, err := NewOpusEncoder(48000, 2, 960)
	if err != nil {
		t.Fatalf("NewOpusEncoder failed: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestNewPionOpusEncoder_RejectsNon48k(t *testing.T) {
	if _, err := NewOpusEncoder(44100, 2, 960); err == nil {
		t.Fatal("expected error for non-48k rate")
	}
}

func TestNewPionOpusEncoder_RejectsBadFrameSize(t *testing.T) {
	if _, err := NewOpusEncoder(48000, 2, 480); err == nil {
		t.Fatal("expected error for non-20ms frame size")
	}
}

func TestPionOpusEncoder_Roundtrip(t *testing.T) {
	enc, err := NewOpusEncoder(48000, 2, 960)
	if err != nil {
		t.Fatalf("NewOpusEncoder failed: %v", err)
	}
	defer func() { _ = enc.Close() }()

	// 440 Hz sine, stereo interleaved, one 20 ms frame.
	pcm := make([]int16, 960*2)
	for i := 0; i < 960; i++ {
		v := int16(20000 * math.Sin(2*math.Pi*440*float64(i)/48000))
		pcm[i*2] = v
		pcm[i*2+1] = v
	}

	packet, err := enc.Encode(pcm)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}
	if len(packet) == 0 {
		t.Fatal("empty packet")
	}
	t.Logf("encoded 960 stereo samples into %d bytes", len(packet))

	dec, err := decode.NewOpus(audio.Format{Codec: "opus", Channels: 2, SampleRate: 48000, BitDepth: 16})
	if err != nil {
		t.Fatalf("decoder setup failed: %v", err)
	}
	defer func() { _ = dec.Close() }()

	out, err := dec.Decode(packet)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}
	if len(out) != 960*2 {
		t.Fatalf("expected 1920 samples, got %d", len(out))
	}

	// Rough fidelity: decoded energy must be in the same ballpark
	// as the input (guards against silent/garbage output).
	var inE, outE float64
	for i := range pcm {
		inE += float64(pcm[i]) * float64(pcm[i])
		v := float64(out[i]) / 256.0 // int32 24-bit range back to int16 scale
		outE += v * v
	}
	ratio := outE / inE
	t.Logf("energy ratio out/in: %.3f", ratio)
	if ratio < 0.25 || ratio > 4.0 {
		t.Fatalf("decoded energy out of plausible range (ratio %.3f)", ratio)
	}
}

func TestPionOpusEncoder_RejectsShortFrame(t *testing.T) {
	enc, err := NewOpusEncoder(48000, 1, 960)
	if err != nil {
		t.Fatalf("NewOpusEncoder failed: %v", err)
	}
	defer func() { _ = enc.Close() }()

	if _, err := enc.Encode(make([]int16, 100)); err == nil {
		t.Fatal("expected error for short frame")
	}
}
