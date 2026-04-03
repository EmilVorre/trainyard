package scaffold

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Emilvorre/trainyard/internal/tui"
	"github.com/spf13/cobra"
)

// Command returns the cobra command for `yard init`.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold PR preview config and workflow for this repo",
		Long: `Interactive wizard that generates two files in your repo:

  .github/pr-preview.yml          — service definitions for the preview environment
  .github/workflows/preview.yml   — GitHub Actions workflow that calls Trainyard

Run this from the root of the consuming repo.`,
		RunE: runInit,
	}

	cmd.Flags().Bool("force", false, "Overwrite existing files without prompting")
	cmd.Flags().String("domain", "", "Preview domain (skips prompt)")
	cmd.Flags().String("stack", "", "Stack type: single|app+db|frontend+backend|frontend+backend+db|custom")

	return cmd
}

func runInit(cmd *cobra.Command, _ []string) error {
	force, _ := cmd.Flags().GetBool("force")
	domainFlag, _ := cmd.Flags().GetString("domain")
	stackFlag, _ := cmd.Flags().GetString("stack")

	tui.Banner()
	fmt.Println("  This wizard generates the Trainyard config and workflow for this repo.")
	fmt.Println()

	cfg := DefaultConfig()

	// ── Domain ───────────────────────────────────────────────────────────────
	tui.Section("Domain")
	if domainFlag != "" {
		cfg.Domain = domainFlag
		tui.Info("Using domain: %s", cfg.Domain)
	} else {
		cfg.Domain = tui.Prompt("Preview domain", cfg.Domain)
	}

	// ── Stack type ────────────────────────────────────────────────────────────
	tui.Section("Stack type")

	if stackFlag != "" {
		cfg.Stack = parseStackFlag(stackFlag)
	} else {
		cfg.Stack = pickStack()
	}

	// ── Stack-specific prompts ────────────────────────────────────────────────
	tui.Section("Service configuration")

	switch cfg.Stack {
	case StackSingle, StackAppDB, StackCustom:
		portStr := tui.Prompt("App port", strconv.Itoa(cfg.AppPort))
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.AppPort = p
		}

	case StackFrontendBackend, StackFrontendBackendDB:
		tui.Info("Frontend port assumed: 3000")
		tui.Info("Backend port assumed:  8080")
		tui.Info("(Edit .github/pr-preview.yml to change)")
	}

	// ── Ingress / TLS ─────────────────────────────────────────────────────────
	tui.Section("Ingress")
	cfg.IngressClass = tui.Prompt("Ingress class", cfg.IngressClass)
	tlsStr := tui.Prompt("Enable TLS? (true/false)", "true")
	cfg.TLS = strings.ToLower(tlsStr) == "true"

	// ── Label ─────────────────────────────────────────────────────────────────
	cfg.Label = tui.Prompt("PR label to trigger previews", cfg.Label)

	// ── Build service list ────────────────────────────────────────────────────
	cfg.Services = ServicesForStack(cfg)

	// ── Preview ───────────────────────────────────────────────────────────────
	tui.Section("Files to generate")
	files := []OutputFile{
		{
			Path:    ".github/pr-preview.yml",
			Content: RenderPreviewConfig(cfg),
		},
		{
			Path:    ".github/workflows/preview.yml",
			Content: RenderWorkflow(cfg),
		},
	}

	for _, f := range files {
		fmt.Printf("  • %s\n", f.Path)
	}
	fmt.Println()

	if !tui.Confirm("Generate these files?") {
		fmt.Println("  Aborted.")
		os.Exit(0)
	}

	// ── Write ─────────────────────────────────────────────────────────────────
	written, err := WriteFiles(files, force)
	if err != nil {
		tui.Fatal("Failed to write files: %v", err)
	}

	// ── Done ──────────────────────────────────────────────────────────────────
	fmt.Println()
	for _, path := range written {
		tui.Success("Written: %s", path)
	}

	fmt.Println()
	tui.Box("Next steps", []string{
		"1. Add KUBE_CONFIG secret to your repo:",
		"   Settings → Secrets and variables → Actions → New repository secret",
		"",
		"2. Review .github/pr-preview.yml and adjust resources/env as needed",
		"",
		"3. Open a PR and add the `" + cfg.Label + "` label",
		"",
		"4. Watch the Actions tab — a preview URL will be posted as a PR comment",
	})

	return nil
}

// pickStack presents a numbered menu and returns the chosen StackType.
func pickStack() StackType {
	for i, opt := range StackOptions {
		fmt.Printf("  [%d] %-30s %s\n", i+1, opt.Label, opt.Description)
	}
	fmt.Println()

	for {
		choice := tui.Prompt("Choose stack", "1")
		n, err := strconv.Atoi(strings.TrimSpace(choice))
		if err != nil || n < 1 || n > len(StackOptions) {
			tui.Warn("Enter a number between 1 and %d", len(StackOptions))
			continue
		}
		selected := StackOptions[n-1]
		tui.Info("Selected: %s", selected.Label)
		return selected.Stack
	}
}

// parseStackFlag converts a --stack flag value to a StackType.
func parseStackFlag(s string) StackType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "single":
		return StackSingle
	case "app+db", "app-db":
		return StackAppDB
	case "frontend+backend", "frontend-backend":
		return StackFrontendBackend
	case "frontend+backend+db", "frontend-backend-db":
		return StackFrontendBackendDB
	case "custom":
		return StackCustom
	default:
		tui.Warn("Unknown stack %q — defaulting to single", s)
		return StackSingle
	}
}
