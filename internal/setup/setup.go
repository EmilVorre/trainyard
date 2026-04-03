package setup
 
import (
	"fmt"
	"os"
 
	"github.com/Emilvorre/trainyard/internal/tui"
	"github.com/spf13/cobra"
)
 
// Command returns the cobra command for `yard setup`.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Set up a server to host PR preview environments",
		Long: `Interactive wizard that installs and configures everything needed
to run Trainyard PR preview environments on this server.
 
Run this command directly on the target server as root.
 
Installs:
  • k3s (Kubernetes)
  • Helm
  • Nginx Ingress Controller (NodePort)
  • cert-manager + ClusterIssuer
  • System nginx wildcard proxy config
  • acme.sh wildcard TLS certificate
  
Outputs a base64-encoded KUBE_CONFIG secret for GitHub Actions.`,
		RunE: runSetup,
	}
 
	// Allow skipping individual steps (useful for re-runs)
	cmd.Flags().Bool("skip-k3s", false, "Skip k3s installation")
	cmd.Flags().Bool("skip-helm", false, "Skip Helm installation")
	cmd.Flags().Bool("skip-ingress", false, "Skip Nginx Ingress installation")
	cmd.Flags().Bool("skip-certmanager", false, "Skip cert-manager installation")
	cmd.Flags().Bool("skip-nginx", false, "Skip system nginx config")
	cmd.Flags().Bool("skip-cert", false, "Skip TLS certificate issuance")
 
	return cmd
}
 
func runSetup(cmd *cobra.Command, _ []string) error {
	skipK3s, _ := cmd.Flags().GetBool("skip-k3s")
	skipHelm, _ := cmd.Flags().GetBool("skip-helm")
	skipIngress, _ := cmd.Flags().GetBool("skip-ingress")
	skipCertManager, _ := cmd.Flags().GetBool("skip-certmanager")
	skipNginx, _ := cmd.Flags().GetBool("skip-nginx")
	skipCert, _ := cmd.Flags().GetBool("skip-cert")
 
	tui.Banner()
 
	// ── Step 0: System check ─────────────────────────────────────────────────
	sysInfo, err := checkSystem()
	if err != nil {
		tui.Fatal("System check failed: %v", err)
	}
	printSystemSummary(sysInfo)
 
	if !tui.Confirm("Proceed with setup?") {
		fmt.Println("  Aborted.")
		os.Exit(0)
	}
 
	// ── Step 1: Gather config ────────────────────────────────────────────────
	tui.Section("Configuration")
 
	serverIP := tui.Prompt("Server public IP", detectPublicIP())
	domain := tui.Prompt("Wildcard domain (e.g. preview.vorre.dev)", "")
	if domain == "" {
		tui.Fatal("Domain is required")
	}
	domain = sanitizeDomain(domain) // strip leading *.
 
	email := tui.Prompt("Email for Let's Encrypt / acme.sh", "")
	if email == "" {
		tui.Fatal("Email is required")
	}
 
	httpPort := 31540
	httpsPort := 30456
 
	cCfg := pickDNSProvider(domain, email)
 
	fmt.Println()
	tui.Info("Domain:    *.%s", domain)
	tui.Info("Server IP: %s", serverIP)
	tui.Info("HTTP port: %d (k3s NodePort)", httpPort)
	tui.Info("Email:     %s", email)
	fmt.Println()
 
	if !tui.Confirm("Start installation?") {
		fmt.Println("  Aborted.")
		os.Exit(0)
	}
 
	// ── Step 2: k3s ──────────────────────────────────────────────────────────
	tui.Section("k3s")
	if skipK3s || k3sRunning() {
		tui.StepSkip("k3s")
	} else {
		if err := tui.Spinner("Installing k3s", installK3s); err != nil {
			tui.Fatal("k3s installation failed: %v", err)
		}
	}
 
	// ── Step 3: Helm ─────────────────────────────────────────────────────────
	tui.Section("Helm")
	if skipHelm || commandExists("helm") {
		tui.StepSkip("Helm")
	} else {
		if err := tui.Spinner("Installing Helm", installHelm); err != nil {
			tui.Fatal("Helm installation failed: %v", err)
		}
	}
 
	// ── Step 4: Nginx Ingress Controller ─────────────────────────────────────
	tui.Section("Nginx Ingress Controller")
	if skipIngress {
		tui.StepSkip("Nginx Ingress Controller")
	} else {
		if err := tui.Spinner("Installing Nginx Ingress Controller", func() error {
			return installNginxIngress(httpPort, httpsPort)
		}); err != nil {
			tui.Fatal("Nginx Ingress installation failed: %v", err)
		}
	}
 
	// ── Step 5: cert-manager ─────────────────────────────────────────────────
	tui.Section("cert-manager")
	if skipCertManager {
		tui.StepSkip("cert-manager")
	} else {
		if err := tui.Spinner("Installing cert-manager", installCertManager); err != nil {
			tui.Fatal("cert-manager installation failed: %v", err)
		}
		if err := tui.Step("Applying ClusterIssuer", func() error {
			return applyClusterIssuer(email)
		}); err != nil {
			tui.Fatal("ClusterIssuer apply failed: %v", err)
		}
	}
 
	// ── Step 6: System nginx config ──────────────────────────────────────────
	tui.Section("System nginx")
	if skipNginx {
		tui.StepSkip("Nginx site config")
	} else {
		if err := tui.Step("Writing nginx site config", func() error {
			return writeNginxSiteConfig(domain, httpPort)
		}); err != nil {
			tui.Fatal("Nginx config failed: %v", err)
		}
	}
 
	// ── Step 7: TLS certificate ──────────────────────────────────────────────
	tui.Section("TLS Certificate (acme.sh)")
	if skipCert {
		tui.StepSkip("TLS certificate")
	} else if certExists(domain) {
		tui.StepSkip(fmt.Sprintf("Certificate for *.%s", domain))
	} else {
		tui.Info("Issuing wildcard certificate for *.%s", domain)
		if err := installAcmeSh(cCfg); err != nil {
			tui.Fatal("Certificate issuance failed: %v", err)
		}
		if err := tui.Step("Installing cert to nginx paths", func() error {
			return installCertToNginx(domain)
		}); err != nil {
			tui.Fatal("Cert install failed: %v", err)
		}
	}
 
	// ── Step 8: Kubeconfig output ────────────────────────────────────────────
	tui.Section("GitHub Actions Secret")
	if err := outputKubeconfig(serverIP); err != nil {
		tui.Fatal("Could not output kubeconfig: %v", err)
	}
 
	// ── Done ─────────────────────────────────────────────────────────────────
	fmt.Println()
	tui.Success("Setup complete! Your server is ready to host PR preview environments.")
	fmt.Println()
	tui.Info("Next steps:")
	fmt.Printf("    1. Add KUBE_CONFIG secret to your GitHub repo\n")
	fmt.Printf("    2. Run `yard init` in each consuming repo\n")
	fmt.Printf("    3. Open a PR and add the `preview` label\n")
	fmt.Println()
 
	return nil
}
 
// detectPublicIP tries to auto-detect the server's public IP.
func detectPublicIP() string {
	ip, err := runOutput("curl", "-sf", "https://ifconfig.me")
	if err != nil {
		return ""
	}
	return ip
}
 