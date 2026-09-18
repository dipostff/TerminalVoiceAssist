package skills

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/you/voice-assistant/internal/core"
)

type ShellSkill struct {
	name    string
	pattern *regexp.Regexp
	command string
}

func NewShellSkill(name, pattern, command string) (*ShellSkill, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &ShellSkill{name: name, pattern: re, command: command}, nil
}

func (s *ShellSkill) Name() string { return s.name }

func (s *ShellSkill) Match(text string) (bool, float64) {
	if s.pattern == nil {
		return false, 0
	}

	normalized := strings.TrimSpace(text)
	normalized = strings.TrimRight(normalized, ".,!?:;\"'“”‘’")
	if normalized == "" {
		return false, 0
	}
	if s.pattern.MatchString(normalized) {
		return true, 0.9
	}
	return false, 0
}

func (s *ShellSkill) Execute(ctx context.Context, in core.SkillInput) (core.SkillResult, error) {
	cmdStr := s.command
	if strings.Contains(cmdStr, "{{.dir}}") {
		if dir, ok := in.Args["dir"]; ok {
			cmdStr = strings.ReplaceAll(cmdStr, "{{.dir}}", dir)
		} else {
			cmdStr = strings.ReplaceAll(cmdStr, "{{.dir}}", ".")
		}
	}

	cmd := exec.CommandContext(ctx, "bash", "-lc", cmdStr)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return core.SkillResult{DisplayText: string(exitErr.Stderr), SpokenResponse: string(exitErr.Stderr)}, nil
		}
		return core.SkillResult{}, err
	}

	resultText := strings.TrimSpace(string(out))
	if resultText == "" {
		resultText = "command completed"
	}
	return core.SkillResult{DisplayText: resultText, SpokenResponse: resultText}, nil
}

func ExampleShellSkill() {
	_ = fmt.Sprintf("%v", "example")
}
