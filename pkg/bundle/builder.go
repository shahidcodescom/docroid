package bundle

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/registry"
)

type Builder struct {
	store *app.Store
}

func NewBuilder(store *app.Store) *Builder {
	return &Builder{store: store}
}

type buildInstruction struct {
	op   string
	args string
	line int
}

// BuildImage builds a local Docroid image from a Docroidfile or Dockerfile and context directory
func (b *Builder) BuildImage(docroidfilePath string, contextDir string, targetTag string, out io.Writer) (*Manifest, error) {
	if contextDir == "" {
		contextDir = "."
	}
	absContextDir, err := filepath.Abs(contextDir)
	if err != nil {
		return nil, fmt.Errorf("invalid context directory: %w", err)
	}

	// Auto-detect build file: Docroidfile or Dockerfile
	if docroidfilePath == "" {
		if _, err := os.Stat(filepath.Join(absContextDir, "Docroidfile")); err == nil {
			docroidfilePath = filepath.Join(absContextDir, "Docroidfile")
		} else if _, err := os.Stat(filepath.Join(absContextDir, "Dockerfile")); err == nil {
			docroidfilePath = filepath.Join(absContextDir, "Dockerfile")
		} else {
			docroidfilePath = filepath.Join(absContextDir, "Docroidfile")
		}
	}

	file, err := os.Open(docroidfilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open build file at %s: %w", docroidfilePath, err)
	}
	defer file.Close()

	// Parse instructions
	var instructions []buildInstruction
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		op := strings.ToUpper(parts[0])
		args := ""
		if len(parts) > 1 {
			args = strings.TrimSpace(parts[1])
		}
		instructions = append(instructions, buildInstruction{op: op, args: args, line: lineNum})
	}

	manifest := &Manifest{
		FormatVersion: CurrentFormatVersion,
		Architecture:  runtime.GOARCH,
		Engine:        app.EngineNative,
		Env:           make(map[string]string),
		CreatedAt:     time.Now().UTC(),
	}

	if targetTag != "" {
		parts := strings.Split(targetTag, ":")
		manifest.Name = parts[0]
		if len(parts) > 1 {
			manifest.Version = parts[1]
		} else {
			manifest.Version = "latest"
		}
	} else {
		manifest.Name = filepath.Base(absContextDir)
		manifest.Version = "latest"
	}

	tagKey := fmt.Sprintf("%s:%s", manifest.Name, manifest.Version)
	imageDir := filepath.Join(b.store.BaseDir(), "images", tagKey)
	payloadDir := filepath.Join(imageDir, "payload")

	_ = os.RemoveAll(imageDir)
	if err := os.MkdirAll(payloadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create image directory: %w", err)
	}

	fmt.Fprintf(out, "Building image '%s' using context '%s'...\n", tagKey, absContextDir)

	totalSteps := len(instructions)
	currentWorkDir := "/"

	for idx, inst := range instructions {
		stepNum := idx + 1
		fmt.Fprintf(out, "Step %d/%d : %s %s\n", stepNum, totalSteps, inst.op, inst.args)

		switch inst.op {
		case "FROM":
			if strings.EqualFold(inst.args, "native") {
				manifest.Engine = app.EngineNative
			} else {
				manifest.Engine = app.EngineRootfs
				baseRef := inst.args

				// Check if base image exists locally
				baseImg, err := GetImage(b.store, baseRef)
				if err == nil && baseImg != nil {
					basePayload := filepath.Join(b.store.BaseDir(), "images", baseImg.TaggedName(), "payload")
					if _, err := os.Stat(basePayload); err == nil {
						fmt.Fprintf(out, "  -> Using local base image %s\n", baseImg.TaggedName())
						if err := copyRecursive(basePayload, payloadDir); err != nil {
							return nil, fmt.Errorf("failed to copy base image payload: %w", err)
						}
						for k, v := range baseImg.Env {
							manifest.Env[k] = v
						}
						if baseImg.Command != "" {
							manifest.Command = baseImg.Command
							manifest.Args = baseImg.Args
						}
						if baseImg.WorkDir != "" {
							currentWorkDir = baseImg.WorkDir
							manifest.WorkDir = currentWorkDir
						}
					}
				} else {
					// Check if base is an installed rootfs
					rootfsDir := b.store.RootfsDir(baseRef)
					if _, err := os.Stat(rootfsDir); err == nil {
						fmt.Fprintf(out, "  -> Using local rootfs %s\n", baseRef)
						manifest.Rootfs = baseRef
					} else {
						// Try pulling base image automatically
						fmt.Fprintf(out, "  -> Base image '%s' not found locally. Attempting pull...\n", baseRef)
						parsedRef, parseErr := registry.ParseReference(baseRef)
						if parseErr == nil {
							client := registry.NewClient()
							tempBaseDir := filepath.Join(b.store.BaseDir(), "images", parsedRef.TaggedName())
							pulledMan, pullErr := client.Pull(parsedRef, tempBaseDir, out)
							if pullErr == nil && pulledMan != nil {
								pulledPayload := filepath.Join(tempBaseDir, "payload")
								_ = copyRecursive(pulledPayload, payloadDir)
								for k, v := range pulledMan.Env {
									manifest.Env[k] = v
								}
								if pulledMan.Command != "" {
									manifest.Command = pulledMan.Command
									manifest.Args = pulledMan.Args
								}
							} else {
								manifest.Rootfs = baseRef
							}
						} else {
							manifest.Rootfs = baseRef
						}
					}
				}
			}

		case "NAME":
			manifest.Name = inst.args

		case "VERSION", "TAG":
			manifest.Version = inst.args

		case "ARCH":
			manifest.Architecture = inst.args

		case "PORT", "EXPOSE":
			if p, err := strconv.Atoi(inst.args); err == nil {
				manifest.Port = p
				fmt.Fprintf(out, "  -> Exposing port %d\n", p)
			}

		case "ENV":
			var k, v string
			if strings.Contains(inst.args, "=") {
				kv := strings.SplitN(inst.args, "=", 2)
				k, v = strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			} else {
				kv := strings.Fields(inst.args)
				if len(kv) >= 2 {
					k, v = kv[0], kv[1]
				}
			}
			if k != "" {
				manifest.Env[k] = v
				fmt.Fprintf(out, "  -> Set ENV %s=%s\n", k, v)
			}

		case "WORKDIR":
			currentWorkDir = inst.args
			manifest.WorkDir = currentWorkDir
			// Ensure workdir exists inside payload
			targetWorkDir := filepath.Join(payloadDir, strings.TrimPrefix(currentWorkDir, "/"))
			_ = os.MkdirAll(targetWorkDir, 0755)
			fmt.Fprintf(out, "  -> WORKDIR %s\n", currentWorkDir)

		case "DESC", "DESCRIPTION", "LABEL":
			manifest.Description = inst.args

		case "COPY":
			fields := strings.Fields(inst.args)
			if len(fields) < 2 {
				return nil, fmt.Errorf("line %d: invalid COPY syntax: %s", inst.line, inst.args)
			}
			srcRel := fields[0]
			destRel := fields[1]

			srcPath := filepath.Join(absContextDir, srcRel)
			destTarget := resolvePayloadPath(payloadDir, currentWorkDir, destRel)

			fmt.Fprintf(out, "  -> COPY %s to %s\n", srcRel, destRel)
			if err := copyRecursive(srcPath, destTarget); err != nil {
				_ = os.RemoveAll(imageDir)
				return nil, fmt.Errorf("line %d COPY failed: %w", inst.line, err)
			}

		case "ADD":
			fields := strings.Fields(inst.args)
			if len(fields) < 2 {
				return nil, fmt.Errorf("line %d: invalid ADD syntax: %s", inst.line, inst.args)
			}
			src := fields[0]
			destRel := fields[1]
			destTarget := resolvePayloadPath(payloadDir, currentWorkDir, destRel)

			fmt.Fprintf(out, "  -> ADD %s to %s\n", src, destRel)
			if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
				// Download remote URL
				if err := downloadFile(src, destTarget); err != nil {
					return nil, fmt.Errorf("line %d ADD URL failed: %w", inst.line, err)
				}
			} else {
				srcPath := filepath.Join(absContextDir, src)
				// Check if tar archive
				if strings.HasSuffix(src, ".tar") || strings.HasSuffix(src, ".tar.gz") || strings.HasSuffix(src, ".tgz") {
					if err := extractArchive(srcPath, destTarget); err != nil {
						// Fallback to regular copy
						_ = copyRecursive(srcPath, destTarget)
					}
				} else {
					if err := copyRecursive(srcPath, destTarget); err != nil {
						return nil, fmt.Errorf("line %d ADD failed: %w", inst.line, err)
					}
				}
			}

		case "RUN":
			execDir := resolvePayloadPath(payloadDir, currentWorkDir, ".")
			cmdStr := inst.args
			fmt.Fprintf(out, "  -> Executing: %s\n", cmdStr)

			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/C", cmdStr)
			} else {
				cmd = exec.Command("sh", "-c", cmdStr)
			}
			cmd.Dir = execDir

			// Inject environment
			cmd.Env = os.Environ()
			for k, v := range manifest.Env {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
			}

			cmd.Stdout = out
			cmd.Stderr = out

			if err := cmd.Run(); err != nil {
				return nil, fmt.Errorf("line %d RUN '%s' failed: %w", inst.line, cmdStr, err)
			}

		case "ENTRYPOINT":
			if strings.HasPrefix(inst.args, "[") && strings.HasSuffix(inst.args, "]") {
				var cmdArr []string
				if err := json.Unmarshal([]byte(inst.args), &cmdArr); err == nil && len(cmdArr) > 0 {
					manifest.Command = cmdArr[0]
					if len(cmdArr) > 1 {
						manifest.Args = append(cmdArr[1:], manifest.Args...)
					}
				}
			} else {
				manifest.Command = inst.args
			}

		case "CMD":
			if strings.HasPrefix(inst.args, "[") && strings.HasSuffix(inst.args, "]") {
				var cmdArr []string
				if err := json.Unmarshal([]byte(inst.args), &cmdArr); err == nil && len(cmdArr) > 0 {
					if manifest.Command == "" {
						manifest.Command = cmdArr[0]
						manifest.Args = cmdArr[1:]
					} else {
						manifest.Args = cmdArr
					}
				}
			} else {
				cmdFields := strings.Fields(inst.args)
				if len(cmdFields) > 0 {
					if manifest.Command == "" {
						manifest.Command = cmdFields[0]
						manifest.Args = cmdFields[1:]
					} else {
						manifest.Args = cmdFields
					}
				}
			}

		default:
			fmt.Fprintf(out, "Warning: unknown directive '%s' on line %d\n", inst.op, inst.line)
		}
	}

	// Ensure default command if none specified
	if manifest.Command == "" {
		if manifest.Engine == app.EngineRootfs {
			manifest.Command = "/bin/sh"
		} else {
			manifest.Command = "sh"
		}
	}

	// Make relative command in payload executable
	cmdRel := strings.TrimPrefix(manifest.Command, "/")
	cmdInPayload := filepath.Join(payloadDir, cmdRel)
	if fi, err := os.Stat(cmdInPayload); err == nil && !fi.IsDir() {
		_ = os.Chmod(cmdInPayload, 0755)
	}

	// Save manifest.json
	manifestPath := filepath.Join(imageDir, "manifest.json")
	if err := manifest.Save(manifestPath); err != nil {
		return nil, fmt.Errorf("failed to save manifest: %w", err)
	}

	fmt.Fprintf(out, "Successfully built image '%s'\n", tagKey)
	fmt.Fprintf(out, "Successfully tagged %s\n", tagKey)
	return manifest, nil
}

func resolvePayloadPath(payloadDir, workDir, targetRel string) string {
	if strings.HasPrefix(targetRel, "/") {
		return filepath.Join(payloadDir, strings.TrimPrefix(targetRel, "/"))
	}
	cleanWorkDir := strings.TrimPrefix(workDir, "/")
	return filepath.Join(payloadDir, cleanWorkDir, targetRel)
}

func copyRecursive(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		if destInfo, err := os.Stat(dest); err == nil && destInfo.IsDir() {
			dest = filepath.Join(dest, filepath.Base(src))
		}
	}

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
			if err := copyRecursive(srcSub, destSub); err != nil {
				return err
			}
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func downloadFile(urlStr, dest string) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractArchive(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	var tr *tar.Reader
	if strings.HasSuffix(archivePath, ".tar.gz") || strings.HasSuffix(archivePath, ".tgz") {
		gz, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gz.Close()
		tr = tar.NewReader(gz)
	} else {
		tr = tar.NewReader(f)
	}

	_ = os.MkdirAll(destDir, 0755)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, hdr.Name)
		if !strings.HasPrefix(target, filepath.Clean(destDir)+string(os.PathSeparator)) && target != destDir {
			continue
		}

		if hdr.Typeflag == tar.TypeDir {
			_ = os.MkdirAll(target, 0755)
		} else if hdr.Typeflag == tar.TypeReg {
			_ = os.MkdirAll(filepath.Dir(target), 0755)
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, hdr.FileInfo().Mode().Perm())
			if err != nil {
				return err
			}
			_, _ = io.Copy(outFile, tr)
			outFile.Close()
		}
	}
	return nil
}
