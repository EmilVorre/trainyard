package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	success = color.New(color.FgGreen, color.Bold)
	fail    = color.New(color.FgRed, color.Bold)
	warn    = color.New(color.FgYellow)
	info    = color.New(color.FgCyan)
	bold    = color.New(color.Bold)
	dim     = color.New(color.Faint)
)

// Banner prints the Trainyard header.
func Banner() {
	fmt.Println()
	bold.Println("  🚂 Trainyard setup wizard")
	dim.Println("  Ephemeral Kubernetes PR preview environments")
	fmt.Println()
}

// Section prints a section header.
func Section(title string) {
	fmt.Println()
	info.Printf("━━ %s\n", strings.ToUpper(title))
}

// Step runs fn, printing a status line. Returns the error.
func Step(label string, fn func() error) error {
	fmt.Printf("  %-50s", label+"…")
	err := fn()
	if err != nil {
		fail.Println(" ✗")
		fail.Printf("    error: %v\n", err)
		return err
	}
	success.Println(" ✓")
	return nil
}

// StepSkip prints a skipped step (already done).
func StepSkip(label string) {
	fmt.Printf("  %-50s", label+"…")
	dim.Println(" (already installed)")
}

// Fatal prints an error and exits.
func Fatal(format string, args ...any) {
	fmt.Fprintln(os.Stderr)
	fail.Fprintf(os.Stderr, "  ✗ "+format+"\n", args...)
	os.Exit(1)
}

// Warn prints a non-fatal warning.
func Warn(format string, args ...any) {
	warn.Printf("  ⚠  "+format+"\n", args...)
}

// Info prints an informational line.
func Info(format string, args ...any) {
	info.Printf("  → "+format+"\n", args...)
}

// Success prints a success message.
func Success(format string, args ...any) {
	success.Printf("  ✓ "+format+"\n", args...)
}

// Spinner runs fn in the background with a spinner, then prints result.
func Spinner(label string, fn func() error) error {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan error, 1)

	go func() { done <- fn() }()

	i := 0
	for {
		select {
		case err := <-done:
			// Clear the spinner line
			fmt.Printf("\r  %-52s\r", "")
			if err != nil {
				fmt.Printf("  %-50s", label+"…")
				fail.Println(" ✗")
				return err
			}
			fmt.Printf("  %-50s", label+"…")
			success.Println(" ✓")
			return nil
		default:
			fmt.Printf("\r  %s %-50s", frames[i%len(frames)], label+"…")
			time.Sleep(80 * time.Millisecond)
			i++
		}
	}
}

// Box prints a highlighted info box.
func Box(title string, lines []string) {
	fmt.Println()
	bold.Printf("  ┌─ %s\n", title)
	for _, l := range lines {
		fmt.Printf("  │  %s\n", l)
	}
	bold.Println("  └" + strings.Repeat("─", 50))
	fmt.Println()
}

// Confirm asks a yes/no question on stderr, returns true if yes.
func Confirm(prompt string) bool {
	fmt.Printf("  %s [y/N]: ", prompt)
	var answer string
	fmt.Scanln(&answer)
	return strings.ToLower(strings.TrimSpace(answer)) == "y"
}

// Prompt asks for a string value with a default.
func Prompt(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("  %s: ", label)
	}
	var val string
	fmt.Scanln(&val)
	val = strings.TrimSpace(val)
	if val == "" {
		return defaultVal
	}
	return val
}
