package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const maxBufferSize = 512 * 1024 // 512KB of scrollback per session

type termSession struct {
	ptmx    *os.File
	cmd     *exec.Cmd
	buffer  []byte
	running bool
}

type TerminalService struct {
	ctx           context.Context
	sessions      map[string]*termSession
	activeProject string
	mu            sync.Mutex
}

func (t *TerminalService) startup(ctx context.Context) {
	t.ctx = ctx
	t.sessions = make(map[string]*termSession)
}

// Start launches or resumes a session for the given project.
func (t *TerminalService) Start(projectPath string, cols int, rows int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.activeProject = projectPath

	// If session already exists and running, replay buffer and reattach
	if sess, ok := t.sessions[projectPath]; ok && sess.running {
		if len(sess.buffer) > 0 {
			runtime.EventsEmit(t.ctx, "terminal:output", string(sess.buffer))
		}
		return nil
	}

	// Create new session
	cmd := exec.Command("claude")
	cmd.Dir = projectPath
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	// Set initial PTY size to match the xterm.js terminal
	winSize := &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)}
	if cols <= 0 || rows <= 0 {
		winSize = &pty.Winsize{Rows: 40, Cols: 120}
	}
	ptmx, err := pty.StartWithSize(cmd, winSize)
	if err != nil {
		return err
	}

	sess := &termSession{
		ptmx:    ptmx,
		cmd:     cmd,
		buffer:  make([]byte, 0, maxBufferSize),
		running: true,
	}
	t.sessions[projectPath] = sess

	// Read PTY output — buffer it AND emit if this is the active session
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				chunk := make([]byte, n)
				copy(chunk, buf[:n])

				t.mu.Lock()
				sess.buffer = append(sess.buffer, chunk...)
				if len(sess.buffer) > maxBufferSize {
					sess.buffer = sess.buffer[len(sess.buffer)-maxBufferSize:]
				}
				isActive := t.activeProject == projectPath
				t.mu.Unlock()

				if isActive {
					runtime.EventsEmit(t.ctx, "terminal:output", string(chunk))
				}
			}
			if err != nil {
				break
			}
		}

		t.mu.Lock()
		sess.running = false
		sess.ptmx = nil
		isActive := t.activeProject == projectPath
		t.mu.Unlock()

		if isActive {
			runtime.EventsEmit(t.ctx, "terminal:exit", "")
		}
	}()

	// Wait for process
	go func() {
		cmd.Wait()
		t.mu.Lock()
		sess.running = false
		t.mu.Unlock()
	}()

	return nil
}

// SwitchTo switches the active session to a different project.
// Returns true if an existing session was found (replay will happen).
func (t *TerminalService) SwitchTo(projectPath string) (bool, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.activeProject = projectPath

	sess, exists := t.sessions[projectPath]
	if exists && sess.running {
		if len(sess.buffer) > 0 {
			runtime.EventsEmit(t.ctx, "terminal:output", string(sess.buffer))
		}
		return true, nil
	}

	return false, nil
}

// Write sends input to the active session's PTY.
func (t *TerminalService) Write(data string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	sess, ok := t.sessions[t.activeProject]
	if !ok || !sess.running || sess.ptmx == nil {
		return fmt.Errorf("no active terminal session")
	}
	_, err := sess.ptmx.Write([]byte(data))
	return err
}

// Resize updates the active session's PTY window size.
func (t *TerminalService) Resize(cols int, rows int) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	sess, ok := t.sessions[t.activeProject]
	if !ok || sess.ptmx == nil {
		return nil
	}
	return pty.Setsize(sess.ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

// Stop kills the active session.
func (t *TerminalService) Stop() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	sess, ok := t.sessions[t.activeProject]
	if !ok {
		return nil
	}

	if sess.cmd != nil && sess.cmd.Process != nil {
		sess.cmd.Process.Signal(os.Interrupt)
	}
	if sess.ptmx != nil {
		sess.ptmx.Close()
		sess.ptmx = nil
	}
	sess.running = false
	delete(t.sessions, t.activeProject)
	return nil
}

// StopAll kills all sessions (called on app quit).
func (t *TerminalService) StopAll() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for path, sess := range t.sessions {
		if sess.cmd != nil && sess.cmd.Process != nil {
			sess.cmd.Process.Signal(os.Interrupt)
		}
		if sess.ptmx != nil {
			sess.ptmx.Close()
		}
		delete(t.sessions, path)
	}
}

// IsRunning returns true if the active session is running.
func (t *TerminalService) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	sess, ok := t.sessions[t.activeProject]
	return ok && sess.running
}

// ListSessions returns project paths with active sessions.
func (t *TerminalService) ListSessions() []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	var paths []string
	for path, sess := range t.sessions {
		if sess.running {
			paths = append(paths, path)
		}
	}
	return paths
}
