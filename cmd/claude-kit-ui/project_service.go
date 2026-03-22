package main

import "context"

type ProjectState struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ProjectService struct {
	ctx context.Context
}

func (s *ProjectService) startup(ctx context.Context) {
	s.ctx = ctx
}

func (s *ProjectService) GetState() ProjectState {
	return ProjectState{
		Name: "claude-kit-ui",
		Path: "~/workspace/claude-kit",
	}
}
