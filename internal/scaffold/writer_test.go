package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFiles_WritesContent(t *testing.T) {
	dir := t.TempDir()

	files := []OutputFile{
		{
			Path:    filepath.Join(dir, "test.yml"),
			Content: "hello: world\n",
		},
	}

	written, err := WriteFiles(files, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 written file, got %d", len(written))
	}

	content, err := os.ReadFile(written[0])
	if err != nil {
		t.Fatalf("could not read written file: %v", err)
	}
	if string(content) != "hello: world\n" {
		t.Errorf("unexpected file content: %s", string(content))
	}
}

func TestWriteFiles_CreatesIntermediateDirectories(t *testing.T) {
	dir := t.TempDir()

	files := []OutputFile{
		{
			Path:    filepath.Join(dir, ".github", "workflows", "preview.yml"),
			Content: "name: PR Preview\n",
		},
	}

	written, err := WriteFiles(files, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 written file, got %d", len(written))
	}

	if _, err := os.Stat(written[0]); os.IsNotExist(err) {
		t.Error("expected file to exist after writing")
	}
}

func TestWriteFiles_MultipleFiles(t *testing.T) {
	dir := t.TempDir()

	files := []OutputFile{
		{Path: filepath.Join(dir, "a.yml"), Content: "a: 1\n"},
		{Path: filepath.Join(dir, "b.yml"), Content: "b: 2\n"},
	}

	written, err := WriteFiles(files, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(written) != 2 {
		t.Fatalf("expected 2 written files, got %d", len(written))
	}
}

func TestWriteFiles_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "existing.yml")

	if err := os.WriteFile(path, []byte("original\n"), 0644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	files := []OutputFile{
		{Path: path, Content: "overwritten\n"},
	}

	written, err := WriteFiles(files, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(written) != 1 {
		t.Fatalf("expected 1 written file, got %d", len(written))
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(content) != "overwritten\n" {
		t.Errorf("expected overwritten content, got: %s", string(content))
	}
}

func TestWriteFiles_PreservesPermissions(t *testing.T) {
	dir := t.TempDir()

	files := []OutputFile{
		{Path: filepath.Join(dir, "test.yml"), Content: "x: 1\n"},
	}

	written, err := WriteFiles(files, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(written[0])
	if err != nil {
		t.Fatalf("could not stat file: %v", err)
	}
	if info.Mode() != 0644 {
		t.Errorf("expected file mode 0644, got %v", info.Mode())
	}
}

func TestWriteFiles_EmptyList(t *testing.T) {
	written, err := WriteFiles([]OutputFile{}, true)
	if err != nil {
		t.Fatalf("unexpected error for empty file list: %v", err)
	}
	if len(written) != 0 {
		t.Errorf("expected 0 written files, got %d", len(written))
	}
}

func TestWriteFiles_WritesCorrectContent(t *testing.T) {
	dir := t.TempDir()
	content := RenderPreviewConfig(func() Config {
		cfg := DefaultConfig()
		cfg.Stack = StackSingle
		cfg.Services = ServicesForStack(cfg)
		return cfg
	}())

	files := []OutputFile{
		{Path: filepath.Join(dir, ".github", "pr-preview.yml"), Content: content},
	}

	written, err := WriteFiles(files, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(written[0])
	if err != nil {
		t.Fatalf("could not read file: %v", err)
	}
	if string(got) != content {
		t.Error("written content does not match rendered preview config")
	}
}
