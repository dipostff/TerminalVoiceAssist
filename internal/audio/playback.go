package audio

import (
	"context"
	"fmt"
	"io"
	"math"

	"github.com/gen2brain/malgo"
)

// Player plays PCM audio via the default output device.
type Player struct {
	ctx    context.Context
	cancel context.CancelFunc
	device *malgo.Device
}

func NewPlayer() *Player {
	return &Player{}
}

func (p *Player) Play(ctx context.Context, pcm []byte, sampleRate int) error {
	if len(pcm) == 0 {
		return nil
	}
	if sampleRate <= 0 {
		sampleRate = 16000
	}

	malCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(string) {})
	if err != nil {
		return fmt.Errorf("init malgo context: %w", err)
	}
	defer func() {
		_ = malCtx.Uninit()
		malCtx.Free()
	}()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Playback)
	deviceConfig.Playback.Format = malgo.FormatS16
	deviceConfig.Playback.Channels = 1
	deviceConfig.SampleRate = uint32(sampleRate)
	deviceConfig.Alsa.NoMMap = 1

	reader := &bytesReader{data: pcm}
	callbacks := malgo.DeviceCallbacks{
		Data: func(pOutputSample, _ []byte, framecount uint32) {
			_, _ = io.ReadFull(reader, pOutputSample)
		},
	}

	device, err := malgo.InitDevice(malCtx.Context, deviceConfig, callbacks)
	if err != nil {
		return fmt.Errorf("init playback device: %w", err)
	}
	defer device.Uninit()

	if err := device.Start(); err != nil {
		return fmt.Errorf("start playback device: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		for device.IsStarted() {
			if reader.Len() == 0 {
				return nil
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
		}
	}
	return nil
}

type bytesReader struct { data []byte }

func (r *bytesReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	count := copy(p, r.data)
	r.data = r.data[count:]
	return count, nil
}

func (r *bytesReader) Len() int { return len(r.data) }

func _demo() {
	_ = math.Pi
}
