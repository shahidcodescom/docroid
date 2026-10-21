package utils

import (
	"io"
	"os"
	"path/filepath"
)

// CopyRecursive copies a file or directory tree recursively from src to dest,
// properly preserving symlinks and directory permissions.
func CopyRecursive(src, dest string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	// Handle Symlink
	if info.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(src)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		_ = os.Remove(dest)
		return os.Symlink(linkTarget, dest)
	}

	// Handle Directory
	if info.IsDir() {
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			srcSub := filepath.Join(src, entry.Name())
			destSub := filepath.Join(dest, entry.Name())
			if err := CopyRecursive(srcSub, destSub); err != nil {
				return err
			}
		}
		return nil
	}

	// Handle Regular File
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
