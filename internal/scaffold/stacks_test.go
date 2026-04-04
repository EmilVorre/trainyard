package scaffold

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Domain != "preview.vorre.dev" {
		t.Errorf("expected default domain preview.vorre.dev, got %s", cfg.Domain)
	}
	if cfg.IngressClass != "nginx" {
		t.Errorf("expected default ingress class nginx, got %s", cfg.IngressClass)
	}
	if !cfg.TLS {
		t.Error("expected TLS to be true by default")
	}
	if cfg.Label != "preview" {
		t.Errorf("expected default label preview, got %s", cfg.Label)
	}
	if cfg.AppPort != 3000 {
		t.Errorf("expected default app port 3000, got %d", cfg.AppPort)
	}
}

func TestServicesForStack_Single(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stack = StackSingle
	cfg.AppPort = 3000

	svcs := ServicesForStack(cfg)

	if len(svcs) != 1 {
		t.Fatalf("expected 1 service, got %d", len(svcs))
	}
	svc := svcs[0]
	if svc.Name != "app" {
		t.Errorf("expected service name app, got %s", svc.Name)
	}
	if !svc.Public {
		t.Error("expected single service to be public")
	}
	if svc.Subdomain != "pr-{number}" {
		t.Errorf("expected subdomain pr-{number}, got %s", svc.Subdomain)
	}
	if svc.Port != 3000 {
		t.Errorf("expected port 3000, got %d", svc.Port)
	}
	if len(svc.DependsOn) != 0 {
		t.Errorf("expected no dependsOn, got %v", svc.DependsOn)
	}
}

func TestServicesForStack_AppDB(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stack = StackAppDB
	cfg.AppPort = 8080

	svcs := ServicesForStack(cfg)

	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}

	app := svcs[0]
	if app.Name != "app" {
		t.Errorf("expected first service name app, got %s", app.Name)
	}
	if !app.Public {
		t.Error("expected app service to be public")
	}
	if len(app.DependsOn) != 1 || app.DependsOn[0] != "db" {
		t.Errorf("expected app to depend on db, got %v", app.DependsOn)
	}

	db := svcs[1]
	if db.Name != "db" {
		t.Errorf("expected second service name db, got %s", db.Name)
	}
	if db.Public {
		t.Error("expected db service to be private")
	}
	if db.Image != "postgres:16-alpine" {
		t.Errorf("expected postgres image, got %s", db.Image)
	}
	if len(db.DependsOn) != 0 {
		t.Errorf("expected db to have no dependsOn, got %v", db.DependsOn)
	}
}

func TestServicesForStack_FrontendBackend(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stack = StackFrontendBackend

	svcs := ServicesForStack(cfg)

	if len(svcs) != 2 {
		t.Fatalf("expected 2 services, got %d", len(svcs))
	}

	fe := svcs[0]
	if fe.Name != "frontend" {
		t.Errorf("expected first service frontend, got %s", fe.Name)
	}
	if !fe.Public {
		t.Error("expected frontend to be public")
	}
	if fe.Port != 3000 {
		t.Errorf("expected frontend port 3000, got %d", fe.Port)
	}
	if len(fe.DependsOn) != 1 || fe.DependsOn[0] != "backend" {
		t.Errorf("expected frontend to depend on backend, got %v", fe.DependsOn)
	}

	be := svcs[1]
	if be.Name != "backend" {
		t.Errorf("expected second service backend, got %s", be.Name)
	}
	if be.Public {
		t.Error("expected backend to be private")
	}
	if be.Port != 8080 {
		t.Errorf("expected backend port 8080, got %d", be.Port)
	}
}

func TestServicesForStack_FrontendBackendDB(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stack = StackFrontendBackendDB

	svcs := ServicesForStack(cfg)

	if len(svcs) != 3 {
		t.Fatalf("expected 3 services, got %d", len(svcs))
	}

	names := []string{svcs[0].Name, svcs[1].Name, svcs[2].Name}
	expected := []string{"frontend", "backend", "db"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected service[%d] = %s, got %s", i, expected[i], name)
		}
	}

	be := svcs[1]
	if len(be.DependsOn) != 1 || be.DependsOn[0] != "db" {
		t.Errorf("expected backend to depend on db, got %v", be.DependsOn)
	}

	db := svcs[2]
	if db.Image != "postgres:16-alpine" {
		t.Errorf("expected postgres image, got %s", db.Image)
	}
}

func TestServicesForStack_Custom(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stack = StackCustom
	cfg.AppPort = 9000

	svcs := ServicesForStack(cfg)

	// Custom falls through to default — single app service
	if len(svcs) != 1 {
		t.Fatalf("expected 1 service for custom stack, got %d", len(svcs))
	}
	if svcs[0].Port != 9000 {
		t.Errorf("expected port 9000, got %d", svcs[0].Port)
	}
}

func TestServicesForStack_AppPortPropagation(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Stack = StackSingle
	cfg.AppPort = 4321

	svcs := ServicesForStack(cfg)

	if svcs[0].Port != 4321 {
		t.Errorf("expected port 4321, got %d", svcs[0].Port)
	}

	// PORT env var should match
	var portEnv string
	for _, e := range svcs[0].EnvVars {
		if e.Name == "PORT" {
			portEnv = e.Value
		}
	}
	if portEnv != "4321" {
		t.Errorf("expected PORT env var 4321, got %s", portEnv)
	}
}

func TestParseStackFlag(t *testing.T) {
	tests := []struct {
		input    string
		expected StackType
	}{
		{"single", StackSingle},
		{"SINGLE", StackSingle},
		{"app+db", StackAppDB},
		{"app-db", StackAppDB},
		{"frontend+backend", StackFrontendBackend},
		{"frontend-backend", StackFrontendBackend},
		{"frontend+backend+db", StackFrontendBackendDB},
		{"frontend-backend-db", StackFrontendBackendDB},
		{"custom", StackCustom},
		{"unknown", StackSingle}, // defaults to single
		{"", StackSingle},        // defaults to single
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseStackFlag(tt.input)
			if got != tt.expected {
				t.Errorf("parseStackFlag(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
