package main

import (
	"encoding/json"
	"os"
	"path/filepath"

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
		cwd, err := os.Getwd()
		if err != nil {
			return []Story{}, nil
		}
		projectPath = cwd
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

func (s *StoryService) GetStats(projectPath string) (map[string]int, error) {
	if projectPath == "" {
		cwd, _ := os.Getwd()
		projectPath = cwd
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

func allDepsDone(deps []string, doneSet map[string]bool) bool {
	for _, d := range deps {
		if !doneSet[d] {
			return false
		}
	}
	return true
}
