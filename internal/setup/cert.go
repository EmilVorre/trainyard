package setup
 
import (
	"fmt"
	"os"
	"strings"

	"github.com/Emilvorre/trainyard/internal/tui"
)
 
type dnsProvider struct {
	Label   string
	Hook    string
	EnvVars []dnsEnvVar
	DocsURL string
}
 
type dnsEnvVar struct {
	Key    string
	Label  string
	Secret bool
}
 
var supportedProviders = []dnsProvider{
	{
		Label:   "Cloudflare",
		Hook:    "dns_cf",
		EnvVars: []dnsEnvVar{{Key: "CF_Token", Label: "API Token", Secret: true}},
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_cf",
	},
	{
		Label:   "name.com",
		Hook:    "dns_namecom",
		EnvVars: []dnsEnvVar{{Key: "Namecom_Username", Label: "username"}, {Key: "Namecom_Token", Label: "API token", Secret: true}},
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_namecom",
	},
	{
		Label:   "AWS Route 53",
		Hook:    "dns_aws",
		EnvVars: []dnsEnvVar{{Key: "AWS_ACCESS_KEY_ID", Label: "Access Key ID"}, {Key: "AWS_SECRET_ACCESS_KEY", Label: "Secret Access Key", Secret: true}},
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_aws",
	},
	{
		Label:   "Hetzner DNS",
		Hook:    "dns_hetzner",
		EnvVars: []dnsEnvVar{{Key: "HETZNER_Token", Label: "DNS API token", Secret: true}},
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_hetzner",
	},
	{
		Label:   "DigitalOcean",
		Hook:    "dns_dgon",
		EnvVars: []dnsEnvVar{{Key: "DO_API_KEY", Label: "API key", Secret: true}},
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_dgon",
	},
	{
		Label:   "Porkbun",
		Hook:    "dns_porkbun",
		EnvVars: []dnsEnvVar{{Key: "PORKBUN_API_KEY", Label: "API key", Secret: true}, {Key: "PORKBUN_SECRET_API_KEY", Label: "secret API key", Secret: true}},
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_porkbun",
	},
	{
		Label:   "Namecheap",
		Hook:    "dns_namecheap",
		EnvVars: []dnsEnvVar{{Key: "NAMECHEAP_USERNAME", Label: "username"}, {Key: "NAMECHEAP_API_KEY", Label: "API key", Secret: true}},
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_namecheap",
	},
	{
		Label:   "GoDaddy",
		Hook:    "dns_gd",
		EnvVars: []dnsEnvVar{{Key: "GD_Key", Label: "API key", Secret: true}, {Key: "GD_Secret", Label: "API secret", Secret: true}},
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi#dns_gd",
	},
	{
		Label:   "Manual (any provider)",
		Hook:    "manual",
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/DNS-manual-mode",
	},
	{
		Label:   "Other (acme.sh DNS API)",
		Hook:    "other",
		DocsURL: "https://github.com/acmesh-official/acme.sh/wiki/dnsapi",
	},
}
 
type certConfig struct {
	Domain     string
	Email      string
	Provider   dnsProvider
	EnvVals    map[string]string
	CustomHook string
}
 
func pickDNSProvider(domain, email string) certConfig {
	cfg := certConfig{
		Domain:  domain,
		Email:   email,
		EnvVals: map[string]string{},
	}
 
	fmt.Println()
	fmt.Println("  DNS provider for wildcard certificate (DNS-01 challenge):")
	fmt.Println()
	for i, p := range supportedProviders {
		fmt.Printf("  [%d] %s\n", i+1, p.Label)
	}
	fmt.Println()
 
	var chosen dnsProvider
	for {
		choice := tui.Prompt("Choose provider", "1")
		n := 0
		fmt.Sscanf(strings.TrimSpace(choice), "%d", &n)
		if n >= 1 && n <= len(supportedProviders) {
			chosen = supportedProviders[n-1]
			break
		}
		tui.Warn("Enter a number between 1 and %d", len(supportedProviders))
	}
 
	cfg.Provider = chosen
 
	switch chosen.Hook {
	case "manual":
		fmt.Println()
		tui.Info("You will be prompted to add TXT records to your DNS manually.")
 
	case "other":
		fmt.Println()
		tui.Info("Full provider list: %s", chosen.DocsURL)
		cfg.CustomHook = tui.Prompt("acme.sh DNS hook name (e.g. dns_cf)", "")
		fmt.Println()
		fmt.Println("  Enter the environment variables required by this hook.")
		fmt.Println("  Press Enter with an empty key name to finish.")
		fmt.Println()
		for {
			key := tui.Prompt("Env var name", "")
			if key == "" {
				break
			}
			val := tui.Prompt(fmt.Sprintf("Value for %s", key), "")
			cfg.EnvVals[key] = val
		}
 
	default:
		if len(chosen.EnvVars) > 0 {
			fmt.Println()
			tui.Info("Credentials for %s", chosen.Label)
			tui.Info("Docs: %s", chosen.DocsURL)
			fmt.Println()
			for _, ev := range chosen.EnvVars {
				val := tui.Prompt(ev.Label, "")
				cfg.EnvVals[ev.Key] = val
			}
		}
	}
 
	return cfg
}
 
func installAcmeSh(cfg certConfig) error {
	if !commandExists("/root/.acme.sh/acme.sh") {
		if err := run("sh", "-c",
			fmt.Sprintf(`curl -fsSL https://get.acme.sh | sh -s email=%s`, cfg.Email),
		); err != nil {
			return fmt.Errorf("installing acme.sh: %w", err)
		}
	}
 
	acme := "/root/.acme.sh/acme.sh"
	_ = runSilent(acme, "--register-account", "-m", cfg.Email, "--server", "letsencrypt")
 
	for k, v := range cfg.EnvVals {
		os.Setenv(k, v)
	}
 
	hook := cfg.Provider.Hook
	if hook == "other" {
		hook = cfg.CustomHook
	}
 
	if hook == "manual" {
		return issueCertManual(acme, cfg)
	}
	return issueCertDNSAPI(acme, cfg, hook)
}
 
func issueCertDNSAPI(acme string, cfg certConfig, hook string) error {
	return run(acme,
		"--issue",
		"--dns", hook,
		"-d", cfg.Domain,
		"-d", "*."+cfg.Domain,
		"--server", "letsencrypt",
		"--keylength", "ec-256",
	)
}
 
func issueCertManual(acme string, cfg certConfig) error {
	fmt.Println()
	fmt.Println("  acme.sh will print TXT records to add to your DNS.")
	fmt.Println("  Add them, wait for propagation, then press Enter.")
	fmt.Println()
	return run(acme,
		"--issue",
		"--dns",
		"-d", cfg.Domain,
		"-d", "*."+cfg.Domain,
		"--server", "letsencrypt",
		"--keylength", "ec-256",
		"--yes-I-know-dns-manual-mode-enough-go-ahead-please",
	)
}
 
func installCertToNginx(domain string) error {
	acme := "/root/.acme.sh/acme.sh"
	targetDir := fmt.Sprintf("/etc/letsencrypt/live/%s", domain)
 
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}
 
	return run(acme,
		"--install-cert",
		"-d", domain,
		"-d", "*."+domain,
		"--ecc",
		"--cert-file", targetDir+"/cert.pem",
		"--key-file", targetDir+"/privkey.pem",
		"--fullchain-file", targetDir+"/fullchain.pem",
		"--reloadcmd", "systemctl reload nginx",
	)
}
 
func certExists(domain string) bool {
	_, err := os.Stat(fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem", domain))
	return err == nil
}
 
func sanitizeDomain(d string) string {
	return strings.TrimPrefix(d, "*.")
}
 