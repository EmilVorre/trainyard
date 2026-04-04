# Contributing to Trainyard

Thanks for your interest in contributing! This document covers how to get the project running locally and how to submit changes.

---

## Project Structure

```
trainyard/
  cmd/
    yard/                — CLI entrypoint
    cleanup/             — cleanup binary entrypoint
  internal/
    setup/               — yard setup implementation
    scaffold/            — yard init implementation
    validate/            — yard validate implementation
    tui/                 — shared terminal UI helpers
  charts/
    pr-preview/          — Helm chart deployed per PR
      tests/             — helm-unittest tests
    cleanup/             — Helm chart for stale env cleanup
      tests/             — helm-unittest tests
  tests/
    chainsaw/            — Chainsaw e2e tests
  .github/
    workflows/
      tests.yml          — unified CI: Go unit, Helm unit, Chainsaw e2e
      deploy.yml         — reusable deploy workflow
      teardown.yml       — reusable teardown workflow
      publish-*.yml      — chart/image publishing workflows
```

---

## Local Development

### Prerequisites

- Go 1.26+
- Helm 3
- kubectl
- A Kubernetes cluster (k3s locally via [k3d](https://k3d.io) works well)

### Build the CLI

```bash
git clone https://github.com/Emilvorre/trainyard
cd trainyard
go mod tidy
go build -o yard ./cmd/yard
./yard --help
```

### Run tests

```bash
go test ./internal/...
```

### Lint

```bash
go vet ./...
# Install golangci-lint: https://golangci-lint.run/usage/install/
golangci-lint run
```

### Test the Helm chart locally

```bash
# Lint
helm lint charts/pr-preview

# Render templates without deploying
helm template test-pr charts/pr-preview \
  --set global.prNumber=99 \
  --set global.domain=preview.example.com \
  -f charts/pr-preview/values.yaml
```

---

## Making Changes

### Bugs

1. Check existing issues first
2. Open an issue describing the bug and how to reproduce it
3. Submit a PR referencing the issue

### Features

1. Open an issue to discuss the feature before implementing
2. Keep PRs focused — one feature per PR
3. Update documentation if you're changing behaviour
4. Add tests where practical

### Helm chart changes

- Bump `version` in `Chart.yaml` for any change
- Bump `appVersion` only when the app itself changes
- Test with `helm lint` and `helm template` before submitting

### CLI changes

- Follow existing patterns in `internal/`
- Keep commands in their own package under `internal/`
- Update `cmd/yard/main.go` to register new commands

---

## Commit Style

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add support for Redis service type
fix: correct namespace label selector in cleanup chart
docs: update quick start in README
chore: bump k3s version in setup wizard
```

---

## Releasing

Releases are managed by maintainers via semver tags. Pushing a tag triggers GoReleaser and Helm chart publishing automatically.

```bash
git tag v0.2.0
git push origin v0.2.0
```
