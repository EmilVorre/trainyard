package scaffold

import "fmt"

type StackType int

const (
	StackSingle            StackType = iota
	StackAppDB
	StackFrontendBackend
	StackFrontendBackendDB
	StackCustom
)

type StackOption struct {
	Label       string
	Description string
	Stack       StackType
}

var StackOptions = []StackOption{
	{"Single service", "One container — typical for a monolith or API", StackSingle},
	{"App + Database", "Your app alongside a Postgres sidecar", StackAppDB},
	{"Frontend + Backend", "Two public/private services with internal DNS", StackFrontendBackend},
	{"Frontend + Backend + Database", "Full stack — frontend, API, and Postgres", StackFrontendBackendDB},
	{"Custom", "I'll define services myself in the generated config", StackCustom},
}

type ServiceDef struct {
	Name       string
	Port       int
	Public     bool
	Subdomain  string
	Dockerfile string
	Image      string
	DependsOn  []string
	EnvVars    []EnvVar
	CPU        string
	Memory     string
}

type EnvVar struct {
	Name  string
	Value string
}

type Config struct {
	Domain       string
	IngressClass string
	TLS          bool
	Label        string
	Stack        StackType
	Services     []ServiceDef
	AppPort      int
}

func DefaultConfig() Config {
	return Config{
		Domain:       "preview.vorre.dev",
		IngressClass: "nginx",
		TLS:          true,
		Label:        "preview",
		AppPort:      3000,
	}
}

func ServicesForStack(cfg Config) []ServiceDef {
	switch cfg.Stack {
	case StackSingle:
		return []ServiceDef{{
			Name: "app", Port: cfg.AppPort, Public: true, Subdomain: "pr-{number}",
			CPU: "250m", Memory: "128Mi", DependsOn: []string{},
			EnvVars: []EnvVar{{Name: "PORT", Value: fmt.Sprintf("%d", cfg.AppPort)}},
		}}
	case StackAppDB:
		return []ServiceDef{
			{
				Name: "app", Port: cfg.AppPort, Public: true, Subdomain: "pr-{number}",
				CPU: "250m", Memory: "128Mi", DependsOn: []string{"db"},
				EnvVars: []EnvVar{
					{Name: "PORT", Value: fmt.Sprintf("%d", cfg.AppPort)},
					{Name: "DATABASE_URL", Value: "postgres://postgres:postgres@db:5432/app"},
				},
			},
			{
				Name: "db", Image: "postgres:16-alpine", Port: 5432, Public: false,
				CPU: "250m", Memory: "256Mi", DependsOn: []string{},
				EnvVars: []EnvVar{
					{Name: "POSTGRES_USER", Value: "postgres"},
					{Name: "POSTGRES_PASSWORD", Value: "postgres"},
					{Name: "POSTGRES_DB", Value: "app"},
				},
			},
		}
	case StackFrontendBackend:
		return []ServiceDef{
			{
				Name: "frontend", Dockerfile: "frontend/Dockerfile", Port: 3000,
				Public: true, Subdomain: "pr-{number}", CPU: "250m", Memory: "128Mi",
				DependsOn: []string{"backend"},
				EnvVars: []EnvVar{
					{Name: "PORT", Value: "3000"},
					{Name: "API_URL", Value: "http://backend:8080"},
				},
			},
			{
				Name: "backend", Dockerfile: "backend/Dockerfile", Port: 8080,
				Public: false, CPU: "250m", Memory: "256Mi", DependsOn: []string{},
				EnvVars: []EnvVar{{Name: "PORT", Value: "8080"}},
			},
		}
	case StackFrontendBackendDB:
		return []ServiceDef{
			{
				Name: "frontend", Dockerfile: "frontend/Dockerfile", Port: 3000,
				Public: true, Subdomain: "pr-{number}", CPU: "250m", Memory: "128Mi",
				DependsOn: []string{"backend"},
				EnvVars: []EnvVar{
					{Name: "PORT", Value: "3000"},
					{Name: "API_URL", Value: "http://backend:8080"},
				},
			},
			{
				Name: "backend", Dockerfile: "backend/Dockerfile", Port: 8080,
				Public: false, CPU: "250m", Memory: "256Mi", DependsOn: []string{"db"},
				EnvVars: []EnvVar{
					{Name: "PORT", Value: "8080"},
					{Name: "DATABASE_URL", Value: "postgres://postgres:postgres@db:5432/app"},
				},
			},
			{
				Name: "db", Image: "postgres:16-alpine", Port: 5432, Public: false,
				CPU: "250m", Memory: "256Mi", DependsOn: []string{},
				EnvVars: []EnvVar{
					{Name: "POSTGRES_USER", Value: "postgres"},
					{Name: "POSTGRES_PASSWORD", Value: "postgres"},
					{Name: "POSTGRES_DB", Value: "app"},
				},
			},
		}
	default:
		return []ServiceDef{{
			Name: "app", Port: cfg.AppPort, Public: true, Subdomain: "pr-{number}",
			CPU: "250m", Memory: "128Mi", DependsOn: []string{},
			EnvVars: []EnvVar{{Name: "PORT", Value: fmt.Sprintf("%d", cfg.AppPort)}},
		}}
	}
}
