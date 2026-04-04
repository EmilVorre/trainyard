# Changelog

All notable changes to Trainyard are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
Trainyard uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

[Unreleased]

---

## [1.1.1] — 2026-04-05

### Fix
- support plain image strings alongside repository/tag objects in deployment template

---

## [1.1.0] — 2026-04-05

### Added
- support for jobs for multi-services

--- 
 
## [1.0.0] — 2026-04-04
 
### Added
- Go unit tests for `internal/scaffold`, `internal/validate`, `internal/tui`, and `internal/setup`
- Helm unit tests — 80 tests across `pr-preview` and `cleanup` charts
- Chainsaw e2e tests — 4 tests against live k3s cluster (deploy, teardown, cleanup, multi-service)
- Unified `tests.yml` CI workflow consolidating all three test suites into a single required check:
  - Go unit tests (`go test ./internal/...`)
  - Helm unit tests (`helm unittest` on `pr-preview` and `cleanup` charts)
  - Chainsaw e2e tests (4 tests against live k3s cluster)
- CI and release workflows for yard
- Templates for issues and pull requests
- GoReleaser config that defines how the yard binary gets built and published
- Metadata file for Artifact Hub
 
---

## [0.1.3] — 2026-04-03

### Added
- `yard setup` — interactive server setup wizard
- `yard init` — repo scaffolder with 5 stack presets
- `yard validate` — config file linter
- `charts/cleanup` — stale environment cleanup CronJob
- Multi-provider DNS support in `yard setup` (Cloudflare, Route 53, Hetzner, DigitalOcean, Porkbun, Namecheap, GoDaddy, name.com, manual)

---

## [0.1.2] — 2026-04-01

### Added
- Helm chart `pr-preview` v0.1.2
- Multi-service support with `dependsOn` init containers
- Public/private service model with internal DNS
- Wildcard TLS via `nginx.ingress.kubernetes.io/ssl-redirect: "false"`

### Fixed
- Ingress not routing correctly when TLS redirect was enabled

---

## [0.1.1] — 2026-04-01

### Added
- GitHub Actions reusable workflows (`deploy.yml`, `teardown.yml`)
- PR comment with preview URL, updated on every push
- Teardown on label removal or PR close

---

## [0.1.0] — 2026-04-01

### Added
- Initial Helm chart `pr-preview`
- Single-service preview environment support
- Namespace-per-PR isolation
