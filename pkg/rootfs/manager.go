package rootfs

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/utils"
)

type Manager struct {
	store *app.Store
}

func NewManager(store *app.Store) *Manager {
	return &Manager{store: store}
}

// GetAlpineArch maps Go architecture to Alpine Linux release architecture string
func GetAlpineArch(goarch string) string {
	switch goarch {
	case "arm64":
		return "aarch64"
	case "amd64":
		return "x86_64"
	case "arm":
		return "armhf"
	default:
		return "aarch64"
	}
}

// Pull downloads and unpacks a minimal rootfs distribution
func (m *Manager) Pull(name string, out io.Writer) error {
	if name == "" {
		name = "alpine"
	}

	targetDir := m.store.RootfsDir(name)
	if _, err := os.Stat(targetDir); err == nil {
		fmt.Fprintf(out, "Rootfs '%s' is already installed at %s\n", name, targetDir)
		return nil
	}

	arch := GetAlpineArch(runtime.GOARCH)
	// Alpine 3.20 minirootfs URL (official CDN)
	// Size is approximately 3.3 MB
	version := "3.20.3"
	url := fmt.Sprintf("https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/%s/alpine-minirootfs-%s-%s.tar.gz", arch, version, arch)

	fmt.Fprintf(out, "Downloading %s mini-rootfs (%s) for %s...\n", name, version, arch)
	fmt.Fprintf(out, "Source: %s\n", url)

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		// Try fallback to 3.20.0
		url = fmt.Sprintf("https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/%s/alpine-minirootfs-3.20.0-%s.tar.gz", arch, arch)
		resp, err = client.Get(url)
		if err != nil {
			return fmt.Errorf("failed to download rootfs from %s: %w", url, err)
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %s from %s", resp.Status, url)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create rootfs directory: %w", err)
	}

	fmt.Fprintf(out, "Extracting rootfs filesystem to %s...\n", targetDir)
	if err := extractTarGz(resp.Body, targetDir, out); err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("failed to extract rootfs: %w", err)
	}

	// Create basic DNS config in rootfs if missing
	resolvPath := filepath.Join(targetDir, "etc", "resolv.conf")
	if _, err := os.Stat(resolvPath); os.IsNotExist(err) {
		_ = os.MkdirAll(filepath.Join(targetDir, "etc"), 0755)
		_ = os.WriteFile(resolvPath, []byte("nameserver 1.1.1.1\nnameserver 8.8.8.8\n"), 0644)
	}

	fmt.Fprintf(out, "✓ Rootfs '%s' successfully installed!\n", name)
	return nil
}

// List returns installed rootfs distributions
func (m *Manager) List() ([]string, error) {
	rootfsBase := filepath.Join(m.store.BaseDir(), "rootfs")
	entries, err := os.ReadDir(rootfsBase)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, entry := range entries {
		if entry.IsDir() {
			result = append(result, entry.Name())
		}
	}
	return result, nil
}

// Remove deletes an installed rootfs
func (m *Manager) Remove(name string) error {
	dir := m.store.RootfsDir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("rootfs '%s' is not installed", name)
	}
	return os.RemoveAll(dir)
}

func extractTarGz(r io.Reader, destDir string, out io.Writer) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	count := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Security: prevent zip slip path traversal
		target := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != destDir {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode().Perm())
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		case tar.TypeSymlink:
			_ = os.Remove(target)
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			_ = os.Symlink(header.Linkname, target)
		}
		count++
		if count%500 == 0 {
			fmt.Fprintf(out, "Unpacked %d files...\n", count)
		}
	}

	_ = utils.FormatBytes(0) // reference utils
	return nil
}
