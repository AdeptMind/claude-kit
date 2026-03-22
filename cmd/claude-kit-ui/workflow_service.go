package main

import "context"

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

func (s *WorkflowService) GetStatus() WorkflowStatus {
	return WorkflowStatus{
		CurrentPhase: "analyze",
		Phases: []PhaseInfo{
			{ID: "break", Label: "Break", Status: "done", Artifacts: []string{"problem.yaml"}},
			{ID: "model", Label: "Model", Status: "done", Artifacts: []string{"architecture.yaml", "backlog.yaml"}},
			{ID: "analyze", Label: "Analyze", Status: "active", Artifacts: []string{"analysis-report.md"}},
			{ID: "act", Label: "Act", Status: "pending", Artifacts: []string{}},
			{ID: "deliver", Label: "Deliver", Status: "pending", Artifacts: []string{}},
		},
	}
}

func (s *WorkflowService) LaunchCowork() error {
	// TODO: launch cowork session
	return nil
}
