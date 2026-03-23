package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type RoleInfo struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type RoleService struct{}

var roleColors = map[string]string{
	"po": "gold", "architect": "gold", "business-analyst": "gold",
	"project-manager": "gold", "scrum-master": "gold", "ux-designer": "gold",
	"dev": "blue", "backend": "blue", "frontend": "blue",
	"devops": "purple", "security": "red", "sre": "purple",
	"golang": "blue", "python": "blue", "typescript": "blue",
}

func (s *RoleService) GetCurrent(projectPath string) RoleInfo {
	if projectPath == "" {
		return RoleInfo{Name: "po", Color: "gold"}
	}
	settingsPath := filepath.Join(projectPath, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return RoleInfo{Name: "po", Color: "gold"}
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return RoleInfo{Name: "po", Color: "gold"}
	}

	env, ok := settings["env"].(map[string]interface{})
	if !ok {
		return RoleInfo{Name: "po", Color: "gold"}
	}
	role, ok := env["CK_USER_ROLE"].(string)
	if !ok || role == "" {
		return RoleInfo{Name: "po", Color: "gold"}
	}

	color := roleColors[role]
	if color == "" {
		color = "gold"
	}
	return RoleInfo{Name: role, Color: color}
}

func (s *RoleService) SetRole(projectPath, role string) error {
	if projectPath == "" {
		return nil
	}
	settingsPath := filepath.Join(projectPath, ".claude", "settings.json")

	var settings map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		settings = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			settings = make(map[string]interface{})
		}
	}

	env, ok := settings["env"].(map[string]interface{})
	if !ok {
		env = make(map[string]interface{})
	}
	env["CK_USER_ROLE"] = role
	settings["env"] = env

	// Ensure .claude directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return err
	}

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, out, 0o644)
}
