package skills

import "testing"

func TestShellSkillMatchIgnoresTrailingPunctuation(t *testing.T) {
	skill, err := NewShellSkill("hello", `(?i)^(?:hello|hi|hey|привет|здравствуй|здрасьте)(?:\s+.*)?$`, "printf '%s' 'hi'")
	if err != nil {
		t.Fatalf("create skill: %v", err)
	}

	for _, input := range []string{"Привет.", "Hello.", "привет!", "hello,"} {
		if ok, _ := skill.Match(input); !ok {
			t.Fatalf("expected %q to match after punctuation normalization", input)
		}
	}
}
