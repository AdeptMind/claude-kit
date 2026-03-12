package claude

import (
	"context"
	"os/exec"
)

// CommandBuilderFunc is the signature for functions that create exec.Cmd.
type CommandBuilderFunc func(ctx context.Context, name string, args ...string) *exec.Cmd

// SetCommandBuilder overrides the command builder used by RunQuery/RunPrompt.
// Intended for use in tests from other packages.
func SetCommandBuilder(fn CommandBuilderFunc) {
	commandBuilder = fn
}

// ExportCommandBuilder returns the current command builder so it can be restored.
func ExportCommandBuilder() CommandBuilderFunc {
	return commandBuilder
}
