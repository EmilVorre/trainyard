package validate

import (
	"strings"
	"testing"
)

// helpers

func validConfig() previewConfig {
	cfg := previewConfig{}
	cfg.App.Domain = "preview.example.com"
	cfg.Label = "preview"
	cfg.Services = []serviceConfig{
		{
			Name:   "app",
			Build:  &build{Context: ".", Dockerfile: "Dockerfile"},
			Port:   3000,
			Public: true, Subdomain: "pr-{number}",
		},
	}
	return cfg
}

func assertNoErrors(t *testing.T, cfg previewConfig) {
	t.Helper()
	errs := runChecks(cfg)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func assertHasError(t *testing.T, cfg previewConfig, fragment string) {
	t.Helper()
	errs := runChecks(cfg)
	for _, e := range errs {
		if strings.Contains(e, fragment) {
			return
		}
	}
	t.Errorf("expected error containing %q, got: %v", fragment, errs)
}

// valid config

func TestRunChecks_ValidConfig(t *testing.T) {
	assertNoErrors(t, validConfig())
}

func TestRunChecks_ValidConfig_ImageService(t *testing.T) {
	cfg := validConfig()
	cfg.Services = []serviceConfig{
		{
			Name:   "app",
			Build:  &build{Context: ".", Dockerfile: "Dockerfile"},
			Port:   3000,
			Public: true, Subdomain: "pr-{number}",
		},
		{
			Name:   "db",
			Image:  "postgres:16-alpine",
			Port:   5432,
			Public: false,
		},
	}
	assertNoErrors(t, cfg)
}

func TestRunChecks_ValidConfig_PrivateServiceNoSubdomain(t *testing.T) {
	cfg := validConfig()
	cfg.Services = append(cfg.Services, serviceConfig{
		Name:   "backend",
		Build:  &build{Context: ".", Dockerfile: "Dockerfile"},
		Port:   8080,
		Public: false,
	})
	assertNoErrors(t, cfg)
}

func TestRunChecks_ValidConfig_DependsOn(t *testing.T) {
	cfg := validConfig()
	cfg.Services = []serviceConfig{
		{
			Name:      "app",
			Build:     &build{Context: "."},
			Port:      3000,
			Public:    true,
			Subdomain: "pr-{number}",
			DependsOn: []string{"db"},
		},
		{
			Name:   "db",
			Image:  "postgres:16-alpine",
			Port:   5432,
			Public: false,
		},
	}
	assertNoErrors(t, cfg)
}

// domain

func TestRunChecks_MissingDomain(t *testing.T) {
	cfg := validConfig()
	cfg.App.Domain = ""
	assertHasError(t, cfg, "app.domain is required")
}

func TestRunChecks_WhitespaceDomain(t *testing.T) {
	cfg := validConfig()
	cfg.App.Domain = "   "
	assertHasError(t, cfg, "app.domain is required")
}

// label

func TestRunChecks_MissingLabel(t *testing.T) {
	cfg := validConfig()
	cfg.Label = ""
	assertHasError(t, cfg, "label is required")
}

func TestRunChecks_WhitespaceLabel(t *testing.T) {
	cfg := validConfig()
	cfg.Label = "   "
	assertHasError(t, cfg, "label is required")
}

// services

func TestRunChecks_NoServices(t *testing.T) {
	cfg := validConfig()
	cfg.Services = []serviceConfig{}
	assertHasError(t, cfg, "at least one service must be defined")
}

func TestRunChecks_NoServices_EarlyReturn(t *testing.T) {
	// when services is empty, label error should not also be reported
	// (early return after no-services check)
	cfg := validConfig()
	cfg.Services = []serviceConfig{}
	cfg.Label = ""
	errs := runChecks(cfg)
	for _, e := range errs {
		if strings.Contains(e, "label is required") {
			t.Error("expected early return after no-services error, but got label error too")
		}
	}
}

func TestRunChecks_MissingServiceName(t *testing.T) {
	cfg := validConfig()
	cfg.Services = []serviceConfig{
		{Name: "", Build: &build{Context: "."}, Port: 3000, Public: false},
	}
	assertHasError(t, cfg, "missing its `name` field")
}

func TestRunChecks_DuplicateServiceName(t *testing.T) {
	cfg := validConfig()
	cfg.Services = []serviceConfig{
		{Name: "app", Build: &build{Context: "."}, Port: 3000, Public: true, Subdomain: "pr-{number}"},
		{Name: "app", Build: &build{Context: "."}, Port: 3001, Public: false},
	}
	assertHasError(t, cfg, "duplicate service name")
}

// image / build

func TestRunChecks_NeitherImageNorBuild(t *testing.T) {
	cfg := validConfig()
	cfg.Services = []serviceConfig{
		{Name: "app", Port: 3000, Public: true, Subdomain: "pr-{number}"},
	}
	assertHasError(t, cfg, "must have either `image` or `build`")
}

func TestRunChecks_BothImageAndBuild(t *testing.T) {
	cfg := validConfig()
	cfg.Services = []serviceConfig{
		{
			Name:      "app",
			Image:     "myimage:latest",
			Build:     &build{Context: "."},
			Port:      3000,
			Public:    true,
			Subdomain: "pr-{number}",
		},
	}
	assertHasError(t, cfg, "cannot have both `image` and `build`")
}

// port

func TestRunChecks_InvalidPort_Zero(t *testing.T) {
	cfg := validConfig()
	cfg.Services[0].Port = 0
	assertHasError(t, cfg, "invalid port 0")
}

func TestRunChecks_InvalidPort_Negative(t *testing.T) {
	cfg := validConfig()
	cfg.Services[0].Port = -1
	assertHasError(t, cfg, "invalid port -1")
}

func TestRunChecks_InvalidPort_TooHigh(t *testing.T) {
	cfg := validConfig()
	cfg.Services[0].Port = 65536
	assertHasError(t, cfg, "invalid port 65536")
}

func TestRunChecks_ValidPort_Boundaries(t *testing.T) {
	for _, port := range []int{1, 80, 443, 3000, 8080, 65535} {
		cfg := validConfig()
		cfg.Services[0].Port = port
		errs := runChecks(cfg)
		for _, e := range errs {
			if strings.Contains(e, "invalid port") {
				t.Errorf("port %d should be valid, got error: %s", port, e)
			}
		}
	}
}

// subdomain

func TestRunChecks_PublicServiceMissingSubdomain(t *testing.T) {
	cfg := validConfig()
	cfg.Services[0].Subdomain = ""
	assertHasError(t, cfg, "is public but has no subdomain")
}

func TestRunChecks_PublicServiceWhitespaceSubdomain(t *testing.T) {
	cfg := validConfig()
	cfg.Services[0].Subdomain = "   "
	assertHasError(t, cfg, "is public but has no subdomain")
}

func TestRunChecks_SubdomainMissingNumberPlaceholder(t *testing.T) {
	cfg := validConfig()
	cfg.Services[0].Subdomain = "my-preview"
	assertHasError(t, cfg, "should contain {number}")
}

func TestRunChecks_PrivateServiceWithSubdomainNoNumberPlaceholder(t *testing.T) {
	// private services can have a subdomain, but it still must contain {number}
	cfg := validConfig()
	cfg.Services = append(cfg.Services, serviceConfig{
		Name:      "admin",
		Build:     &build{Context: "."},
		Port:      8080,
		Public:    false,
		Subdomain: "admin-panel",
	})
	assertHasError(t, cfg, "should contain {number}")
}

// dependsOn

func TestRunChecks_DependsOnUnknownService(t *testing.T) {
	cfg := validConfig()
	cfg.Services[0].DependsOn = []string{"nonexistent"}
	assertHasError(t, cfg, "references unknown service")
}

func TestRunChecks_DependsOnSelf(t *testing.T) {
	cfg := validConfig()
	cfg.Services[0].DependsOn = []string{"app"}
	assertHasError(t, cfg, "cannot depend on itself")
}

// multiple errors

func TestRunChecks_MultipleErrors(t *testing.T) {
	cfg := previewConfig{}
	cfg.App.Domain = ""
	cfg.Label = ""
	cfg.Services = []serviceConfig{
		{Name: "app", Port: 0, Public: true, Subdomain: ""},
	}
	errs := runChecks(cfg)
	if len(errs) < 3 {
		t.Errorf("expected multiple errors, got %d: %v", len(errs), errs)
	}
}

func TestRunChecks_ErrorCount(t *testing.T) {
	cfg := validConfig()
	errs := runChecks(cfg)
	if len(errs) != 0 {
		t.Errorf("valid config produced %d errors: %v", len(errs), errs)
	}
}
