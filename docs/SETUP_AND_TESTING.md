# Настройка, запуск и тестирование TerminalVoiceAssist

Эта инструкция описывает, как подготовить проект, запустить Go-часть и Python sidecar, а также проверить базовую работоспособность до live-тестов на реальном микрофоне.

## 1. Требования

Перед запуском убедитесь, что установлены:

- Go 1.25+
- Python 3.10+
- `protoc` (protobuf compiler)
- `protoc-gen-go`
- `protoc-gen-go-grpc`
- системные аудиобиблиотеки для Linux (если используются ALSA / PulseAudio / PipeWire)

Проверка:

```bash
go version
python3 --version
protoc --version
which protoc-gen-go
which protoc-gen-go-grpc
```

Если `protoc` или плагины отсутствуют, установите их:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Для Debian/Ubuntu-подобной системы protobuf-компилятор обычно ставится так:

```bash
sudo apt update
sudo apt install -y protobuf-compiler
```

## 2. Установка зависимостей Python

В корне проекта:

```bash
python3 -m pip install --break-system-packages -r sidecar/requirements.txt
```

Содержимое `sidecar/requirements.txt` включает:

```txt
grpcio>=1.60.0
grpcio-tools>=1.60.0
faster-whisper>=1.0.0
```

## 3. Генерация gRPC-кода

В проекте используется схема из `proto/assistant.proto`.

Чтобы сгенерировать Go и Python stubs:

```bash
chmod +x scripts/gen-proto.sh
./scripts/gen-proto.sh
```

Если все зависимости установлены корректно, скрипт создаст файлы в:

- `internal/sidecar/proto/`
- `sidecar/proto/`

## 4. Проверка сборки Go

В корне проекта:

```bash
go test ./...
```

Дополнительно можно проверить синтаксис Python-модулей:

```bash
python3 -m py_compile sidecar/server.py sidecar/stt_engine.py sidecar/tts_engine.py
```

## 5. Запуск sidecar вручную

Если нужно запустить Python-сервис отдельно:

```bash
cd /path/to/TerminalVoiceAssist
python3 sidecar/server.py
```

По умолчанию sidecar слушает порт `127.0.0.1:50051`.

## 6. Запуск основного приложения

В корне проекта:

```bash
go run ./cmd/assistant
```

Это запустит:

- менеджер sidecar;
- STT/TTS клиентов;
- захват аудио;
- маршрутизатор и skills;
- orchestrator, который собирает весь поток.

## 7. Переменные окружения

В проекте есть три основных параметра конфигурации:

```bash
export VOICE_ASSISTANT_SIDECAR_ADDR=127.0.0.1:50051
export VOICE_ASSISTANT_VOICE=ru_RU-irina-medium
export VOICE_ASSISTANT_HOTKEY='Ctrl+Alt+V'
```

Затем можно запускать приложение обычной командой:

```bash
go run ./cmd/assistant
```

## 8. Проверка health-check

Если sidecar уже запущен, можно проверить, что gRPC сервер отвечает:

```bash
python3 - <<'PY'
import grpc
import sys
sys.path.insert(0, 'sidecar/proto')
import assistant_pb2
import assistant_pb2_grpc

channel = grpc.insecure_channel('127.0.0.1:50051')
stub = assistant_pb2_grpc.MLServiceStub(channel)
resp = stub.HealthCheck(assistant_pb2.HealthRequest())
print(resp)
PY
```

Ожидается ответ со статусом `ok` и списком загруженных моделей.

## 9. Базовый smoke test

Чтобы быстро проверить, что Python sidecar и Go-код не падают на синтаксисе и на простом вызове, используйте:

```bash
python3 -m py_compile sidecar/server.py sidecar/stt_engine.py sidecar/tts_engine.py
go test ./...
```

Это не заменяет live-проверку на реальном микрофоне, но позволяет быстро убедиться, что проект не сломан после изменений.

## 10. Проверка TTS

ТTS в проекте использует `piper` при наличии в системе. Если бинарник доступен:

```bash
which piper || which piper-tts
```

Если `piper` отсутствует, проект должен падать в fallback-режим, а это нужно учитывать при live-тестах.

## 11. Проверка audio loop

Реальный end-to-end прогон на железе требует:

- корректно работающего микрофона;
- доступных ALSA/PulseAudio/PipeWire устройств;
- работающего `malgo`-захвата;
- корректного VAD и STT;
- действительного TTS-синтеза через Piper или fallback.

Для live-тестов обычно запускают приложение и проговаривают команду в микрофон, затем проверяют:

- появился ли транскрипт;
- сработал ли skill;
- появился ли синтезированный ответ;
- прозвучал ли аудиовыход.

## 12. Возможные проблемы и диагностика

### Проблема: `protoc` не найден

```bash
sudo apt install -y protobuf-compiler
```

### Проблема: `protoc-gen-go` или `protoc-gen-go-grpc` не найден

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Проблема: sidecar не стартует

Проверьте:

- нет ли конфликта порта `50051`;
- установлен ли Python-пакет `grpcio`;
- корректный ли путь до файла `sidecar/server.py`.

### Проблема: TTS не работает

Проверьте:

```bash
which piper || which piper-tts
ls /usr/share/piper/
```

### Проблема: звук не захватывается

Проверьте:

- права доступа к аудиоустройству;
- наличие микрофона в системе;
- корректность параметров `Capture` и `SampleRate` в `internal/audio/capture.go`.

## 13. Рекомендуемый порядок работы

Для разработки и отладки проекта рекомендуется идти в таком порядке:

1. установить зависимости;
2. сгенерировать protobuf;
3. проверить `go test ./...`;
4. проверить Python compile;
5. запустить sidecar вручную;
6. проверить `HealthCheck`;
7. проверить STT на тестовом PCM;
8. запустить live-цикл через `go run ./cmd/assistant`;
9. отлаживать real-audio и TTS по мере появления ошибок;
10. только после этого добавлять новые skills и UX-фичи.

## 14. Итог

На текущем этапе проект уже является функциональным прототипом и понимает архитектуру voice assistant: capture → VAD → STT → route → TTS → playback. Но live-проверка на реальном железе всё ещё остаётся обязательным шагом перед окончанием разработки.

Если после настройки возникают ошибки, их следует фиксировать по реальным stack trace и логам, а не по предположениям.
