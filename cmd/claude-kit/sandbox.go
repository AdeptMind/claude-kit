package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/AdeptMind/infra-tool/claude-cli/internal/policy"
	"github.com/AdeptMind/infra-tool/claude-cli/internal/sandbox"
)

var sandboxCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Manage sandboxed execution environment",
}

var sandboxCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate Containerfile and compose.yaml from policy",
	RunE: func(cmd *cobra.Command, args []string) error {
		runtime, _ := cmd.Flags().GetString("runtime")
		return runSandboxCreate(runtime)
	},
}

var sandboxRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Build and start the sandbox container",
	RunE: func(cmd *cobra.Command, args []string) error {
		runtime, _ := cmd.Flags().GetString("runtime")
		return runSandboxRun(runtime)
	},
}

var sandboxStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop and remove the sandbox container",
	RunE: func(cmd *cobra.Command, args []string) error {
		runtime, _ := cmd.Flags().GetString("runtime")
		return runSandboxStop(runtime)
	},
}

func init() {
	sandboxCmd.PersistentFlags().String("runtime", "", "Container runtime override: docker, podman, or nerdctl")

	sandboxCmd.AddCommand(sandboxCreateCmd)
	sandboxCmd.AddCommand(sandboxRunCmd)
	sandboxCmd.AddCommand(sandboxStopCmd)
}

func sandboxDir() string {
	return filepath.Join(resolveProjectRoot(), ".claude", "sandbox")
}

func policyPath() string {
	return filepath.Join(resolveProjectRoot(), ".claude", "policy.yaml")
}

func runSandboxCreate(runtimeOverride string) error {
	spec, err := policy.ParseFile(policyPath())
	if err != nil {
		return fmt.Errorf("failed to read policy: %w", err)
	}

	if spec.Sandbox == nil || !spec.Sandbox.Enabled {
		return fmt.Errorf("sandbox not enabled in policy.yaml; set sandbox.enabled: true")
	}

	rt, rtErr := sandbox.Detect(runtimeOverride)
	if rtErr != nil {
		// No runtime found — fallback mode
		enhanced, warnings := sandbox.FallbackPolicy(spec)
		_ = enhanced // enhanced policy could be written; for now just warn

		fmt.Println(warnStyle.Render("No container runtime detected — entering fallback mode"))
		fmt.Println()
		fmt.Print(sandbox.FormatWarnings(warnings))
		return nil
	}

	// Generate Containerfile
	containerfile, err := sandbox.GenerateContainerfile(spec.Sandbox)
	if err != nil {
		return fmt.Errorf("generating Containerfile: %w", err)
	}

	// Generate compose.yaml
	compose, err := sandbox.GenerateCompose(spec, resolveProjectRoot())
	if err != nil {
		return fmt.Errorf("generating compose.yaml: %w", err)
	}

	// Write files
	dir := sandboxDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating sandbox directory: %w", err)
	}

	cfPath := filepath.Join(dir, "Containerfile")
	if err := os.WriteFile(cfPath, []byte(containerfile), 0o644); err != nil {
		return fmt.Errorf("writing Containerfile: %w", err)
	}

	cyPath := filepath.Join(dir, "compose.yaml")
	if err := os.WriteFile(cyPath, []byte(compose), 0o644); err != nil {
		return fmt.Errorf("writing compose.yaml: %w", err)
	}

	fmt.Println(successStyle.Render("Sandbox files generated"))
	fmt.Printf("  %s runtime: %s %s\n", bullet, accentStyle.Render(rt.Name), dimStyle.Render("v"+rt.Version))
	fmt.Printf("  %s %s\n", arrow, dimStyle.Render(cfPath))
	fmt.Printf("  %s %s\n", arrow, dimStyle.Render(cyPath))
	return nil
}

func runSandboxRun(runtimeOverride string) error {
	dir := sandboxDir()
	if _, err := os.Stat(filepath.Join(dir, "Containerfile")); os.IsNotExist(err) {
		return fmt.Errorf("sandbox not created; run 'ck sandbox create' first")
	}

	rt, err := sandbox.Detect(runtimeOverride)
	if err != nil {
		return err
	}

	fmt.Printf("%s Building sandbox image...\n", bullet)
	if err := sandbox.Build(rt, dir); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	fmt.Printf("%s Starting sandbox container...\n", bullet)
	if err := sandbox.Run(rt, dir); err != nil {
		return fmt.Errorf("run failed: %w", err)
	}

	fmt.Println(successStyle.Render("Sandbox running"))
	return nil
}

func runSandboxStop(runtimeOverride string) error {
	dir := sandboxDir()

	rt, err := sandbox.Detect(runtimeOverride)
	if err != nil {
		return err
	}

	fmt.Printf("%s Stopping sandbox...\n", bullet)
	if err := sandbox.Stop(rt, dir); err != nil {
		return fmt.Errorf("stop failed: %w", err)
	}

	fmt.Println(successStyle.Render("Sandbox stopped"))
	return nil
}
