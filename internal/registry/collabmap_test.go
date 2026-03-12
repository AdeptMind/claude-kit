package registry

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCollaborationMaps_Bidirectional(t *testing.T) {
	agents := []AgentEntry{
		{
			Name: "backend",
			Interfaces: AgentInterfaces{
				Produces: []string{"API endpoints"},
				Consumes: []string{"architecture.yaml"},
			},
		},
		{
			Name: "architect",
			Interfaces: AgentInterfaces{
				Produces: []string{"architecture.yaml"},
				Consumes: []string{"API endpoints"},
			},
		},
	}

	maps := BuildCollaborationMaps(agents)

	// backend produces "API endpoints" which architect consumes → backend provides_to architect
	// backend consumes "architecture.yaml" which architect produces → backend depends_on architect
	backendEntries := maps["backend"]
	if len(backendEntries) != 2 {
		t.Fatalf("expected 2 entries for backend, got %d: %+v", len(backendEntries), backendEntries)
	}

	// Sorted: architect/depends_on comes before architect/provides_to
	if backendEntries[0].Agent != "architect" || backendEntries[0].Direction != "depends_on" {
		t.Errorf("backend[0] = %+v, want depends_on architect", backendEntries[0])
	}
	if backendEntries[0].Artifacts[0] != "architecture.yaml" {
		t.Errorf("backend[0].Artifacts = %v, want [architecture.yaml]", backendEntries[0].Artifacts)
	}
	if backendEntries[1].Agent != "architect" || backendEntries[1].Direction != "provides_to" {
		t.Errorf("backend[1] = %+v, want provides_to architect", backendEntries[1])
	}

	// architect produces "architecture.yaml" which backend consumes → architect provides_to backend
	// architect consumes "API endpoints" which backend produces → architect depends_on backend
	architectEntries := maps["architect"]
	if len(architectEntries) != 2 {
		t.Fatalf("expected 2 entries for architect, got %d: %+v", len(architectEntries), architectEntries)
	}

	if architectEntries[0].Agent != "backend" || architectEntries[0].Direction != "depends_on" {
		t.Errorf("architect[0] = %+v, want depends_on backend", architectEntries[0])
	}
	if architectEntries[1].Agent != "backend" || architectEntries[1].Direction != "provides_to" {
		t.Errorf("architect[1] = %+v, want provides_to backend", architectEntries[1])
	}
}

func TestBuildCollaborationMaps_NoOverlap(t *testing.T) {
	agents := []AgentEntry{
		{
			Name: "backend",
			Interfaces: AgentInterfaces{
				Produces: []string{"API endpoints"},
				Consumes: []string{},
			},
		},
		{
			Name: "frontend",
			Interfaces: AgentInterfaces{
				Produces: []string{"UI components"},
				Consumes: []string{"design tokens"},
			},
		},
	}

	maps := BuildCollaborationMaps(agents)

	if len(maps) != 0 {
		t.Errorf("expected empty map, got %+v", maps)
	}
}

func TestBuildCollaborationMaps_MultipleOverlaps(t *testing.T) {
	agents := []AgentEntry{
		{
			Name: "backend",
			Interfaces: AgentInterfaces{
				Produces: []string{"API endpoints", "**/*.go", "OpenAPI spec"},
				Consumes: []string{},
			},
		},
		{
			Name: "frontend",
			Interfaces: AgentInterfaces{
				Produces: []string{},
				Consumes: []string{"API endpoints", "OpenAPI spec"},
			},
		},
	}

	maps := BuildCollaborationMaps(agents)

	backendEntries := maps["backend"]
	if len(backendEntries) != 1 {
		t.Fatalf("expected 1 entry for backend, got %d: %+v", len(backendEntries), backendEntries)
	}
	if backendEntries[0].Direction != "provides_to" || backendEntries[0].Agent != "frontend" {
		t.Errorf("backend[0] = %+v, want provides_to frontend", backendEntries[0])
	}
	if len(backendEntries[0].Artifacts) != 2 {
		t.Errorf("expected 2 shared artifacts, got %v", backendEntries[0].Artifacts)
	}
	if backendEntries[0].Artifacts[0] != "API endpoints" || backendEntries[0].Artifacts[1] != "OpenAPI spec" {
		t.Errorf("artifacts = %v, want [API endpoints, OpenAPI spec]", backendEntries[0].Artifacts)
	}

	frontendEntries := maps["frontend"]
	if len(frontendEntries) != 1 {
		t.Fatalf("expected 1 entry for frontend, got %d: %+v", len(frontendEntries), frontendEntries)
	}
	if frontendEntries[0].Direction != "depends_on" || frontendEntries[0].Agent != "backend" {
		t.Errorf("frontend[0] = %+v, want depends_on backend", frontendEntries[0])
	}
}

func TestInjectCollaborationMaps_NewSection(t *testing.T) {
	dir := t.TempDir()
	content := "---\nname: backend\n---\n\n## Rules\n- DRY\n"
	if err := os.WriteFile(filepath.Join(dir, "backend.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	maps := map[string][]CollabEntry{
		"backend": {
			{Agent: "frontend", Direction: "provides_to", Artifacts: []string{"API endpoints"}},
		},
	}

	if err := InjectCollaborationMaps(dir, maps); err != nil {
		t.Fatalf("InjectCollaborationMaps error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "backend.md"))
	if err != nil {
		t.Fatal(err)
	}
	result := string(data)

	if !strings.Contains(result, "## Collaboration Map") {
		t.Error("missing ## Collaboration Map header")
	}
	if !strings.Contains(result, "| frontend | provides_to | API endpoints |") {
		t.Errorf("missing table row in:\n%s", result)
	}
	if !strings.Contains(result, "## Rules") {
		t.Error("original content was removed")
	}
}

func TestInjectCollaborationMaps_UpdateExisting(t *testing.T) {
	dir := t.TempDir()
	content := `---
name: backend
---

## Rules
- DRY

## Collaboration Map

> Auto-generated by ` + "`ck agents registry --update`" + `. Do not edit manually.

| Agent | Direction | Shared Artifacts |
|-------|-----------|------------------|
| old-agent | provides_to | old-artifact |

## Other Section
- keep this
`
	if err := os.WriteFile(filepath.Join(dir, "backend.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	maps := map[string][]CollabEntry{
		"backend": {
			{Agent: "frontend", Direction: "provides_to", Artifacts: []string{"API endpoints"}},
		},
	}

	if err := InjectCollaborationMaps(dir, maps); err != nil {
		t.Fatalf("InjectCollaborationMaps error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "backend.md"))
	if err != nil {
		t.Fatal(err)
	}
	result := string(data)

	// Old content replaced
	if strings.Contains(result, "old-agent") {
		t.Error("old collaboration entries not removed")
	}

	// New content present
	if !strings.Contains(result, "| frontend | provides_to | API endpoints |") {
		t.Errorf("new table row missing in:\n%s", result)
	}

	// Other sections preserved
	if !strings.Contains(result, "## Other Section") {
		t.Errorf("## Other Section was removed in:\n%s", result)
	}
	if !strings.Contains(result, "## Rules") {
		t.Error("## Rules was removed")
	}
}

func TestInjectCollaborationMaps_PreservesContent(t *testing.T) {
	dir := t.TempDir()
	original := "---\nname: backend\n---\n\n## Rules\n- DRY\n- KISS\n\n## Skills\n- code-reviewer\n"
	if err := os.WriteFile(filepath.Join(dir, "backend.md"), []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	maps := map[string][]CollabEntry{
		"backend": {
			{Agent: "frontend", Direction: "depends_on", Artifacts: []string{"UI specs"}},
		},
	}

	if err := InjectCollaborationMaps(dir, maps); err != nil {
		t.Fatalf("InjectCollaborationMaps error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "backend.md"))
	if err != nil {
		t.Fatal(err)
	}
	result := string(data)

	// All original sections must be intact
	if !strings.Contains(result, "## Rules\n- DRY\n- KISS") {
		t.Errorf("Rules section modified in:\n%s", result)
	}
	if !strings.Contains(result, "## Skills\n- code-reviewer") {
		t.Errorf("Skills section modified in:\n%s", result)
	}

	// Collaboration map added
	if !strings.Contains(result, "## Collaboration Map") {
		t.Error("Collaboration Map not added")
	}
}
