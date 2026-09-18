from __future__ import annotations

import io
import wave
from typing import Optional, Tuple

try:
    from faster_whisper import WhisperModel
except ImportError:  # pragma: no cover - dependency may be absent during initial scaffold
    WhisperModel = None


class STTEngine:
    def __init__(self, model_size: str = "small", device: str = "cpu", compute_type: str = "int8") -> None:
        self.model_size = model_size
        self.device = device
        self.compute_type = compute_type
        self.model = None
        if WhisperModel is not None:
            self.model = WhisperModel(model_size, device=device, compute_type=compute_type)

    def is_ready(self) -> bool:
        return self.model is not None

    def transcribe_pcm(self, audio: bytes, sample_rate: int = 16000, language: str = "") -> Tuple[str, float, str]:
        if not audio:
            return "", 0.0, language or "und"
        if self.model is None:
            raise RuntimeError("faster-whisper model is not loaded")

        wav_buffer = io.BytesIO()
        with wave.open(wav_buffer, "wb") as wf:
            wf.setnchannels(1)
            wf.setsampwidth(2)
            wf.setframerate(sample_rate)
            wf.writeframes(audio)

        wav_bytes = wav_buffer.getvalue()
        with wave.open(io.BytesIO(wav_bytes), "rb") as wf:
            frames = wf.readframes(wf.getnframes())

        # Faster Whisper accepts a file path or a file-like object in some versions.
        # We keep a temporary round trip through bytes to ensure compatibility.
        tmp = io.BytesIO(wav_bytes)
        segments, info = self.model.transcribe(
            tmp,
            language=language or None,
            task="transcribe",
            beam_size=1,
            without_timestamps=True,
        )

        transcript_parts = []
        confidence_values = []
        for segment in segments:
            text = (segment.text or "").strip()
            if text:
                transcript_parts.append(text)
            conf = getattr(segment, "avg_logprob", None)
            if conf is not None:
                confidence_values.append(max(0.0, min(1.0, float(conf))))

        text = " ".join(transcript_parts).strip()
        if not text:
            return "", 0.0, info.language or (language or "und")

        confidence = 0.5
        if confidence_values:
            confidence = sum(confidence_values) / len(confidence_values)
            confidence = max(0.0, min(1.0, confidence))

        return text, confidence, info.language or (language or "und")
