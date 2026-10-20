package bundle

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"docroid/pkg/app"
	"docroid/pkg/registry"
)

// ListImages returns all installed images in ~/.docroid/images/
func ListImages(store *app.Store) ([]*Manifest, error) {
	imagesBase := filepath.Join(store.BaseDir(), "images")
	if _, err := os.Stat(imagesBase); os.IsNotExist(err) {
		return nil, nil
	}

	var results []*Manifest
	seen := make(map[string]bool)

	_ = filepath.Walk(imagesBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "manifest.json" {
			manifest, err := LoadManifest(path)
			if err == nil {
				key := manifest.Name + ":" + manifest.Version
				if !seen[key] {
					seen[key] = true
					results = append(results, manifest)
				}
			}
		}
		return nil
	})

	return results, nil
}

// GetImage retrieves manifest for a given image tag or name
func GetImage(store *app.Store, tag string) (*Manifest, error) {
	all, err := ListImages(store)
	if err != nil {
		return nil, err
	}

	// 1. Exact match name:version
	for _, m := range all {
		if m.Name+":"+m.Version == tag {
			return m, nil
		}
	}

	// 2. Match with :latest suffix
	for _, m := range all {
		if m.Name == tag && (m.Version == "latest" || m.Version == "") {
			return m, nil
		}
		if tag == m.Name+":latest" {
			return m, nil
		}
	}

	// 3. Search for repository basename match (e.g. "alpine" matching "library/alpine")
	for _, m := range all {
		if m.Name == tag || strings.HasSuffix(m.Name, "/"+tag) {
			return m, nil
		}
	}

	return nil, fmt.Errorf("image '%s' not found", tag)
}

// RemoveImage deletes an installed image
func RemoveImage(store *app.Store, tag string) error {
	imagesBase := filepath.Join(store.BaseDir(), "images")
	if _, err := os.Stat(imagesBase); os.IsNotExist(err) {
		return fmt.Errorf("image '%s' does not exist", tag)
	}

	deleted := false
	_ = filepath.Walk(imagesBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == "manifest.json" {
			m, err := LoadManifest(path)
			if err == nil {
				fullTag := m.Name + ":" + m.Version
				if fullTag == tag || m.Name == tag || tag == m.Name+":latest" {
					dir := filepath.Dir(path)
					_ = os.RemoveAll(dir)
					deleted = true
				}
			}
		}
		return nil
	})

	if !deleted {
		imageDir := filepath.Join(imagesBase, tag)
		if _, err := os.Stat(imageDir); err == nil {
			return os.RemoveAll(imageDir)
		}
		return fmt.Errorf("image '%s' does not exist", tag)
	}
	return nil
}

// PullImage downloads an image from an OCI/Docker registry or a .dpack bundle URL
func PullImage(store *app.Store, imageRef string, out io.Writer) (*registry.PullResult, error) {
	if strings.HasPrefix(imageRef, "http://") || strings.HasPrefix(imageRef, "https://") {
		// Remote URL download
		if strings.HasSuffix(imageRef, ".dpack") {
			// Download dpack and import
			resp, err := http.Get(imageRef)
			if err != nil {
				return nil, fmt.Errorf("failed to download bundle: %w", err)
			}
			defer resp.Body.Close()
			tmpFile, err := os.CreateTemp("", "docroid-bundle-*.dpack")
			if err != nil {
				return nil, err
			}
			defer os.Remove(tmpFile.Name())
			if _, err := io.Copy(tmpFile, resp.Body); err != nil {
				tmpFile.Close()
				return nil, fmt.Errorf("failed to save bundle: %w", err)
			}
			tmpFile.Close()
			man, err := ImportPackage(store, tmpFile.Name(), out)
			if err != nil {
				return nil, err
			}
			return &registry.PullResult{
				Name:    man.Name,
				Tag:     man.Version,
				Engine:  string(man.Engine),
				Command: man.Command,
				Args:    man.Args,
			}, nil
		}
	}

	parsedRef, err := registry.ParseReference(imageRef)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference '%s': %w", imageRef, err)
	}

	destDir := filepath.Join(store.BaseDir(), "images", parsedRef.TaggedName())
	client := registry.NewClient()
	return client.Pull(parsedRef, destDir, out)
}

// PushImage pushes a locally installed Docroid image to an OCI/Docker registry
func PushImage(store *app.Store, imageRef string, out io.Writer) (string, error) {
	parsedRef, err := registry.ParseReference(imageRef)
	if err != nil {
		return "", fmt.Errorf("invalid image reference '%s': %w", imageRef, err)
	}

	// Locate image directory
	imageDir := filepath.Join(store.BaseDir(), "images", parsedRef.TaggedName())
	if _, err := os.Stat(imageDir); os.IsNotExist(err) {
		// Try without registry domain (e.g. repo:tag)
		fallbackDir := filepath.Join(store.BaseDir(), "images", parsedRef.Repository+":"+parsedRef.Tag)
		if _, err2 := os.Stat(fallbackDir); err2 == nil {
			imageDir = fallbackDir
		} else {
			// Search installed images
			imgs, _ := ListImages(store)
			found := false
			for _, m := range imgs {
				if m.Name == parsedRef.Repository && (m.Version == parsedRef.Tag || parsedRef.Tag == "latest") {
					imageDir = filepath.Join(store.BaseDir(), "images", m.Name+":"+m.Version)
					found = true
					break
				}
			}
			if !found {
				return "", fmt.Errorf("image '%s' not found locally in image store", imageRef)
			}
		}
	}

	client := registry.NewClient()
	return client.Push(parsedRef, imageDir, out)
}

