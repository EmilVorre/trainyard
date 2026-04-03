package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
)

type OutputFile struct {
	Path    string
	Content string
}

func WriteFiles(files []OutputFile, force bool) ([]string, error) {
	var written []string

	for _, f := range files {
		dir := filepath.Dir(f.Path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return written, fmt.Errorf("creating directory %s: %w", dir, err)
		}

		if _, err := os.Stat(f.Path); err == nil && !force {
			fmt.Printf("\n  File already exists: %s\n", f.Path)
			fmt.Print("  Overwrite? [y/N]: ")
			var answer string
			fmt.Scanln(&answer)
			if answer != "y" && answer != "Y" {
				fmt.Printf("  Skipped %s\n", f.Path)
				continue
			}
		}

		if err := os.WriteFile(f.Path, []byte(f.Content), 0644); err != nil {
			return written, fmt.Errorf("writing %s: %w", f.Path, err)
		}

		written = append(written, f.Path)
	}

	return written, nil
}
