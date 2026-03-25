package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/catalog"
)

// SkillArgs is the input schema for every skill tool.
type SkillArgs struct {
	Arguments string `json:"arguments,omitempty"`
}

// MCPServerService runs a local MCP SSE server exposing installed skills as tools.
type MCPServerService struct {
	mu         sync.Mutex
	server     *mcp.Server
	listener   net.Listener
	httpServer *http.Server
	toolNames  []string
}

func (s *MCPServerService) startup(_ context.Context) {
	s.server = mcp.NewServer(&mcp.Implementation{
		Name:    "claude-kit",
		Version: "0.8.0",
	}, nil)

	handler := mcp.NewSSEHandler(func(*http.Request) *mcp.Server {
		return s.server
	}, nil)

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Printf("mcp-server: failed to listen: %v", err)
		return
	}
	s.listener = ln
	s.httpServer = &http.Server{Handler: handler}

	go func() {
		log.Printf("mcp-server: listening on %s", ln.Addr())
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("mcp-server: serve error: %v", err)
		}
	}()
}

// GetEndpoint returns the SSE endpoint URL for Claude Desktop configuration.
func (s *MCPServerService) GetEndpoint() string {
	if s.listener == nil {
		return ""
	}
	return fmt.Sprintf("http://%s/sse", s.listener.Addr())
}

// LoadSkills scans the project's .claude/skills/ and registers each as an MCP tool.
func (s *MCPServerService) LoadSkills(projectPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove previous tools
	if len(s.toolNames) > 0 {
		s.server.RemoveTools(s.toolNames...)
		s.toolNames = nil
	}

	skillsDir := filepath.Join(projectPath, ".claude", "skills")
	skills := catalog.ScanSkillsDir(skillsDir)

	for _, sk := range skills {
		skillPath := filepath.Join(sk.Path, "SKILL.md")
		toolName := strings.ReplaceAll(sk.Name, "/", "_")
		hint := catalog.ExtractFrontmatterField(skillPath, "argument-hint:")

		desc := sk.Description
		if hint != "" {
			desc += " " + hint
		}

		mcp.AddTool(s.server, &mcp.Tool{
			Name:        toolName,
			Description: desc,
		}, s.makeHandler(skillPath))

		s.toolNames = append(s.toolNames, toolName)
	}

	log.Printf("mcp-server: loaded %d skills from %s", len(skills), projectPath)
	return nil
}

func (s *MCPServerService) makeHandler(skillPath string) func(context.Context, *mcp.CallToolRequest, SkillArgs) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, args SkillArgs) (*mcp.CallToolResult, any, error) {
		body, err := catalog.ExtractSkillBody(skillPath)
		if err != nil {
			return nil, nil, fmt.Errorf("reading skill: %w", err)
		}
		if args.Arguments != "" {
			body += "\n\n## User Arguments\n" + args.Arguments
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: body},
			},
		}, nil, nil
	}
}

// Shutdown stops the HTTP server.
func (s *MCPServerService) Shutdown() {
	if s.httpServer != nil {
		s.httpServer.Close()
	}
}
