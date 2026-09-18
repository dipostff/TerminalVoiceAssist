package sidecar

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	assistantpb "github.com/you/voice-assistant/internal/sidecar/proto/proto"
)

type STTClient struct {
	addr string
	conn *grpc.ClientConn
}

func NewSTTClient(addr string) (*STTClient, error) {
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
		return nil, fmt.Errorf("dial sidecar: %w", err)
	}

	return &STTClient{addr: addr, conn: conn}, nil
}

func (c *STTClient) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *STTClient) Transcribe(ctx context.Context, audio []byte, sampleRate int) (string, error) {
	if c == nil || c.conn == nil {
		return "", fmt.Errorf("stt client is not connected")
	}

	resp, err := assistantpb.NewMLServiceClient(c.conn).Transcribe(ctx, &assistantpb.TranscribeRequest{
		Audio:      audio,
		SampleRate: int32(sampleRate),
		Language:   "ru",
	})
	if err != nil {
		return "", err
	}
	return resp.GetText(), nil
}
