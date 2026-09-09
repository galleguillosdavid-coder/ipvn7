package watchdog

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// NodeSupervisor manages the lifecycle of a canary IPv7 node process
type NodeSupervisor struct {
	BinaryPath    string
	Args          []string
	NodeID        string
	RestartCount  atomic.Int32
	MaxRestarts   int
	BackoffDelay  time.Duration
	cancel        context.CancelFunc
	mu            sync.Mutex
	lastExitError error
}

// NewNodeSupervisor initializes a supervisor for an IPv7 node binary
func NewNodeSupervisor(nodeID, binaryPath string, args []string) *NodeSupervisor {
	return &NodeSupervisor{
		BinaryPath:   binaryPath,
		Args:         args,
		NodeID:       nodeID,
		MaxRestarts:  10,
		BackoffDelay: 1 * time.Second,
	}
}

// Start launches the watchdog supervision loop
func (s *NodeSupervisor) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			fmt.Printf("[WATCHDOG][%s] Iniciando proceso nodo (%s)...\n", s.NodeID, s.BinaryPath)
			cmd := exec.CommandContext(ctx, s.BinaryPath, s.Args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			err := cmd.Run()
			s.mu.Lock()
			s.lastExitError = err
			s.mu.Unlock()

			select {
			case <-ctx.Done():
				fmt.Printf("[WATCHDOG][%s] Detenido limpiamente por contexto.\n", s.NodeID)
				return
			default:
			}

			restarts := s.RestartCount.Add(1)
			fmt.Printf("[WATCHDOG][%s] Proceso finalizó (error: %v). Reinicio #%d en %v...\n",
				s.NodeID, err, restarts, s.BackoffDelay)

			if int(restarts) >= s.MaxRestarts {
				fmt.Printf("[WATCHDOG][%s] Límite de reinicios alcanzado (%d). Pausando watchdog.\n",
					s.NodeID, s.MaxRestarts)
				return
			}

			time.Sleep(s.BackoffDelay)
		}
	}()
}

// Stop terminates the supervisor and child process
func (s *NodeSupervisor) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}
