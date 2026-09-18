package sidecar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	assistantpb "github.com/you/voice-assistant/internal/sidecar/proto/proto"
)

type TTSClient struct {
	addr string
	conn *grpc.ClientConn
}

func NewTTSClient(addr string) (*TTSClient, error) {
	if addr == "" {
		addr = "127.0.0.1:50051"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial sidecar for TTS: %w", err)
	}
	return &TTSClient{addr: addr, conn: conn}, nil
}

func (c *TTSClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *TTSClient) Synthesize(ctx context.Context, text, voice string) ([]byte, int, error) {
	if c == nil || c.conn == nil {
		return nil, 0, fmt.Errorf("tts client is not connected")
	}
	resp, err := assistantpb.NewMLServiceClient(c.conn).Synthesize(ctx, &assistantpb.SynthesizeRequest{
		Text: text,
		Voice: voice,
	})
	if err != nil {
		return nil, 0, err
	}
	return resp.GetAudio(), int(resp.GetSampleRate()), nil
}
