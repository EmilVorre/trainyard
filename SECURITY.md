# Security Policy

## Supported Versions

| Version | Supported |
|---|---|
| Latest release | ✅ |
| Older releases | ❌ |

We only maintain the latest release. Please upgrade before reporting issues.

## Reporting a Vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Report security issues by emailing: **emil@vorre.dev**

Please include:
- A description of the vulnerability
- Steps to reproduce
- Potential impact
- Any suggested fixes (optional)

You can expect an acknowledgement within 48 hours and a resolution timeline within 7 days for confirmed issues.

## Scope

The following are in scope:

- The `yard` CLI binary
- The `pr-preview` Helm chart
- The `cleanup` Helm chart and its container image
- The reusable GitHub Actions workflows

The following are **out of scope**:

- The underlying k3s / Kubernetes installation
- Third-party dependencies (report these upstream)
- Issues in consuming repos using Trainyard

## Security Considerations for Self-Hosted Deployments

- The `KUBE_CONFIG` secret grants full cluster access — treat it like a root password
- Preview environments run in isolated namespaces but share the same cluster node
- Do not run untrusted code in preview environments on shared infrastructure
- Rotate `KUBE_CONFIG` if you suspect it has been compromised (`yard setup` can regenerate it)
- The cleanup CronJob has namespace list/delete permissions cluster-wide — review the RBAC in `charts/cleanup/templates/rbac.yaml`
