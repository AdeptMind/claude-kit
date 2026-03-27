package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/catalog"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/config"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/embedder"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/knowledge"
)

// SkillArgs is the input schema for every skill tool.
type SkillArgs struct {
	Arguments string `json:"arguments,omitempty"`
}

// MCPServerService runs a local MCP SSE server exposing installed skills as tools.
type MCPServerService struct {
	mu             sync.Mutex
	server         *mcp.Server
	listener       net.Listener
	httpServer     *http.Server
	toolNames      []string
	knowledgeStore *knowledge.Store
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

// RegisterInClaudeDesktop adds the claude-kit MCP server entry to
// ~/Library/Application Support/Claude/claude_desktop_config.json.
func (s *MCPServerService) RegisterInClaudeDesktop() error {
	endpoint := s.GetEndpoint()
	if endpoint == "" {
		return fmt.Errorf("MCP server not running")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, "Library", "Application Support", "Claude")
	configPath := filepath.Join(configDir, "claude_desktop_config.json")

	// Read existing config or start fresh
	cfg := map[string]any{}
	if data, err := os.ReadFile(configPath); err == nil {
		_ = json.Unmarshal(data, &cfg)
	}

	// Ensure mcpServers map exists
	servers, ok := cfg["mcpServers"].(map[string]any)
	if !ok {
		servers = map[string]any{}
	}
	servers["claude-kit"] = map[string]any{"url": endpoint}
	cfg["mcpServers"] = servers

	// Write back
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, append(out, '\n'), 0o644)
}

// BootstrapClaudeDesktop generates cowork profiles for all installed agents
// and registers the MCP server in Claude Desktop config — one-click setup.
func (s *MCPServerService) BootstrapClaudeDesktop(projectPath string) (int, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, err
	}

	profileSvc := &ProfileService{}
	agents, err := profileSvc.ListAgents(projectPath)
	if err != nil {
		return 0, fmt.Errorf("listing agents: %w", err)
	}

	coworkDir := filepath.Join(home, ".claude", "cowork-profiles")
	count := 0
	for _, agent := range agents {
		form, err := profileSvc.LoadAgent(projectPath, agent.Name)
		if err != nil {
			continue
		}
		// Write profile to global cowork dir (not project-local)
		profileDir := filepath.Join(coworkDir, form.AgentName)
		if err := os.MkdirAll(profileDir, 0o755); err != nil {
			continue
		}
		writeProfileFiles(profileDir, form)
		count++
	}

	// Register MCP server
	_ = s.RegisterInClaudeDesktop()

	return count, nil
}

func writeProfileFiles(dir string, form ProfileForm) {
	files := map[string]string{
		"about-me.md": fmt.Sprintf("# About Me — %s\n\n## Role\n%s\n\n## Responsibilities\n%s\n\n## Focus Areas\n%s\n\n## Communication Style\n%s\n",
			form.RoleTitle, form.Description, form.Responsibilities, form.FocusAreas, form.CommunicationStyle),
		"working-rules.md": fmt.Sprintf("# Working Rules — %s\n\n%s\n", form.RoleTitle, form.CustomRules),
		"voice-and-style.md": fmt.Sprintf("# Voice and Style — %s\n\n## Language\n%s\n\n## Communication\n%s\n",
			form.RoleTitle, form.LanguagePreference, form.CommunicationStyle),
	}
	for name, content := range files {
		_ = os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644)
	}
}

// LoadKnowledge initializes the knowledge store and registers knowledge tools.
func (s *MCPServerService) LoadKnowledge() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.knowledgeStore != nil {
		s.knowledgeStore.Close()
	}

	dbPath := filepath.Join(config.DataDir(), "knowledge.sqlite")
	store, err := knowledge.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening knowledge store: %w", err)
	}
	s.knowledgeStore = store

	ctx := context.Background()
	emb := embedder.AutoDetect(ctx)
	if emb != nil {
		store.SetEmbedder(emb)
	}

	RegisterKnowledgeTools(s.server, store, emb)
	log.Printf("mcp-server: knowledge tools loaded (db=%s)", dbPath)
	return nil
}

// Shutdown stops the HTTP server.
func (s *MCPServerService) Shutdown() {
	if s.knowledgeStore != nil {
		s.knowledgeStore.Close()
	}
	if s.httpServer != nil {
		s.httpServer.Close()
	}
}
