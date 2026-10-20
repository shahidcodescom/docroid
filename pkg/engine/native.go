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

type NativeEngine struct{}

func NewNativeEngine() *NativeEngine {
	return &NativeEngine{}
}

func (e *NativeEngine) Name() app.EngineType {
	return app.EngineNative
}

func (e *NativeEngine) Prepare(cfg *RunConfig) (*exec.Cmd, error) {
	a := cfg.App

	// Working directory: default to isolated workspace if not set
	workDir := a.WorkDir
	if workDir == "" && cfg.Store != nil {
		workDir = cfg.Store.WorkspaceDir(a.ID)
	}

	cmdPath := a.Command
	cmdArgs := append([]string{}, a.Args...)

	// Check if command exists inside the workspace (e.g. /bin/server.sh -> <workDir>/bin/server.sh)
	if workDir != "" {
		relCmd := strings.TrimPrefix(cmdPath, "/")
		inWs := filepath.Join(workDir, relCmd)
		if fi, err := os.Stat(inWs); err == nil && !fi.IsDir() {
			_ = os.Chmod(inWs, 0755)
			cmdPath = inWs
		}

		// Also check args for paths inside workspace
		for i, arg := range cmdArgs {
			relArg := strings.TrimPrefix(arg, "/")
			inWsArg := filepath.Join(workDir, relArg)
			if fi, err := os.Stat(inWsArg); err == nil && !fi.IsDir() {
				_ = os.Chmod(inWsArg, 0755)
				cmdArgs[i] = inWsArg
			}
		}

		// Process and bind host volumes
		for _, v := range a.Volumes {
			hostPath := v.HostPath
			if strings.HasPrefix(hostPath, "~") {
				if home := os.Getenv("HOME"); home != "" {
					hostPath = filepath.Join(home, strings.TrimPrefix(hostPath, "~"))
				}
			}
			absHost, err := filepath.Abs(hostPath)
			if err != nil {
				continue
			}
			_ = os.MkdirAll(absHost, 0755)

			relCont := strings.TrimPrefix(v.ContainerPath, "/")
			targetInWs := filepath.Join(workDir, relCont)

			_ = os.MkdirAll(filepath.Dir(targetInWs), 0755)
			_ = os.RemoveAll(targetInWs)
			_ = os.Symlink(absHost, targetInWs)
		}
	}

	// If command is not an absolute path or was not in workspace, search in PATH or check if it's relative
	if !filepath.IsAbs(cmdPath) {
		if _, err := os.Stat(cmdPath); err == nil {
			// Found in current directory, make it absolute
			if abs, err := filepath.Abs(cmdPath); err == nil {
				cmdPath = abs
			}
		} else {
			// Look in system PATH
			if path, err := exec.LookPath(cmdPath); err == nil {
				cmdPath = path
			}
		}
	}

	cmd := exec.Command(cmdPath, cmdArgs...)
	cmd.Dir = workDir

	// Environment variables
	envMap := make(map[string]string)
	for _, kv := range os.Environ() {
		// populate host env
		envMap[kv] = "" // standard env format
	}

	cmd.Env = os.Environ()
	// Add app-specific environment
	for k, v := range a.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Inject metadata environment variables
	cmd.Env = append(cmd.Env,
		fmt.Sprintf("DOCROID_APP_ID=%s", a.ID),
		fmt.Sprintf("DOCROID_APP_NAME=%s", a.Name),
		fmt.Sprintf("DOCROID_WORKSPACE=%s", cmd.Dir),
	)
	if a.Port > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", a.Port))
	}
	if cfg.User != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("USER=%s", cfg.User), fmt.Sprintf("LOGNAME=%s", cfg.User))
	}

	// Network and inter-container service discovery (link envs & HOSTALIASES)
	netName := a.Network
	if netName == "" {
		netName = "bridge"
	}
	if cfg.Store != nil {
		if netMgr, err := network.NewNetworkManager(cfg.Store.BaseDir()); err == nil {
			_, _ = netMgr.Connect(netName, a.ID, a.Name, a.Port, a.NetworkAliases)
			if hostsPath, err := netMgr.WriteContainerHosts(workDir, netName, a.ID); err == nil && hostsPath != "" {
				auxHosts := filepath.Join(workDir, "hosts")
				cmd.Env = append(cmd.Env, fmt.Sprintf("HOSTALIASES=%s", auxHosts))
			}
			if linkEnvs, err := netMgr.GenerateLinkEnvs(netName, a.ID); err == nil {
				for k, v := range linkEnvs {
					cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
				}
			}
		}
	}

	// Configure process group attributes for clean signal handling in background/detached mode.
	// In interactive foreground mode, avoid Setpgid: true so the process stays in the foreground process group.
	if !cfg.Interactive {
		setSysProcAttr(cmd)
	}

	return cmd, nil
}

func (e *NativeEngine) Run(ctx context.Context, cfg *RunConfig) (*os.Process, error) {
	cmd, err := e.Prepare(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Detached {
		// Background detached mode: redirect logs to store
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
			return nil, fmt.Errorf("failed to start background process: %w", err)
		}

		// Don't close outFile/errFile immediately; process has inherited the file descriptors
		go func() {
			_ = cmd.Wait()
			_ = outFile.Close()
			_ = errFile.Close()
		}()

		return cmd.Process, nil
	}

	// Foreground interactive mode
	cmd.Stdin = cfg.Stdin
	cmd.Stdout = cfg.Stdout
	cmd.Stderr = cfg.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start process: %w", err)
	}

	return cmd.Process, nil
}
