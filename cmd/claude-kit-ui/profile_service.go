package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

)

var managementAgents = map[string]bool{
	"product-owner": true, "business-analyst": true, "project-manager": true,
	"scrum-master": true, "people-ops": true, "change-manager": true,
	"architect": true, "ux-designer": true, "ui-designer": true,
	"finops": true, "data-governance": true, "soc2-compliance": true,
	"access-review": true,
}

var specialAgents = map[string]bool{
	"eval-grader": true, "eval-comparator": true, "eval-analyzer": true,
}

type AgentInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

type ProfileForm struct {
	AgentName          string `json:"agentName"`
	RoleTitle          string `json:"roleTitle"`
	Description        string `json:"description"`
	Responsibilities   string `json:"responsibilities"`
	FocusAreas         string `json:"focusAreas"`
	CommunicationStyle string `json:"communicationStyle"`
	LanguagePreference string `json:"languagePreference"`
	CustomRules        string `json:"customRules"`
}

type ProfileService struct{}

func (s *ProfileService) ListAgents(projectPath string) ([]AgentInfo, error) {
	// Read from project's installed agents, not template dir
	agentsDir := filepath.Join(projectPath, ".claude", "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		return nil, fmt.Errorf("cannot read agents dir: %w", err)
	}

	var agents []AgentInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		name, desc := parseFrontmatterLight(filepath.Join(agentsDir, e.Name()))
		if name == "" {
			name = strings.TrimSuffix(e.Name(), ".md")
		}
		agents = append(agents, AgentInfo{
			Name:        name,
			Description: desc,
			Category:    categorize(name),
		})
	}
	return agents, nil
}

func (s *ProfileService) LoadAgent(projectPath string, agentName string) (ProfileForm, error) {
	agentPath := filepath.Join(projectPath, ".claude", "agents", agentName+".md")
	data, err := os.ReadFile(agentPath)
	if err != nil {
		return ProfileForm{}, fmt.Errorf("cannot read agent %s: %w", agentName, err)
	}

	fm := parseFrontmatter(string(data))
	rules := parseRulesSection(string(data))

	cat := categorize(agentName)
	commStyle := "Direct"
	lang := "Technical"
	if cat == "management" {
		commStyle = "Collaborative"
		lang = "Product"
	}

	skills := fm["skills"]
	resp := fm["description"]
	if skills != "" {
		resp += "\n\nSkills: " + skills
	}

	return ProfileForm{
		AgentName:          agentName,
		RoleTitle:          toTitle(agentName),
		Description:        fm["description"],
		Responsibilities:   resp,
		FocusAreas:         skills,
		CommunicationStyle: commStyle,
		LanguagePreference: lang,
		CustomRules:        rules,
	}, nil
}

func (s *ProfileService) SaveProfile(form ProfileForm, targetDir string) error {
	profileDir := filepath.Join(targetDir, ".claude", "cowork-profiles", form.AgentName)
	if err := os.MkdirAll(profileDir, 0o755); err != nil {
		return fmt.Errorf("cannot create profile dir: %w", err)
	}

	aboutMe := fmt.Sprintf(`# About Me — %s

## Role
%s

## Responsibilities
%s

## Focus Areas
%s

## Communication Style
%s
`, form.RoleTitle, form.Description, form.Responsibilities, form.FocusAreas, form.CommunicationStyle)

	workingRules := fmt.Sprintf(`# Working Rules — %s

%s
`, form.RoleTitle, form.CustomRules)

	voiceStyle := fmt.Sprintf(`# Voice and Style — %s

## Language
%s

## Communication
%s
`, form.RoleTitle, form.LanguagePreference, form.CommunicationStyle)

	files := map[string]string{
		"about-me.md":       aboutMe,
		"working-rules.md":  workingRules,
		"voice-and-style.md": voiceStyle,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(profileDir, name), []byte(content), 0o644); err != nil {
			return fmt.Errorf("cannot write %s: %w", name, err)
		}
	}
	return nil
}

func (s *ProfileService) GetSavedProfile(projectPath string) (ProfileForm, error) {
	profilesDir := filepath.Join(projectPath, ".claude", "cowork-profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return ProfileForm{}, fmt.Errorf("no saved profiles: %w", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		aboutPath := filepath.Join(profilesDir, e.Name(), "about-me.md")
		data, err := os.ReadFile(aboutPath)
		if err != nil {
			continue
		}
		return parseAboutMe(e.Name(), string(data)), nil
	}
	return ProfileForm{}, fmt.Errorf("no saved profiles found")
}

// --- helpers ---

func categorize(name string) string {
	if specialAgents[name] {
		return "special"
	}
	if managementAgents[name] {
		return "management"
	}
	return "dev"
}

func toTitle(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

func parseFrontmatterLight(path string) (name, desc string) {
	f, err := os.Open(path)
	if err != nil {
		return "", ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	inFM := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			if inFM {
				break
			}
			inFM = true
			continue
		}
		if !inFM {
			continue
		}
		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
		if strings.HasPrefix(line, "description:") {
			desc = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	return name, desc
}

func parseFrontmatter(content string) map[string]string {
	result := map[string]string{}
	lines := strings.Split(content, "\n")
	inFM := false
	var skillLines []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if inFM {
				break
			}
			inFM = true
			continue
		}
		if !inFM {
			continue
		}
		if strings.HasPrefix(line, "name:") {
			result["name"] = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
		if strings.HasPrefix(line, "description:") {
			result["description"] = strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
		if strings.HasPrefix(line, "skills:") {
			val := strings.TrimSpace(strings.TrimPrefix(line, "skills:"))
			if val != "" {
				// inline list: skills: [a, b, c]  or  skills: a
				result["skills"] = val
			}
			continue
		}
		if strings.HasPrefix(line, "  - ") && len(result["skills"]) == 0 {
			skillLines = append(skillLines, strings.TrimSpace(strings.TrimPrefix(line, "  - ")))
		}
	}
	if len(skillLines) > 0 && result["skills"] == "" {
		result["skills"] = strings.Join(skillLines, ", ")
	}
	return result
}

func parseRulesSection(content string) string {
	idx := strings.Index(content, "## Rules")
	if idx == -1 {
		return ""
	}
	rest := content[idx+len("## Rules"):]
	// find next ## heading
	nextH2 := strings.Index(rest[1:], "\n## ")
	if nextH2 != -1 {
		rest = rest[:nextH2+1]
	}
	return strings.TrimSpace(rest)
}

func parseAboutMe(agentName, content string) ProfileForm {
	form := ProfileForm{AgentName: agentName, RoleTitle: toTitle(agentName)}
	sections := strings.Split(content, "## ")
	for _, sec := range sections {
		lines := strings.SplitN(sec, "\n", 2)
		if len(lines) < 2 {
			continue
		}
		heading := strings.TrimSpace(lines[0])
		body := strings.TrimSpace(lines[1])
		switch heading {
		case "Role":
			form.Description = body
		case "Responsibilities":
			form.Responsibilities = body
		case "Focus Areas":
			form.FocusAreas = body
		case "Communication Style":
			form.CommunicationStyle = body
		}
	}
	return form
}
