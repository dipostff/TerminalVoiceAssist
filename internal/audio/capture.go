package audio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gen2brain/malgo"
)

const (
	defaultSampleRate = 16000
	defaultFrameSize  = 2048
)

// CaptureManager reads PCM16LE mono audio from the default microphone and emits frames on a channel.
type CaptureManager struct {
	ctx    context.Context
	cancel context.CancelFunc
	done   chan struct{}
	mu     sync.Mutex
	device *malgo.Device
	out    chan []byte
	stop   chan struct{}
	stopped bool
	stopOnce sync.Once
}

func NewCaptureManager() *CaptureManager {
	return &CaptureManager{
		out:  make(chan []byte, 32),
		stop: make(chan struct{}),
	}
}

func (c *CaptureManager) Start(ctx context.Context) (<-chan []byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.device != nil {
		return c.out, nil
	}

	captureCtx, cancel := context.WithCancel(ctx)
	c.ctx = captureCtx
	c.cancel = cancel
	c.done = make(chan struct{})

	malCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(msg string) {})
	if err != nil {
		return nil, fmt.Errorf("init malgo context: %w", err)
	}
	defer func() {
		if err != nil {
			_ = malCtx.Uninit()
			malCtx.Free()
		}
	}()

	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = defaultSampleRate
	deviceConfig.PeriodSizeInFrames = defaultFrameSize
	deviceConfig.Alsa.NoMMap = 1

	var frameBuf []byte
	callbacks := malgo.DeviceCallbacks{
		Data: func(_, pInput []byte, framecount uint32) {
			if len(pInput) == 0 {
				return
			}
			frameBuf = append(frameBuf, pInput...)
			for len(frameBuf) >= defaultFrameSize*2 {
				chunk := make([]byte, defaultFrameSize*2)
				copy(chunk, frameBuf[:defaultFrameSize*2])
				frameBuf = frameBuf[defaultFrameSize*2:]
				select {
				case c.out <- chunk:
				default:
				}
			}
		},
	}

	device, err := malgo.InitDevice(malCtx.Context, deviceConfig, callbacks)
	if err != nil {
		return nil, fmt.Errorf("init audio device: %w", err)
	}

	if err = device.Start(); err != nil {
		device.Uninit()
		return nil, fmt.Errorf("start audio device: %w", err)
	}

	c.device = device
	malCtx.Free()
	go func() {
		defer close(c.done)
		<-captureCtx.Done()
		c.mu.Lock()
		if c.device != nil {
			c.device = nil
		}
		c.mu.Unlock()
	}()

	return c.out, nil
}

func (c *CaptureManager) Stop() error {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return nil
	}
	c.stopped = true
	c.mu.Unlock()

	c.stopOnce.Do(func() {
		if c.cancel != nil {
			c.cancel()
		}
		if c.done != nil {
			<-c.done
		}
		c.mu.Lock()
		device := c.device
		c.device = nil
		c.mu.Unlock()
		if device != nil {
			device.Uninit()
		}
		if c.stop != nil {
			select {
			case <-c.stop:
			default:
				close(c.stop)
			}
		}
	})
	return nil
}

func (c *CaptureManager) waitForFrames(timeout time.Duration) []byte {
	select {
	case frame := <-c.out:
		return frame
	case <-time.After(timeout):
		return nil
	}
}
