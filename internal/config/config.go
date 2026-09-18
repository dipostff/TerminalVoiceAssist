package config

import (
	"os"
	"strings"
)

type Config struct {
	SidecarAddr string
	Voice       string
	Hotkey      string
	Language    string
}

func Default() Config {
	return Config{
		SidecarAddr: "127.0.0.1:50051",
		Voice:       "ru_RU-irina-medium",
		Hotkey:      "Ctrl+Alt+V",
		Language:    "ru",
	}
}

func Load() Config {
	cfg := Default()
	if addr := strings.TrimSpace(os.Getenv("VOICE_ASSISTANT_SIDECAR_ADDR")); addr != "" {
		cfg.SidecarAddr = addr
	}
	if voice := strings.TrimSpace(os.Getenv("VOICE_ASSISTANT_VOICE")); voice != "" {
		cfg.Voice = voice
	}
	if hotkey := strings.TrimSpace(os.Getenv("VOICE_ASSISTANT_HOTKEY")); hotkey != "" {
		cfg.Hotkey = hotkey
	}
	if language := strings.TrimSpace(os.Getenv("VOICE_ASSISTANT_LANGUAGE")); language != "" {
		cfg.Language = language
	}
	return cfg
}
