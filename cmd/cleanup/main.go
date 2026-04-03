package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	// Config from env (injected by Helm chart)
	maxAgeDays := envInt("MAX_AGE_DAYS", 7)
	namespaceLabel := envStr("NAMESPACE_LABEL", "trainyard.dev/preview=true")
	deployedAtAnnotation := envStr("DEPLOYED_AT_ANNOTATION", "trainyard.dev/deployed-at")
	logLevel := envStr("LOG_LEVEL", "info")

	// Logger
	level := slog.LevelInfo
	if logLevel == "debug" {
		level = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	logger.Info("trainyard cleanup starting",
		"maxAgeDays", maxAgeDays,
		"namespaceLabel", namespaceLabel,
	)

	// In-cluster Kubernetes client
	cfg, err := rest.InClusterConfig()
	if err != nil {
		logger.Error("failed to get in-cluster config", "err", err)
		os.Exit(1)
	}

	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Error("failed to create kubernetes client", "err", err)
		os.Exit(1)
	}

	ctx := context.Background()
	cutoff := time.Now().Add(-time.Duration(maxAgeDays) * 24 * time.Hour)

	// List namespaces with the preview label
	nsList, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{
		LabelSelector: namespaceLabel,
	})
	if err != nil {
		logger.Error("failed to list namespaces", "err", err)
		os.Exit(1)
	}

	logger.Info("found preview namespaces", "count", len(nsList.Items))

	deleted := 0
	skipped := 0
	errors := 0

	for _, ns := range nsList.Items {
		name := ns.Name

		// Prefer our explicit annotation over namespace creation time.
		// The annotation is stamped by the deploy workflow on every deploy,
		// so it reflects the last actual deployment, not just namespace age.
		var deployedAt time.Time

		if annotationVal, ok := ns.Annotations[deployedAtAnnotation]; ok {
			t, err := time.Parse(time.RFC3339, annotationVal)
			if err != nil {
				logger.Warn("could not parse deployed-at annotation, falling back to creation time",
					"namespace", name,
					"annotation", annotationVal,
					"err", err,
				)
				deployedAt = ns.CreationTimestamp.Time
			} else {
				deployedAt = t
			}
		} else {
			// No annotation — fall back to namespace creation timestamp.
			// This covers namespaces created before the annotation was introduced.
			logger.Debug("no deployed-at annotation, using creation timestamp",
				"namespace", name,
			)
			deployedAt = ns.CreationTimestamp.Time
		}

		age := time.Since(deployedAt)
		logger.Debug("checking namespace",
			"namespace", name,
			"deployedAt", deployedAt.Format(time.RFC3339),
			"ageDays", fmt.Sprintf("%.1f", age.Hours()/24),
		)

		if deployedAt.Before(cutoff) {
			logger.Info("deleting stale namespace",
				"namespace", name,
				"deployedAt", deployedAt.Format(time.RFC3339),
				"ageDays", fmt.Sprintf("%.1f", age.Hours()/24),
				"maxAgeDays", maxAgeDays,
			)

			err := client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
			if err != nil {
				logger.Error("failed to delete namespace",
					"namespace", name,
					"err", err,
				)
				errors++
				continue
			}

			deleted++
		} else {
			logger.Info("namespace is fresh, skipping",
				"namespace", name,
				"ageDays", fmt.Sprintf("%.1f", age.Hours()/24),
			)
			skipped++
		}
	}

	logger.Info("cleanup complete",
		"deleted", deleted,
		"skipped", skipped,
		"errors", errors,
	)

	if errors > 0 {
		os.Exit(1)
	}
}

func envStr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func envInt(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}