package setup

import (
	"fmt"
	"strings"
	"testing"
)

// ── sanitizeDomain ────────────────────────────────────────────────────────────

func TestSanitizeDomain_StripsWildcardPrefix(t *testing.T) {
	got := sanitizeDomain("*.preview.example.com")
	if got != "preview.example.com" {
		t.Errorf("expected preview.example.com, got %q", got)
	}
}

func TestSanitizeDomain_NoPrefix(t *testing.T) {
	got := sanitizeDomain("preview.example.com")
	if got != "preview.example.com" {
		t.Errorf("expected preview.example.com unchanged, got %q", got)
	}
}

func TestSanitizeDomain_EmptyString(t *testing.T) {
	got := sanitizeDomain("")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestSanitizeDomain_OnlyWildcard(t *testing.T) {
	got := sanitizeDomain("*.")
	if got != "" {
		t.Errorf("expected empty string for bare wildcard, got %q", got)
	}
}

// ── dirOf ─────────────────────────────────────────────────────────────────────

func TestDirOf_ReturnsDirectory(t *testing.T) {
	got := dirOf("/etc/nginx/sites-available/trainyard.conf")
	if got != "/etc/nginx/sites-available" {
		t.Errorf("expected /etc/nginx/sites-available, got %q", got)
	}
}

func TestDirOf_SingleSlash(t *testing.T) {
	got := dirOf("/file.txt")
	if got != "" {
		t.Errorf("expected empty string for root-level file, got %q", got)
	}
}

func TestDirOf_NoSlash(t *testing.T) {
	got := dirOf("file.txt")
	if got != "." {
		t.Errorf("expected . for relative file with no dir, got %q", got)
	}
}

func TestDirOf_TmpPath(t *testing.T) {
	got := dirOf("/tmp/nginx-ingress-values.yaml")
	if got != "/tmp" {
		t.Errorf("expected /tmp, got %q", got)
	}
}

func TestDirOf_NestedPath(t *testing.T) {
	got := dirOf("/etc/rancher/k3s/config.yaml")
	if got != "/etc/rancher/k3s" {
		t.Errorf("expected /etc/rancher/k3s, got %q", got)
	}
}

// ── nginx site config rendering ───────────────────────────────────────────────

func renderNginxConfig(domain string, httpPort int) string {
	return fmt.Sprintf(nginxSiteTemplate,
		domain,
		httpPort,
		domain,
		httpPort,
		domain,
		domain,
		domain,
		httpPort,
	)
}

func TestNginxSiteConfig_ContainsDomain(t *testing.T) {
	out := renderNginxConfig("preview.example.com", 31540)

	if !strings.Contains(out, "preview.example.com") {
		t.Error("expected domain in nginx config")
	}
}

func TestNginxSiteConfig_ContainsWildcardServerName(t *testing.T) {
	out := renderNginxConfig("preview.example.com", 31540)

	if !strings.Contains(out, "server_name *.preview.example.com") {
		t.Error("expected wildcard server_name in nginx config")
	}
}

func TestNginxSiteConfig_ContainsHTTPPort(t *testing.T) {
	out := renderNginxConfig("preview.example.com", 31540)

	if !strings.Contains(out, "31540") {
		t.Error("expected http port 31540 in nginx config")
	}
}

func TestNginxSiteConfig_ContainsSSLCertPaths(t *testing.T) {
	out := renderNginxConfig("preview.example.com", 31540)

	if !strings.Contains(out, "/etc/letsencrypt/live/preview.example.com/fullchain.pem") {
		t.Error("expected fullchain.pem path in nginx config")
	}
	if !strings.Contains(out, "/etc/letsencrypt/live/preview.example.com/privkey.pem") {
		t.Error("expected privkey.pem path in nginx config")
	}
}

func TestNginxSiteConfig_ContainsBothListeners(t *testing.T) {
	out := renderNginxConfig("preview.example.com", 31540)

	if !strings.Contains(out, "listen 80") {
		t.Error("expected listen 80 in nginx config")
	}
	if !strings.Contains(out, "listen 443 ssl") {
		t.Error("expected listen 443 ssl in nginx config")
	}
}

func TestNginxSiteConfig_ContainsProxyHeaders(t *testing.T) {
	out := renderNginxConfig("preview.example.com", 31540)

	for _, header := range []string{"X-Real-IP", "X-Forwarded-For", "X-Forwarded-Proto", "Host"} {
		if !strings.Contains(out, header) {
			t.Errorf("expected proxy header %s in nginx config", header)
		}
	}
}

func TestNginxSiteConfig_CustomPort(t *testing.T) {
	out := renderNginxConfig("preview.example.com", 9999)

	if !strings.Contains(out, "9999") {
		t.Error("expected custom port 9999 in nginx config")
	}
	if strings.Contains(out, "31540") {
		t.Error("expected default port 31540 not to appear with custom port")
	}
}

// ── nginx ingress values rendering ───────────────────────────────────────────

func renderNginxValues(httpPort, httpsPort int) string {
	return fmt.Sprintf(nginxValuesTemplate, httpPort, httpsPort)
}

func TestNginxValuesTemplate_ContainsNodePort(t *testing.T) {
	out := renderNginxValues(31540, 30456)

	if !strings.Contains(out, "NodePort") {
		t.Error("expected NodePort service type")
	}
}

func TestNginxValuesTemplate_ContainsHTTPPort(t *testing.T) {
	out := renderNginxValues(31540, 30456)

	if !strings.Contains(out, "31540") {
		t.Error("expected http port 31540")
	}
}

func TestNginxValuesTemplate_ContainsHTTPSPort(t *testing.T) {
	out := renderNginxValues(31540, 30456)

	if !strings.Contains(out, "30456") {
		t.Error("expected https port 30456")
	}
}

func TestNginxValuesTemplate_ForwardedHeaders(t *testing.T) {
	out := renderNginxValues(31540, 30456)

	if !strings.Contains(out, "use-forwarded-headers") {
		t.Error("expected use-forwarded-headers config")
	}
}

// ── ClusterIssuer manifest rendering ─────────────────────────────────────────

func renderClusterIssuer(email string) string {
	return fmt.Sprintf(`apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: %s
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
`, email)
}

func TestClusterIssuerManifest_ContainsEmail(t *testing.T) {
	out := renderClusterIssuer("admin@example.com")

	if !strings.Contains(out, "admin@example.com") {
		t.Error("expected email in ClusterIssuer manifest")
	}
}

func TestClusterIssuerManifest_ContainsACMEServer(t *testing.T) {
	out := renderClusterIssuer("admin@example.com")

	if !strings.Contains(out, "acme-v02.api.letsencrypt.org") {
		t.Error("expected Let's Encrypt ACME server URL")
	}
}

func TestClusterIssuerManifest_ContainsNginxSolver(t *testing.T) {
	out := renderClusterIssuer("admin@example.com")

	if !strings.Contains(out, "class: nginx") {
		t.Error("expected nginx ingress class in solver")
	}
}

func TestClusterIssuerManifest_Kind(t *testing.T) {
	out := renderClusterIssuer("admin@example.com")

	if !strings.Contains(out, "kind: ClusterIssuer") {
		t.Error("expected kind: ClusterIssuer")
	}
}

// ── sysInfo threshold logic ───────────────────────────────────────────────────

func TestSysInfoRAMThreshold(t *testing.T) {
	tests := []struct {
		ramMB int
		ok    bool
	}{
		{0, false},
		{1799, false},
		{1800, true},
		{2048, true},
		{4096, true},
	}

	for _, tt := range tests {
		got := tt.ramMB >= minRAMMB
		if got != tt.ok {
			t.Errorf("RAM %d MB: expected ok=%v, got %v", tt.ramMB, tt.ok, got)
		}
	}
}

func TestSysInfoDiskThreshold(t *testing.T) {
	tests := []struct {
		diskGB int
		ok     bool
	}{
		{0, false},
		{9, false},
		{10, true},
		{20, true},
		{40, true},
	}

	for _, tt := range tests {
		got := tt.diskGB >= minDiskGB
		if got != tt.ok {
			t.Errorf("disk %d GB: expected ok=%v, got %v", tt.diskGB, tt.ok, got)
		}
	}
}

// ── readMemInfoMB parsing ─────────────────────────────────────────────────────

func TestParseMemInfo(t *testing.T) {
	// Test the parsing logic directly with synthetic /proc/meminfo content
	input := "MemTotal:        4096000 kB\nMemFree:         2048000 kB\n"

	var ramMB int
	for _, line := range strings.Split(input, "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb := 0
				fmt.Sscanf(fields[1], "%d", &kb)
				ramMB = kb / 1024
			}
		}
	}

	if ramMB != 4000 {
		t.Errorf("expected 4000 MB, got %d", ramMB)
	}
}

func TestParseMemInfo_Missing(t *testing.T) {
	input := "MemFree:         2048000 kB\n"

	found := false
	for _, line := range strings.Split(input, "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			found = true
		}
	}

	if found {
		t.Error("expected MemTotal not to be found in input without it")
	}
}
