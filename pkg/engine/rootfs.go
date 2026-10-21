package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"docroid/pkg/app"
	"docroid/pkg/network"
)

type RootfsEngine struct {
	prootPath string
}

func NewRootfsEngine() *RootfsEngine {
	// Look for proot
	path, _ := exec.LookPath("proot")
	if path == "" {
		// In Termux, check standard prefix path
		termuxProot := "/data/data/com.termux/files/usr/bin/proot"
		if _, err := os.Stat(termuxProot); err == nil {
			path = termuxProot
		}
	}
	return &RootfsEngine{prootPath: path}
}

func (e *RootfsEngine) Name() app.EngineType {
	return app.EngineRootfs
}

func (e *RootfsEngine) Prepare(cfg *RunConfig) (*exec.Cmd, error) {
	if e.prootPath == "" {
		return nil, fmt.Errorf("proot is not installed. Run 'pkg install proot' in Termux to enable rootfs sandbox")
	}

	a := cfg.App
	rootfsName := a.Rootfs
	if rootfsName == "" {
		rootfsName = "alpine"
	}

	rootfsDir := cfg.Store.RootfsDir(rootfsName)
	if _, err := os.Stat(rootfsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("rootfs '%s' is not installed. Run 'docroid rootfs pull %s' first", rootfsName, rootfsName)
	}

	// Build proot command arguments
	var prootArgs []string
	termuxTmp := "/data/data/com.termux/files/usr/tmp"
	termuxTmpExists := false
	if _, err := os.Stat(termuxTmp); err == nil {
		termuxTmpExists = true
	}

	// On Termux / Android, use --link2symlink to emulate hardlinks with symlinks
	// preventing package managers (apk, apt) from failing on Android storage.
	if strings.Contains(e.prootPath, "com.termux") || termuxTmpExists {
		prootArgs = append(prootArgs, "--link2symlink")
	}

	if cfg.User != "" {
		prootArgs = append(prootArgs, "-i", cfg.User)
	} else {
		prootArgs = append(prootArgs, "-0") // Simulate root (UID 0)
	}
	prootArgs = append(prootArgs,
		"-r", rootfsDir,
		"-b", "/dev",
		"-b", "/proc",
		"-b", "/sys",
	)

	// Mount Termux tmp if present
	if termuxTmpExists {
		prootArgs = append(prootArgs, "-b", termuxTmp+":/tmp")
	}

	// Bind isolated workspace
	wsDir := cfg.Store.WorkspaceDir(a.ID)
	_ = os.MkdirAll(wsDir, 0755)
	prootArgs = append(prootArgs, "-b", fmt.Sprintf("%s:/workspace", wsDir))

	// Bind custom volumes
	for _, v := range a.Volumes {
		hostPath := v.HostPath
		if strings.HasPrefix(hostPath, "~") {
			if home := os.Getenv("HOME"); home != "" {
				hostPath = filepath.Join(home, strings.TrimPrefix(hostPath, "~"))
			}
		}
		if absHost, err := filepath.Abs(hostPath); err == nil {
			_ = os.MkdirAll(absHost, 0755)
			prootArgs = append(prootArgs, "-b", fmt.Sprintf("%s:%s", absHost, v.ContainerPath))
		}
	}

	// Working directory inside container
	workDir := a.WorkDir
	if workDir == "" {
		workDir = "/workspace"
	}
	prootArgs = append(prootArgs, "-w", workDir)

	// Environment variables
	defaultEnvs := map[string]string{
		"PATH":              "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME":              "/root",
		"TERM":              "xterm-256color",
		"DOCROID_APP_ID":    a.ID,
		"DOCROID_APP_NAME":  a.Name,
		"DOCROID_WORKSPACE": "/workspace",
	}
	if cfg.User != "" {
		defaultEnvs["USER"] = cfg.User
		defaultEnvs["LOGNAME"] = cfg.User
		if cfg.User != "root" && cfg.User != "0" {
			defaultEnvs["HOME"] = "/home/" + cfg.User
		}
	}
	if a.Port > 0 {
		defaultEnvs["PORT"] = fmt.Sprintf("%d", a.Port)
	}

	// Network and inter-container service discovery (/etc/hosts & link envs)
	netName := a.Network
	if netName == "" {
		netName = "bridge"
	}
	if cfg.Store != nil {
		if netMgr, err := network.NewNetworkManager(cfg.Store.BaseDir()); err == nil {
			_, _ = netMgr.Connect(netName, a.ID, a.Name, a.Port, a.NetworkAliases)
			if hostsPath, err := netMgr.WriteContainerHosts(wsDir, netName, a.ID); err == nil && hostsPath != "" {
				prootArgs = append(prootArgs, "-b", fmt.Sprintf("%s:/etc/hosts", hostsPath))
			}
			if linkEnvs, err := netMgr.GenerateLinkEnvs(netName, a.ID); err == nil {
				for k, v := range linkEnvs {
					defaultEnvs[k] = v
				}
			}
		}
	}

	// The command to run inside rootfs
	cmdInside := a.Command
	if cmdInside == "" {
		cmdInside = "/bin/sh"
	}
	prootArgs = append(prootArgs, cmdInside)
	prootArgs = append(prootArgs, a.Args...)

	cmd := exec.Command(e.prootPath, prootArgs...)

	// Construct clean environment for proot and the guest container process.
	// Inherit safe host environment variables (like TERM, LANG, etc.),
	// but filter out LD_PRELOAD, LD_LIBRARY_PATH, and host paths (PATH, HOME, USER, LOGNAME)
	// so Termux libraries and host configurations don't interfere with rootfs binaries.
	var cmdEnv []string
	for _, envEntry := range os.Environ() {
		parts := strings.SplitN(envEntry, "=", 2)
		if len(parts) == 2 {
			k := parts[0]
			if strings.HasPrefix(k, "LD_") || k == "PATH" || k == "HOME" || k == "USER" || k == "LOGNAME" || k == "TMPDIR" {
				continue
			}
			if _, exists := defaultEnvs[k]; exists {
				continue
			}
			if _, exists := a.Env[k]; exists {
				continue
			}
			cmdEnv = append(cmdEnv, envEntry)
		}
	}

	for k, v := range defaultEnvs {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range a.Env {
		cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%s", k, v))
	}

	if termuxTmpExists {
		cmdEnv = append(cmdEnv, fmt.Sprintf("PROOT_TMP_DIR=%s", termuxTmp))
	}

	cmd.Env = cmdEnv

	// Configure process group attributes for clean signal handling in background/detached mode.
	// In interactive foreground mode, avoid Setpgid: true so the process stays in the foreground process group.
	if !cfg.Interactive {
		setSysProcAttr(cmd)
	}

	return cmd, nil
}

func (e *RootfsEngine) Run(ctx context.Context, cfg *RunConfig) (*os.Process, error) {
	cmd, err := e.Prepare(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Detached {
		stdoutPath, stderrPath := cfg.Store.LogPaths(cfg.App.ID)
		outFile, err := os.OpenFile(stdoutPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open stdout log: %w", err)
		}
		errFile, err := os.OpenFile(stderrPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			outFile.Close()
			return nil, fmt.Errorf("failed to open stderr log: %w", err)
		}

		cmd.Stdout = outFile
		cmd.Stderr = errFile
		cmd.Stdin = nil

		if err := cmd.Start(); err != nil {
			outFile.Close()
			errFile.Close()
			return nil, fmt.Errorf("failed to start background rootfs process: %w", err)
		}

		go func() {
			_ = cmd.Wait()
			_ = outFile.Close()
			_ = errFile.Close()
		}()

		return cmd.Process, nil
	}

	cmd.Stdin = cfg.Stdin
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start rootfs process: %w", err)
	}

	return cmd.Process, nil
}
