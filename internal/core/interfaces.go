package core

import "context"

// --- Audio ---

// AudioCapturer reads raw sound from the microphone (malgo implementation).
type AudioCapturer interface {
	Start(ctx context.Context) (<-chan []byte, error)
	Stop() error
}

// AudioPlayer plays synthesized assistant response.
type AudioPlayer interface {
	Play(ctx context.Context, pcm []byte, sampleRate int) error
}

// VAD determines whether the user has stopped speaking.
type VAD interface {
	IsSpeech(frame []byte) bool
}

// --- ML sidecar (wrappers around gRPC client) ---

type STTClient interface {
	Transcribe(ctx context.Context, audio []byte, sampleRate int) (text string, err error)
}

type TTSClient interface {
	Synthesize(ctx context.Context, text, voice string) (pcm []byte, sampleRate int, err error)
}

// SidecarManager starts and monitors the Python process.
type SidecarManager interface {
	Start(ctx context.Context) error
	Stop() error
	Addr() string
}

// --- Skills / plugins ---

type SkillInput struct {
	RawText string
	Intent  string
	Args    map[string]string
}

type SkillResult struct {
	DisplayText    string
	SpokenResponse string
}

type Skill interface {
	Name() string
	Match(text string) (matched bool, confidence float64)
	Execute(ctx context.Context, in SkillInput) (SkillResult, error)
}

type Router interface {
	Register(skill Skill)
	Route(ctx context.Context, text string) (SkillResult, error)
}

// --- Orchestration ---

type Orchestrator interface {
	Run(ctx context.Context) error
}
