package setup
 
import "fmt"
 
// ── Helm ──────────────────────────────────────────────────────────────────────
 
func installHelm() error {
	return run("sh", "-c",
		`curl -fsSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash`,
	)
}
 
// ── Nginx Ingress Controller ──────────────────────────────────────────────────
 
const nginxValuesTemplate = `controller:
  service:
    type: NodePort
    nodePorts:
      http: "%d"
      https: "%d"
  config:
    use-forwarded-headers: "true"
    compute-full-forwarded-for: "true"
`
 
func installNginxIngress(httpPort, httpsPort int) error {
	values := fmt.Sprintf(nginxValuesTemplate, httpPort, httpsPort)
 
	valuesPath := "/tmp/nginx-ingress-values.yaml"
	if err := writeFile(valuesPath, values, 0644); err != nil {
		return err
	}
 
	if err := runSilent("helm", "repo", "add", "ingress-nginx",
		"https://kubernetes.github.io/ingress-nginx"); err != nil {
		// Already added is fine
		_ = err
	}
	if err := runSilent("helm", "repo", "update"); err != nil {
		return err
	}
 
	return run("helm", "upgrade", "--install",
		"ingress-nginx", "ingress-nginx/ingress-nginx",
		"--namespace", "ingress-nginx",
		"--create-namespace",
		"--values", valuesPath,
		"--wait",
	)
}
 
// ── cert-manager ──────────────────────────────────────────────────────────────
 
func installCertManager() error {
	if err := runSilent("helm", "repo", "add", "jetstack",
		"https://charts.jetstack.io"); err != nil {
		_ = err
	}
	if err := runSilent("helm", "repo", "update"); err != nil {
		return err
	}
	return run("helm", "upgrade", "--install",
		"cert-manager", "jetstack/cert-manager",
		"--namespace", "cert-manager",
		"--create-namespace",
		"--set", "installCRDs=true",
		"--wait",
	)
}
 
// applyClusterIssuer writes a self-signed or ACME ClusterIssuer.
// For setup purposes we apply a self-signed one; the user can update it later.
func applyClusterIssuer(email string) error {
	manifest := fmt.Sprintf(`apiVersion: cert-manager.io/v1
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
 
	if err := writeFile("/tmp/cluster-issuer.yaml", manifest, 0644); err != nil {
		return err
	}
	return runSilent("kubectl", "apply", "-f", "/tmp/cluster-issuer.yaml")
}