package audio

import (
	"math"
	"time"
)

type EnergyVAD struct {
	threshold float64
	window    int
	silenceMs int
	lastSpeech time.Time
}

func NewEnergyVAD(threshold int, silenceMs int) *EnergyVAD {
	return &EnergyVAD{threshold: float64(threshold), window: 160, silenceMs: silenceMs}
}

func (v *EnergyVAD) IsSpeech(frame []byte) bool {
	if len(frame) == 0 || len(frame) < 2 {
		return false
	}

	var sum float64
	for i := 0; i+1 < len(frame); i += 2 {
		sample := int16(frame[i]) | (int16(frame[i+1]) << 8)
		sum += math.Abs(float64(sample))
	}

	avg := sum / float64(len(frame)/2)
	if avg > v.threshold {
		v.lastSpeech = time.Now()
		return true
	}
	return false
}

func (v *EnergyVAD) IsSpeechWithWindow(frame []byte) bool {
	if !v.IsSpeech(frame) {
		return false
	}
	if len(frame) < 256 {
		return false
	}
	return true
}

func (v *EnergyVAD) SilenceExceeded() bool {
	return time.Since(v.lastSpeech) > time.Duration(v.silenceMs)*time.Millisecond
}
