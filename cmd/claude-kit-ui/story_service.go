package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Story struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Status             string   `json:"status"`
	Priority           int      `json:"priority"`
	Round              int      `json:"round"`
	Component          string   `json:"component"`
	Type               string   `json:"type"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	DependsOn          []string `json:"dependsOn"`
	Locked             bool     `json:"locked"`
}

type StoryService struct{}

// ralph-prd.json structures
type ralphPRD struct {
	UserStories []ralphStory `json:"userStories"`
}

type ralphStory struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Priority           int      `json:"priority"`
	Round              int      `json:"round"`
	Passes             bool     `json:"passes"`
	Component          string   `json:"component"`
	Type               string   `json:"type"`
	AcceptanceCriteria []string `json:"acceptanceCriteria"`
	DependsOn          []string `json:"dependsOn"`
	Locked             bool     `json:"locked,omitempty"`
}

// backlog.md structures
type backlogFile struct {
	Tasks []backlogTask `yaml:"tasks"`
}

type backlogTask struct {
	ID                 string   `yaml:"id"`
	Title              string   `yaml:"title"`
	Component          string   `yaml:"component"`
	Priority           int      `yaml:"priority"`
	Type               string   `yaml:"type"`
	Description        string   `yaml:"description"`
	DependsOn          []string `yaml:"depends_on"`
	AcceptanceCriteria []string `yaml:"acceptance_criteria"`
}

func (s *StoryService) List(projectPath string) ([]Story, error) {
	if projectPath == "" {
		projectPath = s.findProjectPath()
	}
	if projectPath == "" {
		return []Story{}, nil
	}

	stories, err := s.loadFromRalph(projectPath)
	if err == nil && len(stories) > 0 {
		return stories, nil
	}

	stories, err = s.loadFromBacklog(projectPath)
	if err != nil {
		return []Story{}, nil
	}
	return stories, nil
}

// findProjectPath tries multiple locations to find a project with .claude/
func (s *StoryService) findProjectPath() string {
	// 1. Try cwd
	if cwd, err := os.Getwd(); err == nil {
		if _, err := os.Stat(filepath.Join(cwd, ".claude", "ralph-prd.json")); err == nil {
			return cwd
		}
		if _, err := os.Stat(filepath.Join(cwd, ".claude", "output", "backlog.md")); err == nil {
			return cwd
		}
	}

	// 2. Try last project from ~/.claude-kit/state.json
	home, err := os.UserHomeDir()
	if err == nil {
		stateFile := filepath.Join(home, ".claude-kit", "state.json")
		if data, err := os.ReadFile(stateFile); err == nil {
			var state struct {
				LastProject string `json:"lastProject"`
			}
			if json.Unmarshal(data, &state) == nil && state.LastProject != "" {
				return state.LastProject
			}
		}
	}

	// 3. Scan common workspace locations
	if home != "" {
		for _, dir := range []string{
			filepath.Join(home, "workspace"),
			filepath.Join(home, "Library", "Mobile Documents", "com~apple~CloudDocs", "workspace"),
		} {
			entries, err := os.ReadDir(dir)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				candidate := filepath.Join(dir, e.Name())
				if _, err := os.Stat(filepath.Join(candidate, ".claude", "ralph-prd.json")); err == nil {
					return candidate
				}
			}
		}
	}

	return ""
}

func (s *StoryService) GetStats(projectPath string) (map[string]int, error) {
	if projectPath == "" {
		projectPath = s.findProjectPath()
	}
	stories, err := s.List(projectPath)
	if err != nil {
		return nil, err
	}

	stats := map[string]int{
		"done":       0,
		"inProgress": 0,
		"todo":       0,
		"total":      len(stories),
	}
	for _, st := range stories {
		switch st.Status {
		case "done":
			stats["done"]++
		case "in-progress":
			stats["inProgress"]++
		default:
			stats["todo"]++
		}
	}
	return stats, nil
}

func (s *StoryService) loadFromRalph(projectPath string) ([]Story, error) {
	data, err := os.ReadFile(filepath.Join(projectPath, ".claude", "ralph-prd.json"))
	if err != nil {
		return nil, err
	}

	var prd ralphPRD
	if err := json.Unmarshal(data, &prd); err != nil {
		return nil, err
	}

	// Build set of done story IDs for dependency resolution
	doneSet := make(map[string]bool)
	for _, rs := range prd.UserStories {
		if rs.Passes {
			doneSet[rs.ID] = true
		}
	}

	stories := make([]Story, 0, len(prd.UserStories))
	for _, rs := range prd.UserStories {
		status := "todo"
		if rs.Passes {
			status = "done"
		} else if allDepsDone(rs.DependsOn, doneSet) {
			status = "in-progress"
		}

		stories = append(stories, Story{
			ID:                 rs.ID,
			Title:              rs.Title,
			Status:             status,
			Priority:           rs.Priority,
			Round:              rs.Round,
			Component:          rs.Component,
			Type:               rs.Type,
			AcceptanceCriteria: rs.AcceptanceCriteria,
			DependsOn:          rs.DependsOn,
			Locked:             rs.Locked,
		})
	}
	return stories, nil
}

func (s *StoryService) loadFromBacklog(projectPath string) ([]Story, error) {
	data, err := os.ReadFile(filepath.Join(projectPath, ".claude", "output", "backlog.md"))
	if err != nil {
		return nil, err
	}

	var bl backlogFile
	if err := yaml.Unmarshal(data, &bl); err != nil {
		return nil, err
	}

	stories := make([]Story, 0, len(bl.Tasks))
	for _, t := range bl.Tasks {
		stories = append(stories, Story{
			ID:                 t.ID,
			Title:              t.Title,
			Status:             "todo",
			Priority:           t.Priority,
			Component:          t.Component,
			Type:               t.Type,
			AcceptanceCriteria: t.AcceptanceCriteria,
			DependsOn:          t.DependsOn,
		})
	}
	return stories, nil
}

// GetStoryMarkdown returns a story formatted as markdown for editing
func (s *StoryService) GetStoryMarkdown(projectPath string, storyID string) (string, error) {
	if projectPath == "" {
		projectPath = s.findProjectPath()
	}
	stories, err := s.List(projectPath)
	if err != nil {
		return "", err
	}

	var story *Story
	for i := range stories {
		if stories[i].ID == storyID {
			story = &stories[i]
			break
		}
	}
	if story == nil {
		return "", fmt.Errorf("story %s not found", storyID)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "# %s: %s\n\n", story.ID, story.Title)
	fmt.Fprintf(&b, "**Priority:** P%d\n", story.Priority)
	if story.Round > 0 {
		fmt.Fprintf(&b, "**Round:** %d\n", story.Round)
	}
	if story.Component != "" {
		fmt.Fprintf(&b, "**Component:** %s\n", story.Component)
	}
	fmt.Fprintf(&b, "**Status:** %s\n\n", story.Status)

	if len(story.AcceptanceCriteria) > 0 {
		b.WriteString("## Acceptance Criteria\n")
		for _, ac := range story.AcceptanceCriteria {
			fmt.Fprintf(&b, "- %s\n", ac)
		}
		b.WriteString("\n")
	}

	if len(story.DependsOn) > 0 {
		b.WriteString("## Dependencies\n")
		for _, d := range story.DependsOn {
			fmt.Fprintf(&b, "- %s\n", d)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Notes\n_Add your notes here..._\n")
	return b.String(), nil
}

// SaveStoryMarkdown parses edited markdown and updates ralph-prd.json
func (s *StoryService) SaveStoryMarkdown(projectPath string, storyID string, markdown string) error {
	if projectPath == "" {
		projectPath = s.findProjectPath()
	}

	prdPath := filepath.Join(projectPath, ".claude", "ralph-prd.json")
	data, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("cannot read ralph-prd.json: %w", err)
	}

	var prd ralphPRD
	if err := json.Unmarshal(data, &prd); err != nil {
		return fmt.Errorf("cannot parse ralph-prd.json: %w", err)
	}

	// Parse title from "# ID: title"
	var newTitle string
	var newAC []string
	lines := strings.Split(markdown, "\n")
	inAC := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "# ") {
			// Parse "# ID: title"
			after := strings.TrimPrefix(trimmed, "# ")
			if idx := strings.Index(after, ": "); idx >= 0 {
				newTitle = strings.TrimSpace(after[idx+2:])
			}
			inAC = false
			continue
		}

		if trimmed == "## Acceptance Criteria" {
			inAC = true
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			inAC = false
			continue
		}

		if inAC && strings.HasPrefix(trimmed, "- ") {
			newAC = append(newAC, strings.TrimPrefix(trimmed, "- "))
		}
	}

	// Update matching story — don't touch passes
	found := false
	for i := range prd.UserStories {
		if prd.UserStories[i].ID == storyID {
			if newTitle != "" {
				prd.UserStories[i].Title = newTitle
			}
			if len(newAC) > 0 {
				prd.UserStories[i].AcceptanceCriteria = newAC
			}
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("story %s not found in ralph-prd.json", storyID)
	}

	out, err := json.MarshalIndent(prd, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(prdPath, out, 0644)
}

func (s *StoryService) LockStory(projectPath string, storyID string) error {
	return s.setStoryLock(projectPath, storyID, true)
}

func (s *StoryService) UnlockStory(projectPath string, storyID string) error {
	return s.setStoryLock(projectPath, storyID, false)
}

func (s *StoryService) setStoryLock(projectPath string, storyID string, locked bool) error {
	if projectPath == "" {
		projectPath = s.findProjectPath()
	}
	if projectPath == "" {
		return fmt.Errorf("no project path found")
	}

	prdPath := filepath.Join(projectPath, ".claude", "ralph-prd.json")
	data, err := os.ReadFile(prdPath)
	if err != nil {
		return fmt.Errorf("cannot read ralph-prd.json: %w", err)
	}

	var prd ralphPRD
	if err := json.Unmarshal(data, &prd); err != nil {
		return fmt.Errorf("cannot parse ralph-prd.json: %w", err)
	}

	found := false
	for i := range prd.UserStories {
		if prd.UserStories[i].ID == storyID {
			prd.UserStories[i].Locked = locked
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("story %s not found in ralph-prd.json", storyID)
	}

	out, err := json.MarshalIndent(prd, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(prdPath, out, 0644)
}

func allDepsDone(deps []string, doneSet map[string]bool) bool {
	for _, d := range deps {
		if !doneSet[d] {
			return false
		}
	}
	return true
}
