package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
)

var graphifyDep = Dependency{
	Name:           "graphify",
	Description:    "Knowledge graph builder — queryable graph from your entire codebase",
	Type:           DepTypeShell,
	Source:         "graphifyy",
	VersionCmd:     "graphify --version",
	InstallCmd:     "pipx install graphifyy",
	PostInstallCmd: "$HOME/.local/bin/graphify install",
	PostInstallMsg: "Run '/graphify' in Claude Code to build your first knowledge graph.",
}

func promptGraphifyInstall() (bool, error) {
	fmt.Println(sectionHeader("Optional: Graphify — Knowledge Graph for Your Codebase"))
	fmt.Println(dimStyle.Render("  Builds a queryable knowledge graph from your codebase, docs, and media."))
	fmt.Println(dimStyle.Render("  AI queries traverse the graph instead of re-reading raw files each time."))
	fmt.Println()

	pros := []string{
		"Code never leaves your machine (tree-sitter AST extraction, no LLM)",
		"Up to 71.5× fewer tokens per query via graph traversal (self-reported)",
		"23 languages supported (Python, TS, Go, Rust, Java, C, etc.)",
		"Local audio/video transcription via faster-whisper",
		"Incremental rebuilds — only reprocesses changed files (SHA256 cache)",
		"MCP server, Obsidian vault, Neo4j, and HTML graph exports",
	}
	fmt.Println(accentStyle.Render("  Pros:"))
	for _, p := range pros {
		fmt.Printf("  %s %s\n", successStyle.Render("+"), dimStyle.Render(p))
	}
	fmt.Println()

	cons := []string{
		"Single-developer project (v0.4.x, no institutional backing)",
		"PyPI package is 'graphifyy' (double-y) — verify name before installing",
		"Documents and images ARE sent to your AI API provider (Pass 3)",
		"First ingest can be costly on large document-heavy repos",
		"No independent accuracy benchmarks for semantic extraction",
		"Overkill for small projects (<20 files)",
	}
	fmt.Println(accentStyle.Render("  Cons:"))
	for _, c := range cons {
		fmt.Printf("  %s %s\n", errorStyle.Render("-"), dimStyle.Render(c))
	}
	fmt.Println()

	var install bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Install Graphify globally (~/.claude)?").
				Value(&install),
		),
	).WithTheme(ckTheme())

	if err := form.Run(); err != nil {
		return false, err
	}

	return install, nil
}
