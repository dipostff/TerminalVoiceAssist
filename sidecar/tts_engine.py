from __future__ import annotations

import io
import os
import shutil
import subprocess
import tempfile
import wave
from pathlib import Path
from typing import Tuple


class TTSEngine:
    def __init__(self, voice: str = "default") -> None:
        self.voice = voice
        self._available = self._check_available()

    def _check_available(self) -> bool:
        return shutil.which("piper") is not None or shutil.which("piper-tts") is not None

    def _fallback_silence(self, duration_seconds: float = 0.2, sample_rate: int = 16000) -> Tuple[bytes, int]:
        frame_count = int(sample_rate * duration_seconds)
        buffer = bytearray()
        for _ in range(frame_count):
            buffer.extend((0).to_bytes(2, byteorder="little", signed=True))
        return bytes(buffer), sample_rate

    def synthesize(self, text: str, voice: str = "default") -> Tuple[bytes, int]:
        if not text:
            return b"", 16000

        if not self._available:
            return self._fallback_silence()

        exe = shutil.which("piper") or shutil.which("piper-tts")
        if exe is None:
            return self._fallback_silence()

        model_path = Path("/usr/share/piper/model.onnx")
        if not model_path.exists():
            return self._fallback_silence()

        with tempfile.TemporaryDirectory() as tmpdir:
            wav_path = os.path.join(tmpdir, "out.wav")
            cmd = [
                exe,
                "--model",
                str(model_path),
                "--output_file",
                wav_path,
            ]
            result = subprocess.run(cmd, input=text.encode("utf-8"), capture_output=True, check=False)
            if result.returncode != 0:
                return self._fallback_silence()
            with open(wav_path, "rb") as fh:
                data = fh.read()

        return data, 16000
