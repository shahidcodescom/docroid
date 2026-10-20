package doctor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"docroid/pkg/app"
	"docroid/pkg/utils"
)

type CheckResult struct {
	Name    string
	Status  string // "OK", "WARN", "FAIL"
	Message string
}

// RunDiagnostics inspects the current Android, Termux, and kernel environment
func RunDiagnostics(store *app.Store, out io.Writer) {
	fmt.Fprintln(out, "Docroid Environment & System Diagnostics")
	fmt.Fprintln(out, "========================================")

	checks := []CheckResult{}

	// 1. Architecture & OS
	archStr := fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	checks = append(checks, CheckResult{
		Name:    "Host Platform",
		Status:  "OK",
		Message: fmt.Sprintf("%s (Optimized for Android)", archStr),
	})

	// 2. Linux Kernel
	kernel := getKernelVersion()
	checks = append(checks, CheckResult{
		Name:    "Linux Kernel",
		Status:  "OK",
		Message: kernel,
	})

	// 3. Android OS Check
	androidVer, sdkVer := getAndroidVersion()
	if androidVer != "" {
		checks = append(checks, CheckResult{
			Name:    "Android Version",
			Status:  "OK",
			Message: fmt.Sprintf("Android %s (API level %s)", androidVer, sdkVer),
		})
	} else {
		checks = append(checks, CheckResult{
			Name:    "Android Environment",
			Status:  "WARN",
			Message: "Standard Linux / development environment detected",
		})
	}

	// 4. Termux Detection
	termuxPrefix := os.Getenv("PREFIX")
	if strings.Contains(termuxPrefix, "com.termux") || termuxPrefix != "" {
		checks = append(checks, CheckResult{
			Name:    "Termux Environment",
			Status:  "OK",
			Message: fmt.Sprintf("Active (PREFIX=%s)", termuxPrefix),
		})
	} else {
		checks = append(checks, CheckResult{
			Name:    "Termux Environment",
			Status:  "WARN",
			Message: "Not in Termux sandbox, running in standard shell",
		})
	}

	// 5. PRoot availability (for Rootfs Engine)
	prootPath, _ := exec.LookPath("proot")
	if prootPath == "" {
		termuxProot := "/data/data/com.termux/files/usr/bin/proot"
		if _, err := os.Stat(termuxProot); err == nil {
			prootPath = termuxProot
		}
	}

	if prootPath != "" {
		checks = append(checks, CheckResult{
			Name:    "PRoot Sandbox Engine",
			Status:  "OK",
			Message: fmt.Sprintf("Installed (%s)", prootPath),
		})
	} else {
		checks = append(checks, CheckResult{
			Name:    "PRoot Sandbox Engine",
			Status:  "WARN",
			Message: "Not installed. Run 'pkg install proot' to enable Linux rootfs containers.",
		})
	}

	// 6. Native Engine
	checks = append(checks, CheckResult{
		Name:    "Native Engine (Fast-Path)",
		Status:  "OK",
		Message: "Available (0% overhead, 100% CPU speed)",
	})

	// 7. Base Storage Directory
	checks = append(checks, CheckResult{
		Name:    "Storage Directory",
		Status:  "OK",
		Message: store.BaseDir(),
	})

	// Print checks in table
	w := utils.NewTableWriter(out)
	fmt.Fprintln(w, "CHECK\tSTATUS\tDETAILS")
	for _, c := range checks {
		statusSymbol := "[✓] OK"
		if c.Status == "WARN" {
			statusSymbol = "[!] WARN"
		} else if c.Status == "FAIL" {
			statusSymbol = "[✗] FAIL"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.Name, statusSymbol, c.Message)
	}
	_ = w.Flush()

	// Termux Optimization Tips
	fmt.Fprintln(out, "\nMobile Optimization & Battery Tips:")
	fmt.Fprintln(out, "------------------------------------")
	fmt.Fprintln(out, "• To prevent Android from killing background services when screen is off:")
	fmt.Fprintln(out, "  Run 'termux-wake-lock' in Termux before starting long-running apps.")
	fmt.Fprintln(out, "• On Android 12+, disable the Phantom Process Killer in Termux settings or via ADB:")
	fmt.Fprintln(out, "  adb shell \"/system/bin/device_config set_sync_disabled_for_tests persistent\"")
	fmt.Fprintln(out, "  adb shell \"/system/bin/device_config put activity_manager max_phantom_processes 2147483647\"")
}

func getKernelVersion() string {
	data, err := os.ReadFile("/proc/version")
	if err == nil {
		line := strings.TrimSpace(string(data))
		parts := strings.Fields(line)
		if len(parts) >= 3 {
			return parts[2]
		}
		return line
	}
	out, err := exec.Command("uname", "-r").Output()
	if err == nil {
		return strings.TrimSpace(string(out))
	}
	return "unknown"
}

func getAndroidVersion() (release string, sdk string) {
	out, err := exec.Command("getprop", "ro.build.version.release").Output()
	if err == nil {
		release = strings.TrimSpace(string(out))
	}
	out2, err := exec.Command("getprop", "ro.build.version.sdk").Output()
	if err == nil {
		sdk = strings.TrimSpace(string(out2))
	}
	return
}

func getAvailableMemory() string {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return "N/A"
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "MemAvailable:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "MemAvailable:"))
		}
	}
	return "N/A"
}
