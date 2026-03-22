package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TerminalService struct {
	ctx  context.Context
	ptmx *os.File
	cmd  *exec.Cmd
	mu   sync.Mutex
}

func (t *TerminalService) startup(ctx context.Context) {
	t.ctx = ctx
}

// Start launches claude CLI in a PTY and streams output via events.
func (t *TerminalService) Start(projectPath string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.ptmx != nil {
		return nil // already running
	}

	cmd := exec.Command("claude")
	cmd.Dir = projectPath
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	t.cmd = cmd
	t.ptmx = ptmx

	// Read PTY output and emit to frontend.
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				runtime.EventsEmit(t.ctx, "terminal:output", string(buf[:n]))
			}
			if err != nil {
				if err != io.EOF {
					runtime.EventsEmit(t.ctx, "terminal:error", err.Error())
				}
				runtime.EventsEmit(t.ctx, "terminal:exit", "")
				break
			}
		}
	}()

	// Wait for process to exit.
	go func() {
		t.cmd.Wait()
		t.mu.Lock()
		t.ptmx = nil
		t.cmd = nil
		t.mu.Unlock()
		runtime.EventsEmit(t.ctx, "terminal:exit", "")
	}()

	return nil
}

// Write sends input to the PTY.
func (t *TerminalService) Write(data string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ptmx == nil {
		return fmt.Errorf("terminal not running")
	}
	_, err := t.ptmx.Write([]byte(data))
	return err
}

// Resize updates the PTY window size.
func (t *TerminalService) Resize(cols int, rows int) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.ptmx == nil {
		return nil
	}
	return pty.Setsize(t.ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

// Stop kills the terminal process.
func (t *TerminalService) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.cmd != nil && t.cmd.Process != nil {
		t.cmd.Process.Signal(os.Interrupt)
	}
	if t.ptmx != nil {
		t.ptmx.Close()
		t.ptmx = nil
	}
	return nil
}

// IsRunning returns true if claude is running.
func (t *TerminalService) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ptmx != nil
}
