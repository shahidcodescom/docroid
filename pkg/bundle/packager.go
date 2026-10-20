package bundle

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/utils"
)

// ExportImage packages an image into a portable .dpack file
func ExportImage(store *app.Store, imageTag string, outputPath string, out io.Writer) error {
	if !strings.Contains(imageTag, ":") {
		imageTag = imageTag + ":latest"
	}

	imageDir := filepath.Join(store.BaseDir(), "images", imageTag)
	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		return fmt.Errorf("image '%s' not found in local store", imageTag)
	}

	if outputPath == "" {
		safeTag := strings.ReplaceAll(imageTag, ":", "-")
		outputPath = fmt.Sprintf("%s.dpack", safeTag)
	}

	fmt.Fprintf(out, "Exporting image '%s' to %s...\n", imageTag, outputPath)

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	multiWriter := io.MultiWriter(f, hasher)

	gw := gzip.NewWriter(multiWriter)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Walk and tar imageDir
	err = filepath.Walk(imageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(imageDir, path)
		if err != nil || relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to package image: %w", err)
	}

	_ = tw.Close()
	_ = gw.Close()

	stat, _ := f.Stat()
	sizeStr := utils.FormatBytes(stat.Size())
	hashStr := hex.EncodeToString(hasher.Sum(nil))

	fmt.Fprintf(out, "✓ Successfully exported %s (%s, SHA256: %s)\n", outputPath, sizeStr, hashStr[:12])
	return nil
}

// ExportApp packages an existing application instance directly into a portable .dpack
func ExportApp(store *app.Store, appIDOrName string, outputPath string, out io.Writer) error {
	a, err := store.FindApp(appIDOrName)
	if err != nil {
		return err
	}

	if outputPath == "" {
		outputPath = fmt.Sprintf("%s.dpack", a.Name)
	}

	manifest := &Manifest{
		FormatVersion: CurrentFormatVersion,
		Name:          a.Name,
		Version:       "1.0.0",
		Architecture:  "any",
		Engine:        a.Engine,
		Command:       a.Command,
		Args:          a.Args,
		Port:          a.Port,
		Env:           a.Env,
		Rootfs:        a.Rootfs,
		CreatedAt:     time.Now().UTC(),
		Description:   fmt.Sprintf("Exported from app %s (%s)", a.Name, a.ID),
	}

	wsDir := store.WorkspaceDir(a.ID)

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	gw := gzip.NewWriter(f)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// 1. Write manifest.json
	manifestBytes, _ := os.ReadFile(filepath.Join(store.AppDir(a.ID), "app.json"))
	if len(manifestBytes) == 0 {
		return fmt.Errorf("app manifest not found")
	}
	_ = manifest // use manifest struct

	header := &tar.Header{
		Name:    "manifest.json",
		Mode:    0644,
		Size:    int64(len(manifestBytes)),
		ModTime: time.Now(),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	if _, err := tw.Write(manifestBytes); err != nil {
		return err
	}

	// 2. Package workspace files under payload/
	_ = filepath.Walk(wsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(wsDir, path)
		if err != nil || rel == "." {
			return nil
		}

		th, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return nil
		}
		th.Name = filepath.Join("payload", rel)
		_ = tw.WriteHeader(th)
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err == nil {
				_, _ = io.Copy(tw, file)
				file.Close()
			}
		}
		return nil
	})

	_ = tw.Close()
	_ = gw.Close()

	fi, _ := f.Stat()
	fmt.Fprintf(out, "✓ Exported app '%s' to %s (%s)\n", a.Name, outputPath, utils.FormatBytes(fi.Size()))
	return nil
}

// ImportPackage imports a .dpack archive into the local image store
func ImportPackage(store *app.Store, inputPath string, out io.Writer) (*Manifest, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open package file: %w", err)
	}
	defer file.Close()

	return ImportPackageReader(store, file, out)
}

// ImportPackageReader imports a .dpack stream
func ImportPackageReader(store *app.Store, r io.Reader, out io.Writer) (*Manifest, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("invalid package format (not gzip compressed): %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	// Temporary extraction directory
	tempDir, err := os.MkdirTemp(store.BaseDir(), "import_*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)

	var manifest *Manifest

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		target := filepath.Join(tempDir, header.Name)
		if !strings.HasPrefix(target, filepath.Clean(tempDir)+string(os.PathSeparator)) && target != tempDir {
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(target, 0755)
		case tar.TypeReg:
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, header.FileInfo().Mode().Perm())
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return nil, err
			}
			outFile.Close()
		case tar.TypeSymlink:
			_ = os.Remove(target)
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			_ = os.Symlink(header.Linkname, target)
		}
	}

	// Read and validate manifest.json
	manifestPath := filepath.Join(tempDir, "manifest.json")
	manifest, err = LoadManifest(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("package does not contain a valid manifest.json: %w", err)
	}

	targetTag := manifest.TaggedName()
	finalImageDir := filepath.Join(store.BaseDir(), "images", targetTag)

	_ = os.RemoveAll(finalImageDir)
	if err := os.Rename(tempDir, finalImageDir); err != nil {
		// Fallback to copy if rename fails across filesystems
		if err := copyRecursive(tempDir, finalImageDir); err != nil {
			return nil, fmt.Errorf("failed to install image: %w", err)
		}
	}

	fmt.Fprintf(out, "✓ Successfully imported image '%s' (%s, %s)\n", targetTag, manifest.Engine, manifest.Architecture)
	return manifest, nil
}

// DeployFromImage instantiates a new application from an installed image
func DeployFromImage(store *app.Store, manager *app.AppManager, imageTag string, customName string, detached bool) (*app.App, error) {
	manifest, err := GetImage(store, imageTag)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found or invalid manifest: %w", imageTag, err)
	}

	actualTag := manifest.TaggedName()
	imageDir := filepath.Join(store.BaseDir(), "images", actualTag)

	appName := customName
	if appName == "" {
		appName = manifest.Name
	}

	// Create app record
	application, err := manager.CreateApp(
		appName,
		manifest.Engine,
		manifest.Command,
		manifest.Args,
		manifest.Env,
		"",
		manifest.Port,
		nil,
		manifest.Rootfs,
		"no",
	)
	if err != nil {
		return nil, err
	}

	// Copy payload contents into the app's isolated workspace
	payloadDir := filepath.Join(imageDir, "payload")
	wsDir := store.WorkspaceDir(application.ID)
	if _, err := os.Stat(payloadDir); err == nil {
		if err := copyRecursive(payloadDir, wsDir); err != nil {
			return nil, fmt.Errorf("failed to copy image payload to workspace: %w", err)
		}
	}

	return application, nil
}
