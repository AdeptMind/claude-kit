package main

import (
	"context"
	"os"
	"path/filepath"
)

type PhaseInfo struct {
	ID        string   `json:"id"`
	Label     string   `json:"label"`
	Status    string   `json:"status"`
	Artifacts []string `json:"artifacts"`
}

type WorkflowStatus struct {
	CurrentPhase string      `json:"currentPhase"`
	Phases       []PhaseInfo `json:"phases"`
}

type WorkflowService struct {
	ctx context.Context
}

func (s *WorkflowService) startup(ctx context.Context) {
	s.ctx = ctx
}

var phaseDefinitions = []struct {
	id        string
	label     string
	artifacts []string
}{
	{"break", "Break", []string{"problem.md"}},
	{"model", "Model", []string{"architecture.md", "backlog.md"}},
	{"analyze", "Analyze", []string{"analysis-report.md"}},
	{"act", "Act", []string{"gsd-prep.md"}},
	{"deliver", "Deliver", []string{"release-notes.md"}},
}

func (s *WorkflowService) GetStatus(projectPath string) WorkflowStatus {
	outputDir := filepath.Join(projectPath, ".claude", "output")

	var phases []PhaseInfo
	lastDone := -1

	for i, def := range phaseDefinitions {
		var found []string
		for _, a := range def.artifacts {
			full := filepath.Join(outputDir, a)
			if _, err := os.Stat(full); err == nil {
				found = append(found, a)
			}
		}
		status := "pending"
		if len(found) == len(def.artifacts) {
			status = "done"
			lastDone = i
		}
		phases = append(phases, PhaseInfo{
			ID:        def.id,
			Label:     def.label,
			Status:    status,
			Artifacts: found,
		})
	}

	// The phase right after the last done one is "active"
	currentPhase := "break"
	if lastDone >= 0 && lastDone+1 < len(phases) {
		phases[lastDone+1].Status = "active"
		currentPhase = phases[lastDone+1].ID
	} else if lastDone == len(phases)-1 {
		currentPhase = "deliver"
	}

	return WorkflowStatus{
		CurrentPhase: currentPhase,
		Phases:       phases,
	}
}

func (s *WorkflowService) LaunchCowork() error {
	// TODO: launch cowork session
	return nil
}
