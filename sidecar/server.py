from __future__ import annotations

import asyncio
import logging
import os
import sys
from concurrent import futures
from pathlib import Path

import grpc

ROOT_DIR = Path(__file__).resolve().parent.parent
if str(ROOT_DIR) not in sys.path:
    sys.path.insert(0, str(ROOT_DIR))

try:
    from sidecar.stt_engine import STTEngine
    from sidecar.tts_engine import TTSEngine
except ModuleNotFoundError:
    from stt_engine import STTEngine
    from tts_engine import TTSEngine

PROTO_DIR = Path(__file__).resolve().parent / "proto"
if str(PROTO_DIR) not in sys.path:
    sys.path.insert(0, str(PROTO_DIR))

import assistant_pb2
import assistant_pb2_grpc

logger = logging.getLogger(__name__)


class MLServiceServicer(assistant_pb2_grpc.MLServiceServicer):
    def __init__(self):
        self._ready = False
        self._loaded_models = []
        try:
            self.stt_engine = STTEngine(model_size="small", device="cpu", compute_type="int8")
            self.tts_engine = TTSEngine(voice="default")
            if self.stt_engine.is_ready():
                self._loaded_models.append("faster-whisper-small-cpu-int8")
            else:
                self._loaded_models.append("faster-whisper-not-loaded")
            if self.tts_engine._available:
                self._loaded_models.append("piper-tts")
            else:
                self._loaded_models.append("piper-fallback")
            self._ready = self.stt_engine.is_ready() or self.tts_engine is not None
        except Exception as exc:  # pragma: no cover - expected until dependency is installed
            logger.warning("STT/TTS engine init failed: %s", exc)
            self.stt_engine = None
            self.tts_engine = TTSEngine(voice="default")
            self._ready = False
            self._loaded_models = ["faster-whisper-not-loaded", "piper-fallback"]

    async def Transcribe(self, request, context):
        if not request.audio:
            return assistant_pb2.TranscribeResponse(
                text="",
                confidence=0.0,
                detected_language=request.language or "und",
            )
        if self.stt_engine is None:
            raise RuntimeError("STT engine is not loaded")

        text, confidence, language = self.stt_engine.transcribe_pcm(
            request.audio,
            sample_rate=request.sample_rate or 16000,
            language=request.language or "",
        )
        return assistant_pb2.TranscribeResponse(
            text=text,
            confidence=float(confidence),
            detected_language=language,
        )

    async def StreamTranscribe(self, request_iterator, context):
        raise NotImplementedError("StreamTranscribe is not implemented yet")

    async def Synthesize(self, request, context):
        if not request.text:
            return assistant_pb2.SynthesizeResponse(audio=b"", sample_rate=16000, format="wav")
        if self.tts_engine is None:
            raise RuntimeError("TTS engine is not loaded")
        audio, sample_rate = self.tts_engine.synthesize(request.text, request.voice or "default")
        return assistant_pb2.SynthesizeResponse(audio=audio, sample_rate=sample_rate, format="wav")

    def HealthCheck(self, request, context):
        return assistant_pb2.HealthResponse(
            ok=self._ready,
            version="0.1.0",
            loaded_models=self._loaded_models,
        )


async def serve() -> None:
    server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=4))
    assistant_pb2_grpc.add_MLServiceServicer_to_server(MLServiceServicer(), server)
    server.add_insecure_port("0.0.0.0:50051")
    await server.start()
    logger.info("sidecar server started on 0.0.0.0:50051")
    try:
        await server.wait_for_termination()
    except asyncio.CancelledError:
        logger.info("shutdown signal received, stopping gRPC server")
    finally:
        await server.stop(grace=1)


if __name__ == "__main__":
    logging.basicConfig(level=logging.INFO)
    try:
        asyncio.run(serve())
    except KeyboardInterrupt:
        logger.info("keyboard interrupt received")
