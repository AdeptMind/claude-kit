package validate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidationError represents a single validation issue with file and line info.
type ValidationError struct {
	File    string // file path
	Line    int    // line number (0 if unknown)
	Field   string // field name
	Message string // human-readable error
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("%s:%d: %s: %s", e.File, e.Line, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s: %s", e.File, e.Field, e.Message)
}

// knownTools is the set of valid tool names for agent frontmatter.
var knownTools = map[string]bool{
	"Read":         true,
	"Write":        true,
	"Edit":         true,
	"Bash":         true,
	"Grep":         true,
	"Glob":         true,
	"Task":         true,
	"WebFetch":     true,
	"WebSearch":    true,
	"NotebookEdit": true,
}

// ValidateSkill validates a skill directory and returns any errors found.
// skillDir should point to the skill directory (e.g. skills/code-reviewer/).
func ValidateSkill(skillDir string) []ValidationError {
	var errs []ValidationError

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		errs = append(errs, ValidationError{
			File:    skillDir,
			Field:   "SKILL.md",
			Message: "required file SKILL.md not found",
		})
		return errs
	}

	fields, lineMap, err := parseFrontmatter(skillFile)
	if err != nil {
		errs = append(errs, ValidationError{
			File:    skillFile,
			Field:   "frontmatter",
			Message: fmt.Sprintf("failed to parse frontmatter: %v", err),
		})
		return errs
	}

	for _, required := range []string{"name", "description"} {
		if val, ok := fields[required]; !ok || strings.TrimSpace(val) == "" {
			line := lineMap[required]
			if line == 0 {
				// Field is missing entirely; report at line 1 (file level)
				line = 0
			}
			errs = append(errs, ValidationError{
				File:    skillFile,
				Line:    line,
				Field:   required,
				Message: "required field is missing or empty",
			})
		}
	}

	// Validate evals.json if present
	evalsFile := filepath.Join(skillDir, "evals.json")
	if _, err := os.Stat(evalsFile); err == nil {
		data, readErr := os.ReadFile(evalsFile)
		if readErr != nil {
			errs = append(errs, ValidationError{
				File:    evalsFile,
				Field:   "evals.json",
				Message: fmt.Sprintf("failed to read file: %v", readErr),
			})
		} else if !json.Valid(data) {
			errs = append(errs, ValidationError{
				File:    evalsFile,
				Field:   "evals.json",
				Message: "file contains invalid JSON",
			})
		}
	}

	return errs
}

// ValidateAgent validates an agent markdown file and returns any errors found.
// agentPath is the path to the agent .md file.
// skillsDir is the directory containing skill subdirectories (used to verify skill refs).
func ValidateAgent(agentPath string, skillsDir string) []ValidationError {
	var errs []ValidationError

	if _, err := os.Stat(agentPath); err != nil {
		errs = append(errs, ValidationError{
			File:    agentPath,
			Field:   "file",
			Message: "agent file not found",
		})
		return errs
	}

	if !strings.HasSuffix(agentPath, ".md") {
		errs = append(errs, ValidationError{
			File:    agentPath,
			Field:   "file",
			Message: "agent file must have .md extension",
		})
		return errs
	}

	fields, lineMap, err := parseFrontmatter(agentPath)
	if err != nil {
		errs = append(errs, ValidationError{
			File:    agentPath,
			Field:   "frontmatter",
			Message: fmt.Sprintf("failed to parse frontmatter: %v", err),
		})
		return errs
	}

	for _, required := range []string{"name", "description"} {
		if val, ok := fields[required]; !ok || strings.TrimSpace(val) == "" {
			errs = append(errs, ValidationError{
				File:    agentPath,
				Line:    lineMap[required],
				Field:   required,
				Message: "required field is missing or empty",
			})
		}
	}

	// Validate tools if listed
	if toolsRaw, ok := fields["tools"]; ok && toolsRaw != "" {
		tools := parseInlineList(toolsRaw)
		for _, tool := range tools {
			if !knownTools[tool] {
				errs = append(errs, ValidationError{
					File:    agentPath,
					Line:    lineMap["tools"],
					Field:   "tools",
					Message: fmt.Sprintf("unknown tool %q", tool),
				})
			}
		}
	}

	// Validate skill references if listed
	if _, ok := fields["skills"]; ok {
		skills := parseSkillsList(agentPath)
		for _, skill := range skills {
			skillPath := filepath.Join(skillsDir, skill, "SKILL.md")
			if _, err := os.Stat(skillPath); err != nil {
				errs = append(errs, ValidationError{
					File:    agentPath,
					Line:    lineMap["skills"],
					Field:   "skills",
					Message: fmt.Sprintf("skill %q not found (expected %s)", skill, skillPath),
				})
			}
		}
	}

	return errs
}

// parseFrontmatter reads YAML frontmatter from a markdown file.
// Returns a map of field name -> raw string value, and field name -> line number.
func parseFrontmatter(path string) (map[string]string, map[string]int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	fields := make(map[string]string)
	lineMap := make(map[string]int)
	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if strings.TrimSpace(line) == "---" {
			if inFrontmatter {
				break // end of frontmatter
			}
			inFrontmatter = true
			continue
		}

		if !inFrontmatter {
			continue
		}

		// Skip lines that are continuation (start with whitespace or -)
		if len(line) == 0 || line[0] == ' ' || line[0] == '\t' || line[0] == '-' {
			continue
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}

		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])

		// Strip surrounding quotes
		if len(val) >= 2 && ((val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'')) {
			val = val[1 : len(val)-1]
		}

		fields[key] = val
		lineMap[key] = lineNum
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return fields, lineMap, nil
}

// parseInlineList parses a YAML inline list like "[Read, Write, Edit]" or "Read, Write, Edit".
func parseInlineList(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")

	var items []string
	for _, s := range strings.Split(raw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			items = append(items, s)
		}
	}
	return items
}

// parseSkillsList extracts the skills list from an agent file's frontmatter.
// Handles both inline (skills: [a, b]) and block list (skills:\n  - a\n  - b) formats.
func parseSkillsList(path string) []string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFrontmatter := false
	inSkills := false
	var skills []string

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

		if strings.HasPrefix(line, "skills:") {
			rest := strings.TrimPrefix(line, "skills:")
			rest = strings.TrimSpace(rest)
			if rest != "" && rest != "|" {
				// Inline list: skills: [a, b] or skills: a, b
				for _, s := range parseInlineList(rest) {
					skills = append(skills, s)
				}
				inSkills = false
			} else {
				inSkills = true
			}
			continue
		}

		if inSkills {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "- ") {
				skill := strings.TrimPrefix(trimmed, "- ")
				skill = strings.TrimSpace(skill)
				if skill != "" {
					skills = append(skills, skill)
				}
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				inSkills = false
			}
		}
	}

	return skills
}
