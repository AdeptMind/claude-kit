package registry

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// AgentInterfaces holds what an agent produces and consumes.
type AgentInterfaces struct {
	Produces []string
	Consumes []string
}

// AgentEntry represents a single agent in the registry.
type AgentEntry struct {
	Name        string
	Description string
	Model       string
	Skills      []string
	Interfaces  AgentInterfaces
}

// Registry holds all agent entries.
type Registry struct {
	Agents []AgentEntry
}

// ScanAgents reads all .md files in agentsDir and returns parsed agent entries sorted by name.
func ScanAgents(agentsDir string) ([]AgentEntry, error) {
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("reading agents directory: %w", err)
	}

	var agents []AgentEntry
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		path := filepath.Join(agentsDir, entry.Name())
		agent := parseAgentFile(path)
		agents = append(agents, agent)
	}

	sort.Slice(agents, func(i, j int) bool {
		return agents[i].Name < agents[j].Name
	})

	return agents, nil
}

// parseAgentFile extracts all registry fields from an agent markdown file.
func parseAgentFile(path string) AgentEntry {
	f, err := os.Open(path)
	if err != nil {
		name := strings.TrimSuffix(filepath.Base(path), ".md")
		return AgentEntry{Name: name, Skills: []string{}, Interfaces: AgentInterfaces{Produces: []string{}, Consumes: []string{}}}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false

	var agent AgentEntry
	// Track nested parsing state
	var currentList *[]string // points to the list we're appending to (skills, produces, or consumes)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}

		if !inFrontmatter {
			continue
		}

		// Top-level keys (no leading whitespace)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			currentList = nil

			if strings.HasPrefix(line, "name:") {
				agent.Name = extractValue(line, "name:")
			} else if strings.HasPrefix(line, "description:") {
				agent.Description = extractValue(line, "description:")
			} else if strings.HasPrefix(line, "model:") {
				agent.Model = extractValue(line, "model:")
			} else if strings.HasPrefix(line, "skills:") {
				rest := strings.TrimSpace(strings.TrimPrefix(line, "skills:"))
				if rest != "" {
					agent.Skills = parseInlineList(rest)
				} else {
					currentList = &agent.Skills
				}
			} else if strings.HasPrefix(line, "interfaces:") {
				// interfaces: is a parent key; children follow
				continue
			}
			continue
		}

		trimmed := strings.TrimSpace(line)

		// Nested keys under interfaces
		if strings.HasPrefix(trimmed, "produces:") {
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "produces:"))
			if rest != "" {
				agent.Interfaces.Produces = parseInlineList(rest)
				currentList = nil
			} else {
				currentList = &agent.Interfaces.Produces
			}
			continue
		}
		if strings.HasPrefix(trimmed, "consumes:") {
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "consumes:"))
			if rest != "" {
				agent.Interfaces.Consumes = parseInlineList(rest)
				currentList = nil
			} else {
				currentList = &agent.Interfaces.Consumes
			}
			continue
		}

		// List items
		if currentList != nil && strings.HasPrefix(trimmed, "- ") {
			item := strings.TrimPrefix(trimmed, "- ")
			item = strings.TrimSpace(item)
			item = unquote(item)
			if item != "" {
				*currentList = append(*currentList, item)
			}
		}
	}

	// Ensure no nil slices
	if agent.Skills == nil {
		agent.Skills = []string{}
	}
	if agent.Interfaces.Produces == nil {
		agent.Interfaces.Produces = []string{}
	}
	if agent.Interfaces.Consumes == nil {
		agent.Interfaces.Consumes = []string{}
	}

	return agent
}

// extractValue extracts and unquotes the value after a YAML key prefix.
func extractValue(line, prefix string) string {
	val := strings.TrimPrefix(line, prefix)
	val = strings.TrimSpace(val)
	return unquote(val)
}

// unquote removes surrounding quotes from a string.
func unquote(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

// parseInlineList parses a YAML inline list like [a, b, c] or "a, b, c".
func parseInlineList(s string) []string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") {
		s = s[1 : len(s)-1]
	}
	var items []string
	for _, item := range strings.Split(s, ",") {
		item = strings.TrimSpace(item)
		item = unquote(item)
		if item != "" {
			items = append(items, item)
		}
	}
	if items == nil {
		items = []string{}
	}
	return items
}

// ExtractInterfaces parses the interfaces block from an agent file.
func ExtractInterfaces(path string) AgentInterfaces {
	agent := parseAgentFile(path)
	return agent.Interfaces
}

// GenerateYAML produces the YAML content for the agent registry.
func GenerateYAML(agents []AgentEntry) []byte {
	var b strings.Builder
	b.WriteString("# Auto-generated by ck agents registry\n")
	b.WriteString("agents:\n")

	for _, a := range agents {
		b.WriteString(fmt.Sprintf("  - name: %s\n", a.Name))
		b.WriteString(fmt.Sprintf("    description: %q\n", a.Description))
		if a.Model != "" {
			b.WriteString(fmt.Sprintf("    model: %s\n", a.Model))
		}
		writeList(&b, "skills", a.Skills, 4)
		b.WriteString("    interfaces:\n")
		writeList(&b, "produces", a.Interfaces.Produces, 6)
		writeList(&b, "consumes", a.Interfaces.Consumes, 6)
	}

	return []byte(b.String())
}

// writeList writes a YAML list at the given indentation level.
func writeList(b *strings.Builder, key string, items []string, indent int) {
	prefix := strings.Repeat(" ", indent)
	if len(items) == 0 {
		b.WriteString(fmt.Sprintf("%s%s: []\n", prefix, key))
		return
	}
	b.WriteString(fmt.Sprintf("%s%s:\n", prefix, key))
	for _, item := range items {
		b.WriteString(fmt.Sprintf("%s  - %q\n", prefix, item))
	}
}

// WriteRegistry generates the YAML and writes it to the given path.
func WriteRegistry(path string, agents []AgentEntry) error {
	data := GenerateYAML(agents)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}
