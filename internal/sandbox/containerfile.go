package sandbox

import (
	"bytes"
	"text/template"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
)

const defaultImageBase = "node:22-slim"

var containerfileTmpl = template.Must(template.New("Containerfile").Parse(`FROM {{.ImageBase}}

RUN groupadd -r claude && useradd -r -g claude -m claude
RUN npm install -g @anthropic-ai/claude-code

WORKDIR /workspace
USER claude

CMD ["claude"]
`))

// GenerateContainerfile renders an OCI Containerfile from the given sandbox policy.
// If policy is nil or ImageBase is empty, "node:22-slim" is used as the base image.
func GenerateContainerfile(p *policy.SandboxPolicy) (string, error) {
	imageBase := defaultImageBase
	if p != nil && p.ImageBase != "" {
		imageBase = p.ImageBase
	}

	data := struct{ ImageBase string }{ImageBase: imageBase}

	var buf bytes.Buffer
	if err := containerfileTmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
