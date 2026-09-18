package sidecar

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	assistantpb "github.com/you/voice-assistant/internal/sidecar/proto/proto"
)

type Manager struct {
	addr string
	cmd  *exec.Cmd
	mu   sync.Mutex
}

func NewManager(addr string) *Manager {
	if addr == "" {
		addr = "127.0.0.1:50051"
	}
	return &Manager{addr: addr}
}

func (m *Manager) Addr() string {
	return m.addr
}

func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd != nil && m.cmd.Process != nil {
		return nil
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working dir: %w", err)
	}

	serverPath := filepath.Join(root, "sidecar", "server.py")
	if _, err := os.Stat(serverPath); err != nil {
		return fmt.Errorf("sidecar server not found at %s: %w", serverPath, err)
	}

	cmd := exec.CommandContext(ctx, "python3", serverPath)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start python sidecar: %w", err)
	}

	m.cmd = cmd
	if err := m.waitForHealth(15 * time.Second); err != nil {
		_ = m.Stop()
		return err
	}

	return nil
}

func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cmd == nil || m.cmd.Process == nil {
		return nil
	}

	if err := m.cmd.Process.Signal(os.Interrupt); err != nil && !isProcessDoneErr(err) {
		return fmt.Errorf("signal sidecar: %w", err)
	}

	waitDone := make(chan error, 1)
	go func() {
		waitDone <- m.cmd.Wait()
	}()

	select {
	case err := <-waitDone:
		if err != nil && !isProcessDoneErr(err) {
			return fmt.Errorf("wait sidecar exit: %w", err)
		}
	case <-time.After(5 * time.Second):
		if killErr := m.cmd.Process.Kill(); killErr != nil && !isProcessDoneErr(killErr) {
			return fmt.Errorf("kill sidecar: %w", killErr)
		}
		<-waitDone
	}

	m.cmd = nil
	return nil
}

func (m *Manager) waitForHealth(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		conn, err := grpc.DialContext(context.Background(), m.addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
			grpc.WithTimeout(500*time.Millisecond),
		)
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			resp, healthErr := assistantpb.NewMLServiceClient(conn).HealthCheck(ctx, &assistantpb.HealthRequest{})
			cancel()
			_ = conn.Close()
			if healthErr == nil && resp != nil && resp.Ok {
				return nil
			}
		}

		if m.cmd != nil && m.cmd.Process != nil {
			if exitErr := m.cmd.ProcessState; exitErr != nil {
				return fmt.Errorf("sidecar exited before health check: %v", exitErr)
			}
		}

		time.Sleep(250 * time.Millisecond)
	}

	return fmt.Errorf("sidecar health check timed out for %s", m.addr)
}

func isProcessDoneErr(err error) bool {
	return err == nil || err.Error() == "os: process already finished" || err.Error() == "signal: interrupt"
}
