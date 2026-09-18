package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/you/voice-assistant/internal/audio"
	"github.com/you/voice-assistant/internal/config"
	"github.com/you/voice-assistant/internal/orchestrator"
	"github.com/you/voice-assistant/internal/router"
	"github.com/you/voice-assistant/internal/sidecar"
	"github.com/you/voice-assistant/internal/skills"
)

func main() {
	cfg := config.Load()

	manager := sidecar.NewManager(cfg.SidecarAddr)
	sttClient, err := sidecar.NewSTTClient(cfg.SidecarAddr)
	if err != nil {
		log.Fatalf("create STT client: %v", err)
	}
	defer func() { _ = sttClient.Close() }()

	ttsClient, err := sidecar.NewTTSClient(cfg.SidecarAddr)
	if err != nil {
		log.Fatalf("create TTS client: %v", err)
	}
	defer func() { _ = ttsClient.Close() }()

	listFilesSkill, err := skills.NewShellSkill("list_files", `(?i)^(?:покажи|покажите|show|list|display)\s+(?:me\s+)?(?:the\s+)?(?:(?:список)|(?:files?|directory|dir)|(?:файл|файлы|файлов))(?:\s+in\s+(?P<dir>.+))?.*$`, "ls -la {{.dir}}")
	if err != nil {
		log.Fatalf("create list-files skill: %v", err)
	}
	gitStatusSkill, err := skills.NewShellSkill("git_status", `(?i)^(?:статус\s+git|статус\s+гит|git\s+status|git\s+status\s+short|check\s+git\s+status|show\s+git\s+status|status\s+git)$`, "git status --short")
	if err != nil {
		log.Fatalf("create git status skill: %v", err)
	}
	englishHelloSkill, err := skills.NewShellSkill("hello", `(?i)^(?:hello|hi|hey|привет|здравствуй|здрасьте)(?:\s+.*)?$`, "printf '%s' 'hello from terminal voice assistant' ")
	if err != nil {
		log.Fatalf("create english hello skill: %v", err)
	}
	englishStartSkill, err := skills.NewShellSkill("start_firefox", `(?i)^(?:start|launch|open|запусти|открой)\s+(?:firefox|browser|браузер)(?:\s+.*)?$`, "firefox")
	if err != nil {
		log.Fatalf("create english start skill: %v", err)
	}
	englishByeSkill, err := skills.NewShellSkill("bye", `(?i)^(?:bye|goodbye|see you|good night|пока|до свидания)(?:\s+.*)?$`, "printf '%s' 'goodbye' ")
	if err != nil {
		log.Fatalf("create english bye skill: %v", err)
	}

	r := router.New()
	r.Register(listFilesSkill)
	r.Register(gitStatusSkill)
	r.Register(englishHelloSkill)
	r.Register(englishStartSkill)
	r.Register(englishByeSkill)

	capture := audio.NewCaptureManager()
	player := audio.NewPlayer()
	app := orchestrator.New(manager, sttClient, capture, r, ttsClient, player, cfg.Voice)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Printf("starting assistant on %s\n", cfg.SidecarAddr)
	if err := app.Run(ctx); err != nil {
		log.Fatalf("orchestrator failed: %v", err)
	}
}
