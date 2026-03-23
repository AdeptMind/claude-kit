package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type MCPConfig struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

type SyncResult struct {
	StoryID string `json:"storyId"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	ExtID   string `json:"extId"`
}

type MCPService struct {
	ctx context.Context
}

func (s *MCPService) startup(ctx context.Context) {
	s.ctx = ctx
}

func (s *MCPService) configPath(projectPath string) string {
	return filepath.Join(projectPath, ".claude", "mcp-sync.json")
}

func (s *MCPService) GetConfig(projectPath string) (*MCPConfig, error) {
	data, err := os.ReadFile(s.configPath(projectPath))
	if err != nil {
		return nil, nil
	}
	var cfg MCPConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *MCPService) SaveConfig(projectPath string, cfg MCPConfig) error {
	dir := filepath.Join(projectPath, ".claude")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.configPath(projectPath), data, 0644)
}

func (s *MCPService) connect(cfg *MCPConfig) (*mcp.ClientSession, error) {
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "claude-kit",
		Version: "0.6.0",
	}, nil)

	cmd := exec.Command(cfg.Command, cfg.Args...)
	cmd.Env = append(os.Environ(), envMapToSlice(cfg.Env)...)

	transport := &mcp.CommandTransport{Command: cmd}

	session, err := client.Connect(context.Background(), transport, nil)
	if err != nil {
		return nil, fmt.Errorf("MCP connect failed: %w", err)
	}
	return session, nil
}

func (s *MCPService) ListTools(projectPath string) ([]string, error) {
	cfg, err := s.GetConfig(projectPath)
	if err != nil || cfg == nil {
		return nil, fmt.Errorf("no MCP sync configured")
	}

	session, err := s.connect(cfg)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	result, err := session.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}
	return names, nil
}

func (s *MCPService) SyncStories(projectPath string, stories []Story) ([]SyncResult, error) {
	cfg, err := s.GetConfig(projectPath)
	if err != nil || cfg == nil {
		return nil, fmt.Errorf("no MCP sync configured")
	}

	session, err := s.connect(cfg)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	results := make([]SyncResult, 0, len(stories))
	for _, story := range stories {
		issueData := map[string]any{
			"summary":     story.ID + ": " + story.Title,
			"description": buildDescription(story),
			"priority":    fmt.Sprintf("P%d", story.Priority),
			"labels":      []string{"claude-kit", story.Status},
		}

		res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
			Name:      "create_issue",
			Arguments: issueData,
		})

		sr := SyncResult{StoryID: story.ID}
		if err != nil {
			sr.Success = false
			sr.Message = err.Error()
		} else {
			sr.Success = true
			sr.Message = "Synced"
			if res != nil && len(res.Content) > 0 {
				if text, ok := res.Content[0].(*mcp.TextContent); ok {
					sr.ExtID = text.Text
					sr.Message = text.Text
				}
			}
		}
		results = append(results, sr)
	}
	return results, nil
}

func buildDescription(s Story) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**Priority:** P%d\n", s.Priority)
	if s.Round > 0 {
		fmt.Fprintf(&b, "**Round:** %d\n", s.Round)
	}
	if s.Component != "" {
		fmt.Fprintf(&b, "**Component:** %s\n", s.Component)
	}
	fmt.Fprintf(&b, "**Status:** %s\n\n", s.Status)
	if len(s.AcceptanceCriteria) > 0 {
		b.WriteString("## Acceptance Criteria\n")
		for _, ac := range s.AcceptanceCriteria {
			fmt.Fprintf(&b, "- %s\n", ac)
		}
	}
	return b.String()
}

func envMapToSlice(m map[string]string) []string {
	s := make([]string, 0, len(m))
	for k, v := range m {
		s = append(s, k+"="+v)
	}
	return s
}
