# Trainyard

**Ephemeral Kubernetes PR preview environments — simple, self-hosted, zero lock-in.**

Trainyard spins up a full preview environment for every pull request, tears it down when the PR closes, and runs entirely on infrastructure you own.

[![Tests](https://github.com/Emilvorre/trainyard/actions/workflows/tests.yml/badge.svg)](https://github.com/Emilvorre/trainyard/actions/workflows/tests.yml)
[![CI](https://github.com/Emilvorre/trainyard/actions/workflows/ci.yml/badge.svg)](https://github.com/Emilvorre/trainyard/actions/workflows/ci.yml)
[![CodeQL](https://github.com/Emilvorre/trainyard/actions/workflows/codeql.yml/badge.svg)](https://github.com/Emilvorre/trainyard/actions/workflows/codeql.yml)
[![Release](https://github.com/Emilvorre/trainyard/actions/workflows/release.yml/badge.svg)](https://github.com/Emilvorre/trainyard/actions/workflows/release.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/Emilvorre/trainyard)](https://goreportcard.com/report/github.com/Emilvorre/trainyard)
[![Helm Chart](https://img.shields.io/badge/Helm-oci%3A%2F%2Fghcr.io-blue)](https://github.com/Emilvorre/trainyard/pkgs/container/trainyard%2Fcharts%2Fpr-preview)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/trainyard)](https://artifacthub.io/packages/search?repo=trainyard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/Emilvorre/trainyard)](https://github.com/Emilvorre/trainyard/releases)

---

## What is Trainyard?

Trainyard gives every PR its own live URL — `pr-42.preview.yourdomain.com` — deployed automatically when you add a label. No Vercel, no Render, no per-seat pricing. Just a VPS, k3s, and a Helm chart.

- **Multi-service** — frontend, backend, database, all wired together
- **Self-hosted** — runs on any Linux VPS from ~€5/month
- **GitOps-native** — driven entirely by GitHub Actions and labels
- **Zero lock-in** — standard Kubernetes, standard Helm, your own domain

---

## Quick Start

### 1. Install the CLI

```bash
# macOS / Linux (Homebrew)
brew install Emilvorre/tap/yard

# Or download a binary from releases
curl -fsSL https://github.com/Emilvorre/trainyard/releases/latest/download/yard_$(uname -s)_$(uname -m).tar.gz | tar -xz
sudo mv yard /usr/local/bin/
```

### 2. Set up your server

SSH into your server (Ubuntu/Debian, ≥2 GB RAM) and run:

```bash
sudo yard setup
```

The wizard installs k3s, Helm, Nginx Ingress, cert-manager, and a wildcard TLS certificate. At the end it prints a `KUBE_CONFIG` secret to add to GitHub.

### 3. Add the secret to GitHub

In your repo: **Settings → Secrets and variables → Actions → New repository secret**

| Name | Value |
|---|---|
| `KUBE_CONFIG` | *(output from `yard setup`)* |

### 4. Scaffold your repo

In the root of the repo you want previews for:

```bash
yard init
```

This generates `.github/pr-preview.yml` and `.github/workflows/preview.yml`.

### 5. Open a PR and add the label

Add the `preview` label to any PR. Trainyard will:
1. Build and push your Docker image
2. Deploy it to a fresh namespace
3. Post a comment with the preview URL

Remove the label or close the PR to tear it down.

---

## Configuration Reference

### `.github/pr-preview.yml`

```yaml
app:
  domain: preview.yourdomain.com   # wildcard domain configured on your server

ingress:
  tls: true                        # enable HTTPS (requires wildcard cert)
  class: nginx                     # ingress class name

label: preview                     # GitHub label that triggers deployments

services:
  - name: app
    build:
      context: .
      dockerfile: Dockerfile
    port: 3000
    public: true                   # expose via ingress
    subdomain: "pr-{number}"       # {number} is replaced with the PR number
    replicas: 1
    resources:
      limits:
        cpu: "250m"
        memory: "128Mi"
    env:
      - name: PORT
        value: "3000"
    dependsOn: []                  # wait for these services before starting
```

### Multi-service example (frontend + backend + database)

```yaml
app:
  domain: preview.yourdomain.com

ingress:
  tls: true
  class: nginx

label: preview

services:
  - name: frontend
    build:
      context: .
      dockerfile: frontend/Dockerfile
    port: 3000
    public: true
    subdomain: "pr-{number}"
    replicas: 1
    resources:
      limits:
        cpu: "250m"
        memory: "128Mi"
    env:
      - name: API_URL
        value: "http://backend:8080"
    dependsOn:
      - backend

  - name: backend
    build:
      context: .
      dockerfile: backend/Dockerfile
    port: 8080
    public: false
    replicas: 1
    resources:
      limits:
        cpu: "250m"
        memory: "256Mi"
    env:
      - name: DATABASE_URL
        value: "postgres://postgres:postgres@db:5432/app"
    dependsOn:
      - db

  - name: db
    image: postgres:16-alpine
    port: 5432
    public: false
    replicas: 1
    resources:
      limits:
        cpu: "250m"
        memory: "256Mi"
    env:
      - name: POSTGRES_USER
        value: "postgres"
      - name: POSTGRES_PASSWORD
        value: "postgres"
      - name: POSTGRES_DB
        value: "app"
    dependsOn: []
```

### Stale environment cleanup

Install the cleanup chart to automatically delete preview environments older than N days:

```bash
helm upgrade --install trainyard-cleanup \
  oci://ghcr.io/emilvorre/trainyard/charts/cleanup \
  --namespace trainyard-system \
  --create-namespace \
  --set maxAgeDays=7
```

---

## How It Works

```
PR opened + label added
        │
        ▼
  GitHub Actions (deploy.yml)
        │
        ├─ docker build + push → ghcr.io
        ├─ kubectl create namespace preview-pr-{n}
        ├─ helm upgrade --install pr-{n} oci://…/pr-preview
        └─ posts PR comment with URL
        
PR closed / label removed
        │
        ▼
  GitHub Actions (teardown.yml)
        │
        └─ helm uninstall + kubectl delete namespace

Every hour (CronJob)
        │
        ▼
  trainyard-cleanup
        └─ deletes namespaces older than maxAgeDays
```

Each preview environment is an isolated Kubernetes namespace. Services within the same environment communicate over internal DNS (`http://service-name:port`). Only services marked `public: true` get an Ingress.

---

## Comparison

| | Trainyard | Vercel | Render | Argo CD + custom |
|---|---|---|---|---|
| Self-hosted | ✅ | ❌ | ❌ | ✅ |
| Multi-service | ✅ | ❌ | ✅ | ✅ |
| Cost | VPS only | Per seat | Per seat | Complex |
| Setup time | ~10 min | ~2 min | ~5 min | Days |
| Lock-in | None | High | Medium | Low |
| Database previews | ✅ | ❌ | ✅ | ✅ |

---

## Server Requirements

| | Minimum | Recommended |
|---|---|---|
| OS | Ubuntu 22.04 / Debian 12 | Ubuntu 24.04 |
| RAM | 2 GB | 4 GB |
| CPU | 1 vCPU | 2 vCPU |
| Disk | 20 GB | 40 GB |
| Network | Public IP | Public IP |

Tested on Hetzner CX22 (€4/month).

---

## CLI Reference

```
yard setup     Run on your server — installs k3s, Helm, Nginx, cert-manager,
               wildcard TLS cert, and outputs your KUBE_CONFIG secret.

yard init      Run in a consuming repo — scaffolds .github/pr-preview.yml
               and .github/workflows/preview.yml.

yard validate  Validates a pr-preview.yml config file.
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Security

See [SECURITY.md](SECURITY.md).

## License

MIT — see [LICENSE](LICENSE).
