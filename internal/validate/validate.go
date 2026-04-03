package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/Emilvorre/trainyard/internal/tui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const defaultConfigPath = ".github/pr-preview.yml"

func Command() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [config]",
		Short: "Validate a pr-preview.yml config file",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runValidate,
	}
}

type previewConfig struct {
	App struct {
		Domain string `yaml:"domain"`
	} `yaml:"app"`
	Label    string          `yaml:"label"`
	Services []serviceConfig `yaml:"services"`
}

type serviceConfig struct {
	Name      string   `yaml:"name"`
	Image     string   `yaml:"image"`
	Build     *build   `yaml:"build"`
	Port      int      `yaml:"port"`
	Public    bool     `yaml:"public"`
	Subdomain string   `yaml:"subdomain"`
	DependsOn []string `yaml:"dependsOn"`
}

type build struct {
	Context    string `yaml:"context"`
	Dockerfile string `yaml:"dockerfile"`
}

func runValidate(_ *cobra.Command, args []string) error {
	configPath := defaultConfigPath
	if len(args) == 1 {
		configPath = args[0]
	}

	tui.Section(fmt.Sprintf("Validating %s", configPath))

	data, err := os.ReadFile(configPath)
	if err != nil {
		tui.Fatal("Cannot read %s: %v", configPath, err)
	}

	var cfg previewConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		tui.Fatal("YAML parse error: %v", err)
	}

	errs := runChecks(cfg)

	if len(errs) == 0 {
		fmt.Println()
		tui.Success("Config is valid — %d services defined", len(cfg.Services))
		return nil
	}

	fmt.Println()
	for _, e := range errs {
		fmt.Printf("  ✗  %s\n", e)
	}
	fmt.Println()
	tui.Fatal("%d error(s) found in %s", len(errs), configPath)
	return nil
}

func runChecks(cfg previewConfig) []string {
	var errs []string

	if strings.TrimSpace(cfg.App.Domain) == "" {
		errs = append(errs, "app.domain is required")
	}
	if len(cfg.Services) == 0 {
		errs = append(errs, "at least one service must be defined under `services:`")
		return errs
	}
	if strings.TrimSpace(cfg.Label) == "" {
		errs = append(errs, "label is required (e.g. label: preview)")
	}

	names := map[string]int{}
	for _, svc := range cfg.Services {
		names[svc.Name]++
	}

	for _, svc := range cfg.Services {
		prefix := fmt.Sprintf("service %q:", svc.Name)

		if svc.Name == "" {
			errs = append(errs, "a service is missing its `name` field")
			continue
		}
		if names[svc.Name] > 1 {
			errs = append(errs, fmt.Sprintf("%s duplicate service name", prefix))
		}
		if svc.Image == "" && svc.Build == nil {
			errs = append(errs, fmt.Sprintf("%s must have either `image` or `build`", prefix))
		}
		if svc.Image != "" && svc.Build != nil {
			errs = append(errs, fmt.Sprintf("%s cannot have both `image` and `build`", prefix))
		}
		if svc.Port <= 0 || svc.Port > 65535 {
			errs = append(errs, fmt.Sprintf("%s invalid port %d", prefix, svc.Port))
		}
		if svc.Public && strings.TrimSpace(svc.Subdomain) == "" {
			errs = append(errs, fmt.Sprintf("%s is public but has no subdomain", prefix))
		}
		if svc.Subdomain != "" && !strings.Contains(svc.Subdomain, "{number}") {
			errs = append(errs, fmt.Sprintf("%s subdomain %q should contain {number}", prefix, svc.Subdomain))
		}
		for _, dep := range svc.DependsOn {
			if names[dep] == 0 {
				errs = append(errs, fmt.Sprintf("%s dependsOn references unknown service %q", prefix, dep))
			}
			if dep == svc.Name {
				errs = append(errs, fmt.Sprintf("%s cannot depend on itself", prefix))
			}
		}
	}

	return errs
}
