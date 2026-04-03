package setup

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/Emilvorre/trainyard/internal/tui"
)

func outputKubeconfig(serverIP string) error {
	raw, err := os.ReadFile("/etc/rancher/k3s/k3s.yaml")
	if err != nil {
		return fmt.Errorf("reading /etc/rancher/k3s/k3s.yaml: %w", err)
	}

	// Patch the server address from 127.0.0.1 → public IP
	patched := strings.ReplaceAll(string(raw), "127.0.0.1", serverIP)
	patched = strings.ReplaceAll(patched, "localhost", serverIP)

	encoded := base64.StdEncoding.EncodeToString([]byte(patched))

	tui.Box("GitHub Actions Secret", []string{
		"Add this secret to your repo (Settings → Secrets → Actions):",
		"",
		"  Name:  KUBE_CONFIG",
		"  Value: (printed below)",
	})

	fmt.Println(encoded)
	fmt.Println()

	// Also write to a file for convenience
	outPath := "/root/trainyard-kubeconfig.b64"
	if err := os.WriteFile(outPath, []byte(encoded+"\n"), 0600); err != nil {
		tui.Warn("Could not write kubeconfig to %s: %v", outPath, err)
	} else {
		tui.Info("Also saved to %s", outPath)
	}

	return nil
}