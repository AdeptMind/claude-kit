package catalog

import (
	"os"
	"strings"
)

// ExtractSkillBody reads a SKILL.md file and returns everything after the
// YAML frontmatter (the content between the second "---" and the end of file).
func ExtractSkillBody(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	content := string(data)
	// Find the second "---" delimiter
	first := strings.Index(content, "---")
	if first < 0 {
		return content, nil
	}
	second := strings.Index(content[first+3:], "---")
	if second < 0 {
		return content, nil
	}
	body := content[first+3+second+3:]
	return strings.TrimSpace(body), nil
}
