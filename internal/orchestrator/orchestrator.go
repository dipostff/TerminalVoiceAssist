package orchestrator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/you/voice-assistant/internal/audio"
	"github.com/you/voice-assistant/internal/core"
	"github.com/you/voice-assistant/internal/router"
	"github.com/you/voice-assistant/internal/sidecar"
)

type Orchestrator struct {
	manager *sidecar.Manager
	client  *sidecar.STTClient
	tts     *sidecar.TTSClient
	player  *audio.Player
	capture *audio.CaptureManager
	router  *router.Router
	vad     *audio.EnergyVAD
	voice   string
}

func New(manager *sidecar.Manager, client *sidecar.STTClient, capture *audio.CaptureManager, r *router.Router, tts *sidecar.TTSClient, player *audio.Player, voice string) *Orchestrator {
	return &Orchestrator{
		manager: manager,
		client:  client,
		tts:     tts,
		player:  player,
		capture: capture,
		router:  r,
		vad:     audio.NewEnergyVAD(420, 250),
		voice:   voice,
	}
}

func (o *Orchestrator) Run(ctx context.Context) error {
	if o.manager == nil {
		return fmt.Errorf("sidecar manager is nil")
	}
	if o.client == nil {
		return fmt.Errorf("stt client is nil")
	}
	if o.capture == nil {
		return fmt.Errorf("audio capture is nil")
	}
	if o.router == nil {
		return fmt.Errorf("router is nil")
	}

	fmt.Println("starting voice workflow")
	if err := o.manager.Start(ctx); err != nil {
		return err
	}
	defer func() {
		_ = o.manager.Stop()
	}()

	stream, err := o.capture.Start(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = o.capture.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case frame := <-stream:
			if frame == nil {
				continue
			}
			if !o.vad.IsSpeech(frame) {
				continue
			}
			buf := make([]byte, 0, 8192)
			buf = append(buf, frame...)
			lastSpeech := time.Now()
			for {
				select {
				case <-ctx.Done():
					return nil
				case next := <-stream:
					if next == nil {
						continue
					}
					if o.vad.IsSpeech(next) {
						buf = append(buf, next...)
						lastSpeech = time.Now()
						continue
					}
					if len(buf) > 0 && time.Since(lastSpeech) > 180*time.Millisecond {
						goto flush
					}
				case <-time.After(40 * time.Millisecond):
					if len(buf) > 0 && time.Since(lastSpeech) > 180*time.Millisecond {
						goto flush
					}
				}
				if len(buf) > 16000*4 {
					goto flush
				}
			}
		flush:
			if len(buf) == 0 {
				continue
			}
			text, err := o.client.Transcribe(ctx, buf, 16000)
			if err != nil {
				fmt.Printf("transcribe error: %v\n", err)
				continue
			}
			text = strings.TrimSpace(text)
			if text == "" {
				continue
			}
			if isLikelyNoise(text) {
				continue
			}
			if !isMeaningfulCommand(text) {
				continue
			}
			result, err := o.router.Route(ctx, text)
			if err != nil {
				fmt.Printf("route error: %v\n", err)
				continue
			}
			fmt.Printf("text: %s\nresult: %s\n", text, result.DisplayText)
			if o.tts != nil && o.player != nil && result.SpokenResponse != "" {
				pcm, sampleRate, synthErr := o.tts.Synthesize(ctx, result.SpokenResponse, o.voice)
				if synthErr != nil {
					fmt.Printf("tts error: %v\n", synthErr)
					continue
				}
				if err := o.player.Play(ctx, pcm, sampleRate); err != nil {
					fmt.Printf("playback error: %v\n", err)
				}
			}
		}
	}
}

func isMeaningfulCommand(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return false
	}

	allowlist := map[string]struct{}{
		"привет": {},
		"пока":   {},
		"hello":  {},
		"hi":     {},
		"bye":    {},
	}
	if _, ok := allowlist[normalized]; ok {
		return true
	}

	words := strings.Fields(normalized)
	if len(words) >= 2 {
		return true
	}
	return len(normalized) >= 6 && len(normalized) <= 80
}

func isLikelyNoise(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return true
	}

	noiseWords := map[string]struct{}{
		"you":        {},
		"uh":         {},
		"um":         {},
		"huh":        {},
		"ugh":        {},
		"ah":         {},
		"eh":         {},
		"mm":         {},
		"hm":         {},
		"ну":         {},
		"ээ":         {},
		"э":          {},
		"а":          {},
		"м":          {},
		"хм":         {},
		"эм":         {},
		"нуу":        {},
		"угу":        {},
		"аха":        {},
		"ага":        {},
		"okay":       {},
		"subtitle":   {},
		"субтитры": {},
	}
	if _, ok := noiseWords[normalized]; ok {
		return true
	}
	for _, marker := range []string{"субтитры", "subtitle", "новикова", "корректор", "синецкая", "егорова", "подписи", "caption", "captions"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}

	if strings.Contains(normalized, "субтитры") || strings.Contains(normalized, "subtitle") {
		words := strings.Fields(normalized)
		if len(words) >= 3 {
			return true
		}
	}

	if strings.ContainsAny(normalized, "абвгдеёжзийклмнопрстуфхцчшщъыьэюя") {
		return len(normalized) <= 2 && !strings.ContainsAny(normalized, "abcdefghijklmnopqrstuvwxyz")
	}

	return len(normalized) <= 2 && !strings.ContainsAny(normalized, "abcdefghijklmnopqrstuvwxyz")
}

var _ core.Orchestrator = (*Orchestrator)(nil)
