// ABOUTME: Opus audio decoder (pure-Go pion/opus, RFC 6716)
// ABOUTME: Decodes Opus audio to int32 samples on all builds

package decode

import (
	"fmt"

	"github.com/pion/opus"

	"github.com/Sendspin/sendspin-go/pkg/audio"
)

// OpusDecoder decodes Opus packets to int32 samples using a pure-Go
// RFC 6716 implementation, with output semantics matching libopus:
// int16 decode, left-justified into 24-bit range.
type OpusDecoder struct {
	decoder  opus.Decoder
	format   audio.Format
	pcm16Buf []int16 // reusable decode buffer to avoid per-frame allocation
}

func NewOpus(format audio.Format) (Decoder, error) {
	if format.Codec != "opus" {
		return nil, fmt.Errorf("invalid codec for Opus decoder: %s", format.Codec)
	}

	dec, err := opus.NewDecoderWithOutput(format.SampleRate, format.Channels)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus decoder: %w", err)
	}

	return &OpusDecoder{
		decoder:  dec,
		format:   format,
		pcm16Buf: make([]int16, 5760*format.Channels),
	}, nil
}

func (d *OpusDecoder) Decode(data []byte) ([]int32, error) {
	// Reuse pre-allocated int16 buffer for decode (avoids 23KB alloc per frame)
	n, err := d.decoder.DecodeToInt16(data, d.pcm16Buf)
	if err != nil {
		return nil, fmt.Errorf("opus decode failed: %w", err)
	}

	actualSamples := n * d.format.Channels
	pcm32 := make([]int32, actualSamples)
	for i := 0; i < actualSamples; i++ {
		pcm32[i] = audio.SampleFromInt16(d.pcm16Buf[i])
	}
	return pcm32, nil
}

func (d *OpusDecoder) Close() error {
	return nil
}
