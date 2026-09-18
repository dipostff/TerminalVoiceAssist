package audio

import "testing"

func TestEnergyVADDetectsSpeech(t *testing.T) {
	v := NewEnergyVAD(300, 250)

	loud := make([]byte, 160)
	for i := range loud {
		if i%2 == 0 {
			loud[i] = 0xFF
		} else {
			loud[i] = 0x7F
		}
	}
	if !v.IsSpeech(loud) {
		t.Fatal("expected loud signal to be classified as speech")
	}
}

func TestEnergyVADRejectsSilence(t *testing.T) {
	v := NewEnergyVAD(300, 250)
	quiet := make([]byte, 160)
	if v.IsSpeech(quiet) {
		t.Fatal("expected silence to be classified as non-speech")
	}
}
