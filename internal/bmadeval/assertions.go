package bmadeval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/eval"
)

// Phase represents a BMAD workflow phase.
type Phase string

const (
	PhaseBreak Phase = "break"
	PhaseModel Phase = "model"
	PhaseAct   Phase = "act"
)

// AvailablePhases returns the list of supported BMAD phases.
func AvailablePhases() []Phase {
	return []Phase{PhaseBreak, PhaseModel, PhaseAct}
}

// LoadAssertions reads and parses the assertion file for the given phase
// from evalsDir/{phase}-assertions.json.
func LoadAssertions(evalsDir string, phase Phase) ([]eval.GradingAssertion, error) {
	if !isValidPhase(phase) {
		return nil, fmt.Errorf("bmadeval: unknown phase %q", phase)
	}

	filename := fmt.Sprintf("%s-assertions.json", phase)
	path := filepath.Join(evalsDir, filename)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("bmadeval: %w", err)
	}

	var assertions []eval.GradingAssertion
	if err := json.Unmarshal(data, &assertions); err != nil {
		return nil, fmt.Errorf("bmadeval: invalid JSON in %s: %w", filename, err)
	}

	for i, a := range assertions {
		if a.Description == "" {
			return nil, fmt.Errorf("bmadeval: %s: entry %d: description is empty", filename, i)
		}
		if a.Type == "" {
			return nil, fmt.Errorf("bmadeval: %s: entry %d: type is empty", filename, i)
		}
	}

	return assertions, nil
}

func isValidPhase(p Phase) bool {
	for _, valid := range AvailablePhases() {
		if p == valid {
			return true
		}
	}
	return false
}
