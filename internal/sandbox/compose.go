package sandbox

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
)

const composeTpl = `version: "3"
services:
  claude:
    build:
      context: .
      dockerfile: Containerfile
    container_name: ck-sandbox-{{.ProjectName}}
    volumes:
      - "{{.ProjectDir}}:/workspace:ro"
{{- range .WritableMounts}}
      - "{{$.ProjectDir}}/{{.}}:/workspace/{{.}}:rw"
{{- end}}
{{- if eq .NetworkMode "block_all"}}
    network_mode: "none"
{{- end}}
{{- if .EnvPassthrough}}
    environment:
{{- range .EnvPassthrough}}
      - "{{.}}"
{{- end}}
{{- end}}
    working_dir: /workspace
    stdin_open: true
    tty: true
`

type composeData struct {
	ProjectName    string
	ProjectDir     string
	WritableMounts []string
	NetworkMode    string
	EnvPassthrough []string
}

// GenerateCompose produces a compose.yaml string from a PolicySpec and project directory.
func GenerateCompose(spec *policy.PolicySpec, projectDir string) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("policy spec is nil")
	}

	data := composeData{
		ProjectName: filepath.Base(projectDir),
		ProjectDir:  projectDir,
	}

	if spec.Sandbox != nil {
		data.WritableMounts = spec.Sandbox.WritableMounts
		data.EnvPassthrough = spec.Sandbox.EnvPassthrough
	}

	if spec.Network != nil {
		data.NetworkMode = spec.Network.Mode
	}

	t, err := template.New("compose").Parse(composeTpl)
	if err != nil {
		return "", fmt.Errorf("parse compose template: %w", err)
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute compose template: %w", err)
	}

	return buf.String(), nil
}
