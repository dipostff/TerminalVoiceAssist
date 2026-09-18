package router

import (
	"context"
	"fmt"
	"sort"

	"github.com/you/voice-assistant/internal/core"
)

type Router struct {
	skills []core.Skill
}

func New() *Router {
	return &Router{}
}

func (r *Router) Register(skill core.Skill) {
	if skill == nil {
		return
	}
	r.skills = append(r.skills, skill)
}

func (r *Router) Route(ctx context.Context, text string) (core.SkillResult, error) {
	if text == "" {
		return core.SkillResult{}, fmt.Errorf("empty input")
	}

	type candidate struct {
		skill      core.Skill
		confidence float64
	}

	var matches []candidate
	for _, skill := range r.skills {
		matched, confidence := skill.Match(text)
		if matched {
			matches = append(matches, candidate{skill: skill, confidence: confidence})
		}
	}
	if len(matches) == 0 {
		return core.SkillResult{DisplayText: "no matching skill"}, nil
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].confidence > matches[j].confidence
	})

	best := matches[0]
	result, err := best.skill.Execute(ctx, core.SkillInput{RawText: text, Args: map[string]string{}})
	if err != nil {
		return core.SkillResult{}, err
	}
	return result, nil
}
