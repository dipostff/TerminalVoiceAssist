package orchestrator

import "testing"

func TestIsLikelyNoiseRejectsRussianFillers(t *testing.T) {
	cases := []string{"you", "uh", "ну", "а", "э", "мм"}
	for _, txt := range cases {
		if !isLikelyNoise(txt) {
			t.Fatalf("expected %q to be treated as noise", txt)
		}
	}
}

func TestIsLikelyNoiseKeepsRealCommands(t *testing.T) {
	cases := []string{"привет", "hello", "show files", "git status"}
	for _, txt := range cases {
		if isLikelyNoise(txt) {
			t.Fatalf("expected %q to be treated as command text, not noise", txt)
		}
	}
}

func TestIsLikelyNoiseRejectsSubtitleNoise(t *testing.T) {
	cases := []string{"субтитры субтитров а.синецкая корректор а.егорова", "subtitle track overlay", "субтитры по субтитрам"}
	for _, txt := range cases {
		if !isLikelyNoise(txt) {
			t.Fatalf("expected %q to be treated as background subtitle noise", txt)
		}
	}
}
