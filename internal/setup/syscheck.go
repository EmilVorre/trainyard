package setup

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/Emilvorre/trainyard/internal/tui"
)

const (
	minRAMMB  = 1800 // ~2GB
	minDiskGB = 10
)

type sysInfo struct {
	OS      string
	Arch    string
	RAMMB   int
	DiskGB  int
	CPUs    int
	IsRoot  bool
	Missing []string // required tools not found
}

func checkSystem() (*sysInfo, error) {
	info := &sysInfo{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		CPUs: runtime.NumCPU(),
	}

	// Must be root
	info.IsRoot = os.Geteuid() == 0

	// RAM (Linux: /proc/meminfo)
	if ram, err := readMemInfoMB(); err == nil {
		info.RAMMB = ram
	}

	// Disk (df on /)
	if disk, err := rootDiskFreeGB(); err == nil {
		info.DiskGB = disk
	}

	return info, nil
}

func printSystemSummary(info *sysInfo) {
	tui.Section("System Check")

	ramOK := info.RAMMB >= minRAMMB
	diskOK := info.DiskGB >= minDiskGB

	ramStr := fmt.Sprintf("%d MB", info.RAMMB)
	if !ramOK {
		ramStr += " (minimum 2 GB recommended)"
	}

	diskStr := fmt.Sprintf("%d GB free", info.DiskGB)
	if !diskOK {
		diskStr += " (minimum 10 GB recommended)"
	}

	rows := []struct {
		label string
		value string
		ok    bool
	}{
		{"OS", info.OS + "/" + info.Arch, info.OS == "linux"},
		{"CPUs", strconv.Itoa(info.CPUs), info.CPUs >= 1},
		{"RAM", ramStr, ramOK},
		{"Disk", diskStr, diskOK},
		{"Running as root", strconv.FormatBool(info.IsRoot), info.IsRoot},
	}

	for _, r := range rows {
		icon := "✓"
		if !r.ok {
			icon = "⚠"
		}
		fmt.Printf("  %s  %-20s %s\n", icon, r.label, r.value)
	}

	if info.OS != "linux" {
		tui.Fatal("yard setup must run on Linux (detected: %s)", info.OS)
	}
	if !info.IsRoot {
		tui.Fatal("yard setup must run as root (use sudo)")
	}
	if !ramOK {
		tui.Warn("Low RAM detected — k3s may be unstable")
	}
	if !diskOK {
		tui.Warn("Low disk space — recommend freeing space before continuing")
	}
}

// readMemInfoMB reads total RAM from /proc/meminfo
func readMemInfoMB() (int, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, err := strconv.Atoi(fields[1])
				if err != nil {
					return 0, err
				}
				return kb / 1024, nil
			}
		}
	}
	return 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
}

// rootDiskFreeGB uses df to get free space on /
func rootDiskFreeGB() (int, error) {
	out, err := exec.Command("df", "-BG", "--output=avail", "/").Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output")
	}
	val := strings.TrimSuffix(strings.TrimSpace(lines[1]), "G")
	return strconv.Atoi(val)
}
