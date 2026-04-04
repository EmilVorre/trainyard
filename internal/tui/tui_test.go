package tui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// withStdin temporarily replaces os.Stdin with a reader containing the given input,
// runs fn, then restores the original stdin.
func withStdin(t *testing.T, input string, fn func()) {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create pipe: %v", err)
	}

	_, err = io.WriteString(w, input)
	if err != nil {
		t.Fatalf("could not write to pipe: %v", err)
	}
	w.Close()

	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() { os.Stdin = orig; r.Close() })

	fn()
}

// withStdout captures stdout output during fn and returns it.
func withStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("could not create pipe: %v", err)
	}

	orig := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	fn()

	w.Close()
	os.Stdout = orig

	var buf strings.Builder
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("could not read stdout: %v", err)
	}
	return buf.String()
}

// Prompt tests

func TestPrompt_ReturnsInput(t *testing.T) {
	var got string
	withStdin(t, "myvalue\n", func() {
		got = Prompt("Label", "default")
	})
	if got != "myvalue" {
		t.Errorf("expected myvalue, got %q", got)
	}
}

func TestPrompt_ReturnsDefaultOnEmptyInput(t *testing.T) {
	var got string
	withStdin(t, "\n", func() {
		got = Prompt("Label", "mydefault")
	})
	if got != "mydefault" {
		t.Errorf("expected mydefault, got %q", got)
	}
}

func TestPrompt_ReturnsDefaultOnNoInput(t *testing.T) {
	var got string
	withStdin(t, "", func() {
		got = Prompt("Label", "fallback")
	})
	if got != "fallback" {
		t.Errorf("expected fallback, got %q", got)
	}
}

func TestPrompt_TrimsWhitespace(t *testing.T) {
	var got string
	withStdin(t, "  trimmed  \n", func() {
		got = Prompt("Label", "default")
	})
	if got != "trimmed" {
		t.Errorf("expected trimmed, got %q", got)
	}
}

func TestPrompt_WhitespaceOnlyInputUsesDefault(t *testing.T) {
	var got string
	withStdin(t, "   \n", func() {
		got = Prompt("Label", "default")
	})
	if got != "default" {
		t.Errorf("expected default when whitespace-only input, got %q", got)
	}
}

func TestPrompt_EmptyDefaultShowsNoDefault(t *testing.T) {
	var output string
	withStdin(t, "val\n", func() {
		output = withStdout(t, func() {
			Prompt("My label", "")
		})
	})
	if strings.Contains(output, "[") {
		t.Errorf("expected no default bracket when defaultVal is empty, got: %s", output)
	}
}

func TestPrompt_NonEmptyDefaultShowsDefault(t *testing.T) {
	var output string
	withStdin(t, "\n", func() {
		output = withStdout(t, func() {
			Prompt("My label", "nginx")
		})
	})
	if !strings.Contains(output, "[nginx]") {
		t.Errorf("expected default value shown in prompt, got: %s", output)
	}
}

// Confirm tests

func TestConfirm_YesLowercase(t *testing.T) {
	var got bool
	withStdin(t, "y\n", func() {
		got = Confirm("Are you sure?")
	})
	if !got {
		t.Error("expected true for 'y' input")
	}
}

func TestConfirm_YesUppercase(t *testing.T) {
	var got bool
	withStdin(t, "Y\n", func() {
		got = Confirm("Are you sure?")
	})
	if !got {
		t.Error("expected true for 'Y' input")
	}
}

func TestConfirm_NoLowercase(t *testing.T) {
	var got bool
	withStdin(t, "n\n", func() {
		got = Confirm("Are you sure?")
	})
	if got {
		t.Error("expected false for 'n' input")
	}
}

func TestConfirm_EmptyInput(t *testing.T) {
	var got bool
	withStdin(t, "\n", func() {
		got = Confirm("Are you sure?")
	})
	if got {
		t.Error("expected false for empty input (default is N)")
	}
}

func TestConfirm_ArbitraryInput(t *testing.T) {
	for _, input := range []string{"yes\n", "no\n", "maybe\n", "1\n", "\n"} {
		var got bool
		withStdin(t, input, func() {
			got = Confirm("Prompt")
		})
		if got {
			t.Errorf("expected false for input %q (only 'y' or 'Y' should return true)", input)
		}
	}
}

// Step tests

func TestStep_ReturnsNilOnSuccess(t *testing.T) {
	err := Step("doing a thing", func() error { return nil })
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestStep_ReturnsErrorOnFailure(t *testing.T) {
	expected := fmt.Errorf("something went wrong")
	err := Step("doing a thing", func() error { return expected })
	if err != expected {
		t.Errorf("expected %v, got %v", expected, err)
	}
}

func TestStep_RunsFunction(t *testing.T) {
	called := false
	Step("check", func() error {
		called = true
		return nil
	})
	if !called {
		t.Error("expected Step to call the provided function")
	}
}
