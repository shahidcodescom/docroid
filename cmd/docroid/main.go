package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"docroid/pkg/app"
	"docroid/pkg/bundle"
	"docroid/pkg/cluster"
	"docroid/pkg/compose"
	"docroid/pkg/doctor"
	"docroid/pkg/engine"
	"docroid/pkg/metrics"
	"docroid/pkg/names"
	"docroid/pkg/netproxy"
	"docroid/pkg/network"
	"docroid/pkg/rolling"
	"docroid/pkg/rootfs"
	"docroid/pkg/secrets"
	"docroid/pkg/staticsite"
	"docroid/pkg/supervisor"
	"docroid/pkg/tunnel"
	"docroid/pkg/utils"
	"docroid/pkg/web"
)

const (
	Version = "1.0.0"
	Banner  = `
  ____             ____       _     _ 
 |  _ \  ___   ___|  _ \ ___ (_) __| |
 | | | |/ _ \ / __| |_) / _ \| |/ _  |
 | |_| | (_) | (__|  _ < (_) | | (_| |
 |____/ \___/ \___|_| \_\___/|_|\__,_|
 Lightweight Mobile Container Engine for Android (Termux)
`
)

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	subcommand := os.Args[1]

	store, err := app.NewStore("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing store: %v\n", err)
		os.Exit(1)
	}
	manager := app.NewAppManager(store)
	clusterMgr, _ := cluster.NewManager(store, manager)
	if clusterMgr != nil {
		defer clusterMgr.Stop()
	}

	tunnelStore, _ := tunnel.NewStore(store.BaseDir())
	tunnelMgr := tunnel.NewManager(tunnelStore)

	// Register stop hook to cleanly terminate in-process and background supervisors/proxies
	app.RegisterStopHook(func(appID string) {
		supervisor.GetSupervisor().Cancel(appID)
		netproxy.GetProxyManager().StopProxies(appID)
	})

	switch subcommand {
	case "run":
		handleRun(manager, os.Args[2:])
	case "ps":
		handlePs(manager, os.Args[2:])
	case "stop":
		handleStop(manager, os.Args[2:])
	case "start":
		handleStart(manager, os.Args[2:])
	case "restart":
		handleRestart(manager, os.Args[2:])
	case "update":
		handleUpdate(manager, store, os.Args[2:])
	case "compose":
		handleCompose(manager, store, os.Args[2:])
	case "rm":
		handleRm(manager, os.Args[2:])
	case "logs":
		handleLogs(manager, os.Args[2:])
	case "top":
		handleTop(manager, os.Args[2:])
	case "exec":
		handleExec(manager, os.Args[2:])
	case "secret", "secrets":
		handleSecret(store, os.Args[2:])
	case "network", "net":
		handleNetwork(store, os.Args[2:])
	case "tunnel", "tunnels", "cf-tunnel", "cloudflare":
		handleTunnel(tunnelMgr, manager, store, os.Args[2:])
	case "_supervise":
		handleSupervise(manager, os.Args[2:])
	case "build":
		handleBuild(store, os.Args[2:])
	case "pull":
		handlePull(store, os.Args[2:])
	case "push":
		handlePush(store, os.Args[2:])
	case "export", "save":
		handleExport(store, os.Args[2:])
	case "import", "load":
		handleImport(store, os.Args[2:])
	case "images":
		handleImages(store, os.Args[2:])
	case "rmi":
		handleRmi(store, os.Args[2:])
	case "rootfs":
		handleRootfs(store, os.Args[2:])
	case "static", "site":
		handleStatic(store, os.Args[2:])
	case "_static-serve":
		handleStaticServe(store, os.Args[2:])
	case "cluster":
		handleCluster(clusterMgr, os.Args[2:])
	case "node", "nodes":
		handleNode(clusterMgr, os.Args[2:])
	case "pod", "pods":
		handlePod(clusterMgr, os.Args[2:])
	case "scale":
		handleScale(clusterMgr, os.Args[2:])
	case "deployment", "deployments":
		handleDeployment(clusterMgr, os.Args[2:])
	case "ui", "dashboard":
		handleUI(manager, store, os.Args[2:])
	case "doctor":
		doctor.RunDiagnostics(store, os.Stdout)
	case "version", "-v", "--version":
		fmt.Printf("Docroid version %s (%s/%s)\n", Version, os.Getenv("GOOS"), os.Getenv("GOARCH"))
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command '%s'. Run 'docroid help' for usage.\n", subcommand)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(Banner)
	fmt.Println()
	fmt.Println("Usage: docroid <command> [arguments...]")
	fmt.Println()
	fmt.Println("Application Lifecycle Commands:")
	fmt.Println("  run       Run an application, image, or container (foreground or -d background)")
	fmt.Println("            Flags: -d (detached), -p [host:]port[:cont], -v host:cont,")
	fmt.Println("                   --restart no|always|on-failure, --env-file <file>, --secret <KEY>")
	fmt.Println("  ps        List running or all applications with real-time status")
	fmt.Println("  stop      Gracefully stop a running application")
	fmt.Println("  start     Restart a stopped application")
	fmt.Println("  restart   Restart an application (use -z / --zero-downtime for hot cutover)")
	fmt.Println("  update    Perform zero-downtime rolling update (docroid update --cmd/--image/--env app)")
	fmt.Println("  compose   Manage multi-service stacks (docroid compose up, down, ps, config)")
	fmt.Println("  rm        Remove an application, its logs, and its workspace")
	fmt.Println("  logs      Display or follow application stdout/stderr logs")
	fmt.Println("  top       Display live CPU and memory metrics for an application")
	fmt.Println("  exec      Run a command inside an application's workspace")
	fmt.Println()
	fmt.Println("Distributed Cluster & Pods Scaling (Kubernetes-like):")
	fmt.Println("  cluster   Manage multi-device cluster mesh (init, join, status, leave)")
	fmt.Println("  node      Inspect or manage interconnected device nodes (ls, inspect, rm)")
	fmt.Println("  pod       Manage atomic cluster pods (run, ls, logs, rm)")
	fmt.Println("  scale     Horizontally scale pod replicas across devices (docroid scale <app> <N>)")
	fmt.Println("  deployment Manage scalable multi-device deployments (ls, rm)")
	fmt.Println()
	fmt.Println("Configuration & Security:")
	fmt.Println("  secret    Manage secure key-values / secrets store (set, get, list, rm)")
	fmt.Println()
	fmt.Println("Networking & Inter-Container Communication:")
	fmt.Println("  network   Manage virtual container networks (create, ls, inspect, rm, connect, disconnect)")
	fmt.Println()
	fmt.Println("Cloudflare Tunnels (Remote Port Access & Exposure):")
	fmt.Println("  tunnel    Expose any local port or container to the world via Cloudflare Tunnel")
	fmt.Println("            (docroid tunnel start, ls, stop, restart, rm, logs, status)")
	fmt.Println()
	fmt.Println("Image & Packaging Commands (Portability):")
	fmt.Println("  build     Build a portable image from a Docroidfile or Dockerfile (docroid build -t myapp .)")
	fmt.Println("  pull      Pull an image from an OCI/Docker registry or remote URL (docroid pull alpine:latest)")
	fmt.Println("  push      Push an image to an OCI/Docker registry (docroid push myorg/myapp:latest)")
	fmt.Println("  export    Export an image or app into a portable .dpack file (docroid export myapp)")
	fmt.Println("  import    Import a .dpack bundle into local image store (docroid import -i myapp.dpack)")
	fmt.Println("  images    List installed images / bundles")
	fmt.Println("  rmi       Remove an installed image")
	fmt.Println()
	fmt.Println("Web Dashboard & System Commands:")
	fmt.Println("  ui        Launch the built-in web management dashboard (default :9090)")
	fmt.Println("  static    Host static HTML/assets or ZIP bundles with port and domain routing")
	fmt.Println("            (docroid static serve, ps, start, stop, rm, files, export)")
	fmt.Println("  rootfs    Manage lightweight Linux rootfs environments (list, pull, rm)")
	fmt.Println("  doctor    Inspect Termux, Android kernel, and sandbox environment")
	fmt.Println("  version   Show Docroid client version")
	fmt.Println()
	fmt.Println("Run 'docroid <command> --help' for details on specific commands.")
}

func handleRun(manager *app.AppManager, args []string) {
	args = rearrangeArgs(args)
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	detach := fs.Bool("d", false, "Run in background (detached)")
	name := fs.String("n", "", "Custom application name")
	portStr := fs.String("p", "", "Port mapping (e.g. 8080, 8080:3000, 127.0.0.1:8080:3000)")
	workDir := fs.String("w", "", "Working directory")
	rootfsName := fs.String("rootfs", "", "Run inside lightweight rootfs sandbox (e.g. 'alpine')")
	restartPolicy := fs.String("restart", "no", "Restart policy: no, always, on-failure")
	envFile := fs.String("env-file", "", "Read in a file of environment variables (.env)")

	var envs stringSlice
	fs.Var(&envs, "e", "Set environment variables (KEY=VAL)")

	var secretFlags stringSlice
	fs.Var(&secretFlags, "secret", "Inject secret from secure store (KEY or ENV_KEY=SECRET_KEY)")

	var volumes stringSlice
	fs.Var(&volumes, "v", "Bind mount host directory (host_path:container_path)")

	networkName := fs.String("network", "bridge", "Connect a container to a network (default: bridge)")
	var networkAliases stringSlice
	fs.Var(&networkAliases, "network-alias", "Add network-scoped alias for the container")

	_ = fs.Parse(args)
	remaining := fs.Args()

	if len(remaining) == 0 && *rootfsName == "" {
		fmt.Println("Error: specify a binary or command to run.")
		fmt.Println("Example: docroid run -d -p 8080:3000 ./my-go-server")
		fmt.Println("Example: docroid run -d --restart always -v /data:/app/data myapp")
		fmt.Println("Example: docroid run --rootfs alpine /bin/sh")
		os.Exit(1)
	}

	command := ""
	var cmdArgs []string
	if len(remaining) > 0 {
		command = remaining[0]
		cmdArgs = remaining[1:]
	}

	// Environment parsing: 1) env-file, 2) -e flags, 3) --secret injections
	envMap := make(map[string]string)
	if *envFile != "" {
		fileEnvs, err := secrets.ParseEnvFile(*envFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing env file: %v\n", err)
			os.Exit(1)
		}
		for k, v := range fileEnvs {
			envMap[k] = v
		}
	}
	for _, kv := range envs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}
	if len(secretFlags) > 0 {
		secStore, err := secrets.NewStore(manager.Store().BaseDir())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening secrets store: %v\n", err)
			os.Exit(1)
		}
		for _, sf := range secretFlags {
			envKey := sf
			secKey := sf
			if strings.Contains(sf, "=") {
				parts := strings.SplitN(sf, "=", 2)
				envKey = parts[0]
				secKey = parts[1]
			}
			val, err := secStore.Get(secKey)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: secret '%s' not found: %v\n", secKey, err)
			} else {
				envMap[envKey] = val
			}
		}
	}

	// Volume parsing
	var volMounts []app.VolumeMount
	for _, v := range volumes {
		parts := strings.Split(v, ":")
		if len(parts) >= 2 {
			volMounts = append(volMounts, app.VolumeMount{
				HostPath:      parts[0],
				ContainerPath: parts[1],
			})
		}
	}

	// Port mapping parsing
	var pMap *netproxy.PortMapping
	targetPort := 0
	if *portStr != "" {
		pm, err := netproxy.ParsePortMapping(*portStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid port mapping: %v\n", err)
			os.Exit(1)
		}
		pMap = pm
		if pMap.HostPort == pMap.ContainerPort {
			// On single-network namespace environments (like Android Termux non-root),
			// allocate an ephemeral internal port for the process so the host port
			// is served by Docroid's user-space proxy, enabling zero-downtime hot restarts.
			if freePort, err := utils.GetFreePort(); err == nil {
				oldPortStr := strconv.Itoa(pMap.HostPort)
				pMap.ContainerPort = freePort
				newPortStr := strconv.Itoa(freePort)
				for i, a := range cmdArgs {
					if a == oldPortStr {
						cmdArgs[i] = newPortStr
					}
				}
			}
		}
		targetPort = pMap.ContainerPort
	}

	engineType := app.EngineNative
	if *rootfsName != "" {
		engineType = app.EngineRootfs
	}

	var application *app.App
	var err error

	// Check if target matches an installed image bundle (e.g. docroid run -d myapp)
	img, _ := bundle.GetImage(manager.Store(), command)
	if img == nil && (strings.Contains(command, "/") || strings.Contains(command, ":")) {
		if _, pathErr := exec.LookPath(command); pathErr != nil && !filepath.IsAbs(command) && !strings.HasPrefix(command, "./") {
			fmt.Printf("Unable to find image '%s' locally. Pulling...\n", command)
			if _, pErr := bundle.PullImage(manager.Store(), command, os.Stdout); pErr == nil {
				img, _ = bundle.GetImage(manager.Store(), command)
			}
		}
	}

	if img != nil {
		application, err = bundle.DeployFromImage(manager.Store(), manager, command, *name, *detach)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deploying from image: %v\n", err)
			os.Exit(1)
		}
		if targetPort > 0 {
			application.Port = targetPort
		}
		if *restartPolicy != "no" {
			application.RestartPolicy = *restartPolicy
		}
		for k, v := range envMap {
			application.Env[k] = v
		}
		if len(cmdArgs) > 0 {
			application.Args = append(application.Args, cmdArgs...)
		}
		if len(volMounts) > 0 {
			application.Volumes = append(application.Volumes, volMounts...)
		}
		engineType = application.Engine
		_ = manager.Store().SaveApp(application)
	} else {
		application, err = manager.CreateApp(
			*name,
			engineType,
			command,
			cmdArgs,
			envMap,
			*workDir,
			targetPort,
			volMounts,
			*rootfsName,
			*restartPolicy,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating application: %v\n", err)
			os.Exit(1)
		}
	}

	if pMap != nil {
		application.PortMapping = pMap.String()
		application.Port = pMap.ContainerPort
		_ = manager.Store().SaveApp(application)
	}

	if *networkName != "" {
		application.Network = *networkName
	}
	if len(networkAliases) > 0 {
		application.NetworkAliases = networkAliases
	}
	_ = manager.Store().SaveApp(application)

	// Choose engine
	var eng engine.Engine
	if engineType == app.EngineRootfs {
		eng = engine.NewRootfsEngine()
	} else {
		eng = engine.NewNativeEngine()
	}

	runCfg := &engine.RunConfig{
		App:         application,
		Store:       manager.Store(),
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		Detached:    *detach,
		Interactive: !*detach,
	}

	ctx := context.Background()
	proc, err := eng.Run(ctx, runCfg)
	if err != nil {
		application.Status = app.StatusFailed
		_ = manager.Store().SaveApp(application)
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}

	application.PID = proc.Pid
	application.Status = app.StatusRunning
	application.StartedAt = time.Now()
	_ = manager.Store().SaveApp(application)

	if *detach {
		fmt.Printf("Started %s (ID: %s, PID: %d)\n", application.Name, application.ID, proc.Pid)
		if application.PortMapping != "" {
			fmt.Printf("Ports: %s\n", application.PortMapping)
		}
		if application.RestartPolicy != "" && application.RestartPolicy != "no" {
			fmt.Printf("Restart policy: %s\n", application.RestartPolicy)
		}
		fmt.Printf("Run 'docroid logs %s' or 'docroid top %s' to monitor.\n", application.Name, application.Name)

		if application.RestartPolicy != "no" || (pMap != nil && pMap.HostPort != pMap.ContainerPort) {
			selfPath, err := os.Executable()
			if err != nil {
				selfPath = os.Args[0]
			}
			supCmd := exec.Command(selfPath, "_supervise", application.ID)
			supCmd.Stdin = nil
			supCmd.Stdout = nil
			supCmd.Stderr = nil
			if err := supCmd.Start(); err == nil {
				_ = os.WriteFile(filepath.Join(manager.Store().AppDir(application.ID), "supervisor.pid"), []byte(strconv.Itoa(supCmd.Process.Pid)), 0644)
				go func() {
					_ = supCmd.Process.Release()
				}()
			}
		}
		return
	}

	// Foreground interactive mode: start proxy if ports differ
	if pMap != nil && pMap.HostPort != pMap.ContainerPort {
		if err := netproxy.GetProxyManager().StartProxy(application.ID, pMap); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to start port proxy: %v\n", err)
		} else {
			defer netproxy.GetProxyManager().StopProxies(application.ID)
		}
	}

	// Wait for process in foreground, handling interrupts
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for sig := range sigChan {
			_ = proc.Signal(sig)
		}
	}()

	state, err := proc.Wait()
	application.Status = app.StatusStopped
	if state != nil {
		application.ExitCode = state.ExitCode()
		if !state.Success() {
			application.Status = app.StatusFailed
		}
	}
	_ = manager.Store().SaveApp(application)
}

func handlePs(manager *app.AppManager, args []string) {
	fs := flag.NewFlagSet("ps", flag.ExitOnError)
	all := fs.Bool("a", false, "Show all applications (default shows running only)")
	_ = fs.Parse(args)

	apps, err := manager.ListAppsWithStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing applications: %v\n", err)
		os.Exit(1)
	}

	w := utils.NewTableWriter(os.Stdout)
	fmt.Fprintln(w, "APP ID\tNAME\tENGINE\tSTATUS\tCPU\tMEM\tPORT\tCOMMAND")

	count := 0
	for _, item := range apps {
		a := item.App
		st := item.Stats

		if !*all && a.Status != app.StatusRunning {
			continue
		}

		statusStr := string(a.Status)
		cpuStr := "-"
		memStr := "-"

		if a.Status == app.StatusRunning && st != nil && st.Alive {
			if a.RestartCount > 0 {
				statusStr = fmt.Sprintf("Up %s (restarts: %d)", utils.FormatTimeAgo(a.StartedAt), a.RestartCount)
			} else {
				statusStr = fmt.Sprintf("Up %s", utils.FormatTimeAgo(a.StartedAt))
			}
			cpuStr = fmt.Sprintf("%.1f%%", st.CPUPercent)
			memStr = utils.FormatBytes(st.MemoryRSS)
		}

		portStr := "-"
		if a.PortMapping != "" {
			portStr = a.PortMapping
		} else if a.Port > 0 {
			portStr = strconv.Itoa(a.Port)
		}

		cmdStr := utils.TruncateString(a.FullCommandString(), 30)

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			a.ID,
			a.Name,
			a.Engine,
			statusStr,
			cpuStr,
			memStr,
			portStr,
			cmdStr,
		)
		count++
	}

	_ = w.Flush()
	if count == 0 {
		if *all {
			fmt.Println("No applications found.")
		} else {
			fmt.Println("No running applications. Use 'docroid ps -a' to view stopped apps.")
		}
	}
}

func handleStop(manager *app.AppManager, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: docroid stop <app_name_or_id>")
		os.Exit(1)
	}
	target := args[0]
	if err := manager.StopApp(target); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop %s: %v\n", target, err)
		os.Exit(1)
	}
	fmt.Printf("Stopped %s\n", target)
}

func handleStart(manager *app.AppManager, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: docroid start <app_name_or_id>")
		os.Exit(1)
	}
	target := args[0]
	a, err := manager.Store().FindApp(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "App not found: %v\n", err)
		os.Exit(1)
	}

	if a.Status == app.StatusRunning {
		fmt.Printf("App '%s' is already running.\n", a.Name)
		return
	}

	var eng engine.Engine
	if a.Engine == app.EngineRootfs {
		eng = engine.NewRootfsEngine()
	} else {
		eng = engine.NewNativeEngine()
	}

	cfg := &engine.RunConfig{
		App:      a,
		Store:    manager.Store(),
		Detached: true,
	}

	proc, err := eng.Run(context.Background(), cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start: %v\n", err)
		os.Exit(1)
	}

	a.PID = proc.Pid
	a.Status = app.StatusRunning
	a.StartedAt = time.Now()
	_ = manager.Store().SaveApp(a)

	fmt.Printf("Started %s (PID: %d)\n", a.Name, proc.Pid)
}

func handleRestart(manager *app.AppManager, args []string) {
	fs := flag.NewFlagSet("restart", flag.ExitOnError)
	zeroDowntime := fs.Bool("z", false, "Perform zero-downtime hot restart")
	fs.BoolVar(zeroDowntime, "zero-downtime", false, "Perform zero-downtime hot restart")
	_ = fs.Parse(rearrangeArgs(args))

	if len(fs.Args()) < 1 {
		fmt.Println("Usage: docroid restart [-z|--zero-downtime] <app_name_or_id>")
		os.Exit(1)
	}

	target := fs.Args()[0]

	if *zeroDowntime {
		fmt.Printf("Conducting zero-downtime restart for '%s'...\n", target)
		res, err := rolling.ExecuteZeroDowntimeRestart(context.Background(), manager.Store(), manager, target, rolling.RolloutOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Zero-downtime restart failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ %s (PID %d -> %d in %v)\n", res.Message, res.OldPID, res.NewPID, res.TotalDuration)
		return
	}

	// Traditional stop and start
	_ = manager.StopApp(target)
	time.Sleep(200 * time.Millisecond)
	handleStart(manager, []string{target})
}

func handleUpdate(manager *app.AppManager, store *app.Store, args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	newCmd := fs.String("cmd", "", "Update binary or command path")
	newImage := fs.String("image", "", "Update from new image tag")
	envFile := fs.String("env-file", "", "Load updated environment variables from .env file")
	drainSec := fs.Float64("drain", 1.5, "Connection drain window in seconds (default 1.5s)")
	timeoutSec := fs.Float64("timeout", 10.0, "Readiness probe timeout in seconds (default 10s)")

	var envs stringSlice
	fs.Var(&envs, "e", "Set/update environment variable (KEY=VAL)")

	var secretFlags stringSlice
	fs.Var(&secretFlags, "secret", "Inject/update secret from secure store (KEY or ENV_KEY=SECRET_KEY)")

	_ = fs.Parse(rearrangeArgs(args))

	if len(fs.Args()) < 1 {
		fmt.Println("Usage: docroid update [flags] <app_name_or_id>")
		fmt.Println()
		fmt.Println("Flags:")
		fmt.Println("  --cmd <cmd>        Update binary or command path")
		fmt.Println("  --image <tag>      Update to a new image bundle version")
		fmt.Println("  -e KEY=VAL         Update/add environment variable")
		fmt.Println("  --env-file <file>  Update environment from .env file")
		fmt.Println("  --secret <KEY>     Inject/update secret from store")
		fmt.Println("  --drain <sec>      Connection drain period in seconds (default 1.5s)")
		fmt.Println("  --timeout <sec>    Readiness probe timeout in seconds (default 10s)")
		os.Exit(1)
	}

	target := fs.Args()[0]

	// Parse new env map
	newEnv := make(map[string]string)
	if *envFile != "" {
		fe, err := secrets.ParseEnvFile(*envFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing env file: %v\n", err)
			os.Exit(1)
		}
		for k, v := range fe {
			newEnv[k] = v
		}
	}
	for _, kv := range envs {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			newEnv[parts[0]] = parts[1]
		}
	}
	if len(secretFlags) > 0 {
		secStore, err := secrets.NewStore(store.BaseDir())
		if err == nil {
			for _, sf := range secretFlags {
				envKey := sf
				secKey := sf
				if strings.Contains(sf, "=") {
					parts := strings.SplitN(sf, "=", 2)
					envKey = parts[0]
					secKey = parts[1]
				}
				val, err := secStore.Get(secKey)
				if err == nil {
					newEnv[envKey] = val
				}
			}
		}
	}

	opts := rolling.RolloutOptions{
		NewCommand:   *newCmd,
		NewImage:     *newImage,
		ProbeTimeout: time.Duration(*timeoutSec * float64(time.Second)),
		DrainTimeout: time.Duration(*drainSec * float64(time.Second)),
	}
	if len(newEnv) > 0 {
		opts.NewEnv = newEnv
	}

	fmt.Printf("Initiating zero-downtime rolling update for '%s'...\n", target)
	res, err := rolling.ExecuteZeroDowntimeRestart(context.Background(), store, manager, target, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Rolling update failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %s (took %v)\n", res.Message, res.TotalDuration)
}

func handleRm(manager *app.AppManager, args []string) {
	fs := flag.NewFlagSet("rm", flag.ExitOnError)
	force := fs.Bool("f", false, "Force removal if running")
	_ = fs.Parse(args)

	if len(fs.Args()) < 1 {
		fmt.Println("Usage: docroid rm [-f] <app_name_or_id>")
		os.Exit(1)
	}

	target := fs.Args()[0]
	if err := manager.RemoveApp(target, *force); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to remove: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Removed %s\n", target)
}

func handleLogs(manager *app.AppManager, args []string) {
	fs := flag.NewFlagSet("logs", flag.ExitOnError)
	follow := fs.Bool("f", false, "Follow log output")
	lines := fs.Int("n", 50, "Number of lines to show from end")
	_ = fs.Parse(args)

	if len(fs.Args()) < 1 {
		fmt.Println("Usage: docroid logs [-f] [-n lines] <app_name_or_id>")
		os.Exit(1)
	}

	target := fs.Args()[0]
	if err := manager.StreamLogs(target, *follow, *lines, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading logs: %v\n", err)
		os.Exit(1)
	}
}

func handleTop(manager *app.AppManager, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: docroid top <app_name_or_id>")
		os.Exit(1)
	}

	target := args[0]
	a, err := manager.Store().FindApp(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "App not found: %v\n", err)
		os.Exit(1)
	}

	if a.Status != app.StatusRunning || a.PID <= 0 {
		fmt.Printf("Application '%s' is not running.\n", a.Name)
		return
	}

	stats, err := metrics.GetProcessStats(a.PID)
	if err != nil || !stats.Alive {
		fmt.Printf("Application '%s' (PID %d) is dead or not accessible.\n", a.Name, a.PID)
		return
	}

	fmt.Printf("Live Metrics for %s (PID %d)\n", a.Name, a.PID)
	fmt.Println("-----------------------------------------")
	fmt.Printf("State:       %s\n", stats.State)
	fmt.Printf("CPU Usage:   %.2f%%\n", stats.CPUPercent)
	fmt.Printf("Memory RSS:  %s\n", utils.FormatBytes(stats.MemoryRSS))
	fmt.Printf("Threads:     %d\n", stats.ThreadCount)
	fmt.Printf("Uptime:      %s\n", utils.FormatTimeAgo(a.StartedAt))
	if a.Port > 0 {
		fmt.Printf("Port:        %d\n", a.Port)
	}
	fmt.Printf("Workspace:   %s\n", manager.Store().WorkspaceDir(a.ID))
}

func printExecUsage() {
	fmt.Println("Usage:  docroid exec [OPTIONS] CONTAINER [COMMAND] [ARG...]")
	fmt.Println()
	fmt.Println("Execute a command inside a running application container (similar to docker exec)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -d, --detach               Detached mode: run command in the background")
	fmt.Println("  -e, --env list             Set environment variables (e.g. -e KEY=VAL)")
	fmt.Println("  -f, --force                Execute even if the container is stopped")
	fmt.Println("  -i, --interactive          Keep STDIN open even if not attached")
	fmt.Println("  -t, --tty                  Allocate a pseudo-TTY")
	fmt.Println("  -u, --user string          Username or UID (format: <name|uid>[:<group|gid>])")
	fmt.Println("  -w, --workdir string       Working directory inside the container")
	fmt.Println("      --privileged           Give extended privileges to the process")
	fmt.Println("  -h, --help                 Print usage")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  docroid exec -it myapp sh")
	fmt.Println("  docroid exec -it myapp")
	fmt.Println("  docroid exec myapp ls -la")
	fmt.Println("  docroid exec -e DEBUG=1 -w /workspace myapp cat app.log")
	fmt.Println("  docroid exec -d myapp python3 background_worker.py")
}

func normalizeExecArgs(args []string) []string {
	var out []string
	boolFlags := map[byte]bool{'i': true, 't': true, 'd': true, 'f': true, 'h': true}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			out = append(out, args[i:]...)
			break
		}
		// If argument is a cluster of short flags like -it, -ti, -itd, expand to -i -t -d
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			allBool := true
			for j := 1; j < len(arg); j++ {
				if !boolFlags[arg[j]] {
					allBool = false
					break
				}
			}
			if allBool {
				for j := 1; j < len(arg); j++ {
					out = append(out, fmt.Sprintf("-%c", arg[j]))
				}
				continue
			}
		}
		out = append(out, arg)
	}
	return out
}

func handleExec(manager *app.AppManager, rawArgs []string) {
	normArgs := normalizeExecArgs(rawArgs)

	fs := flag.NewFlagSet("exec", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var (
		interactive bool
		tty         bool
		itCombined  bool
		tiCombined  bool
		detach      bool
		force       bool
		privileged  bool
		help        bool
		workdir     string
		user        string
		envFlags    stringSlice
	)

	fs.BoolVar(&interactive, "i", false, "Keep STDIN open")
	fs.BoolVar(&interactive, "interactive", false, "Keep STDIN open")
	fs.BoolVar(&tty, "t", false, "Allocate pseudo-TTY")
	fs.BoolVar(&tty, "tty", false, "Allocate pseudo-TTY")
	fs.BoolVar(&itCombined, "it", false, "Interactive TTY session")
	fs.BoolVar(&tiCombined, "ti", false, "Interactive TTY session")
	fs.BoolVar(&detach, "d", false, "Detached mode")
	fs.BoolVar(&detach, "detach", false, "Detached mode")
	fs.BoolVar(&force, "f", false, "Force exec even if stopped")
	fs.BoolVar(&force, "force", false, "Force exec even if stopped")
	fs.BoolVar(&privileged, "privileged", false, "Give extended privileges")
	fs.BoolVar(&help, "h", false, "Print usage")
	fs.BoolVar(&help, "help", false, "Print usage")
	fs.StringVar(&workdir, "w", "", "Working directory")
	fs.StringVar(&workdir, "workdir", "", "Working directory")
	fs.StringVar(&user, "u", "", "Username or UID")
	fs.StringVar(&user, "user", "", "Username or UID")
	fs.Var(&envFlags, "e", "Environment variables")
	fs.Var(&envFlags, "env", "Environment variables")

	if err := fs.Parse(normArgs); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		printExecUsage()
		os.Exit(1)
	}

	if help {
		printExecUsage()
		os.Exit(0)
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		printExecUsage()
		os.Exit(1)
	}

	target := remaining[0]
	cmdAndArgs := remaining[1:]

	a, err := manager.Store().FindApp(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: No such container: %s\n", target)
		os.Exit(1)
	}

	if !force {
		if a.Status != app.StatusRunning {
			fmt.Fprintf(os.Stderr, "Error: Container '%s' is not running (status: %s). Use -f / --force to execute anyway.\n", target, a.Status)
			os.Exit(1)
		}
		if a.PID > 0 {
			if proc, err := os.FindProcess(a.PID); err == nil {
				if err := proc.Signal(syscall.Signal(0)); err != nil {
					fmt.Fprintf(os.Stderr, "Error: Container '%s' process (PID %d) is not running. Use -f / --force to execute anyway.\n", target, a.PID)
					os.Exit(1)
				}
			}
		}
	}

	// Resolve command & args
	var cmdName string
	var cmdArgs []string
	if len(cmdAndArgs) == 0 {
		cmdName = "/bin/sh"
		if a.Engine == app.EngineNative {
			if sh := os.Getenv("SHELL"); sh != "" {
				cmdName = sh
			}
		}
		cmdArgs = []string{}
	} else {
		cmdName = cmdAndArgs[0]
		cmdArgs = cmdAndArgs[1:]
	}

	// Resolve working directory
	targetWorkDir := workdir
	if targetWorkDir == "" {
		if a.WorkDir != "" {
			targetWorkDir = a.WorkDir
		} else if a.Engine == app.EngineRootfs {
			targetWorkDir = "/workspace"
		} else {
			targetWorkDir = manager.Store().WorkspaceDir(a.ID)
		}
	}

	// Environment variables: inherit app environment and overlay exec envs
	execEnvs := make(map[string]string)
	for k, v := range a.Env {
		execEnvs[k] = v
	}
	for _, envStr := range envFlags {
		if strings.Contains(envStr, "=") {
			kv := strings.SplitN(envStr, "=", 2)
			execEnvs[kv[0]] = kv[1]
		} else {
			if hostVal, exists := os.LookupEnv(envStr); exists {
				execEnvs[envStr] = hostVal
			}
		}
	}

	tempApp := &app.App{
		ID:             a.ID,
		Name:           a.Name + "-exec",
		Engine:         a.Engine,
		Command:        cmdName,
		Args:           cmdArgs,
		Env:            execEnvs,
		WorkDir:        targetWorkDir,
		Rootfs:         a.Rootfs,
		Volumes:        a.Volumes,
		Network:        a.Network,
		NetworkAliases: a.NetworkAliases,
		Port:           a.Port,
		PortMapping:    a.PortMapping,
	}

	var eng engine.Engine
	if a.Engine == app.EngineRootfs {
		eng = engine.NewRootfsEngine()
	} else {
		eng = engine.NewNativeEngine()
	}

	isInteractive := interactive || tty || itCombined || tiCombined
	if detach {
		isInteractive = false
	}

	cfg := &engine.RunConfig{
		App:         tempApp,
		Store:       manager.Store(),
		Detached:    detach,
		Interactive: isInteractive,
		TTY:         tty || itCombined || tiCombined,
		User:        user,
	}

	if detach {
		cfg.Stdin = nil
		cfg.Stdout = nil
		cfg.Stderr = nil
	} else {
		if isInteractive {
			cfg.Stdin = os.Stdin
		} else {
			cfg.Stdin = nil
		}
		cfg.Stdout = os.Stdout
		cfg.Stderr = os.Stderr
	}

	proc, err := eng.Run(context.Background(), cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: exec failed: %v\n", err)
		os.Exit(1)
	}

	if detach {
		fmt.Printf("Started detached exec process (PID %d)\n", proc.Pid)
		return
	}

	// Forward termination signals to the child process
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		for sig := range sigChan {
			if proc != nil {
				_ = proc.Signal(sig)
			}
		}
	}()

	state, waitErr := proc.Wait()
	signal.Stop(sigChan)
	close(sigChan)

	if waitErr != nil {
		if exitErr, ok := waitErr.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}
	if state != nil {
		if code := state.ExitCode(); code >= 0 {
			os.Exit(code)
		}
	}
}

func handleRootfs(store *app.Store, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: docroid rootfs [list|pull|rm] [name]")
		os.Exit(1)
	}

	m := rootfs.NewManager(store)
	action := args[0]

	switch action {
	case "list", "ls":
		list, err := m.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Installed Rootfs Environments:")
		if len(list) == 0 {
			fmt.Println("  (none installed yet. Run 'docroid rootfs pull alpine')")
			return
		}
		for _, name := range list {
			fmt.Printf("  • %s (%s)\n", name, store.RootfsDir(name))
		}
	case "pull":
		name := "alpine"
		if len(args) > 1 {
			name = args[1]
		}
		if err := m.Pull(name, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Pull failed: %v\n", err)
			os.Exit(1)
		}
	case "rm", "delete":
		if len(args) < 2 {
			fmt.Println("Usage: docroid rootfs rm <name>")
			os.Exit(1)
		}
		name := args[1]
		if err := m.Remove(name); err != nil {
			fmt.Fprintf(os.Stderr, "Remove failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Removed rootfs '%s'\n", name)
	default:
		fmt.Fprintf(os.Stderr, "Unknown rootfs action '%s'\n", action)
		os.Exit(1)
	}
}

func handleUI(manager *app.AppManager, store *app.Store, args []string) {
	fs := flag.NewFlagSet("ui", flag.ExitOnError)
	port := fs.Int("p", 9090, "Port for web dashboard")
	host := fs.String("host", "0.0.0.0", "Host interface to bind")
	_ = fs.Parse(args)

	rootfsMgr := rootfs.NewManager(store)
	if err := web.StartServer(manager, rootfsMgr, *host, *port); err != nil {
		fmt.Fprintf(os.Stderr, "Dashboard server error: %v\n", err)
		os.Exit(1)
	}
}

func handleBuild(store *app.Store, args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	tag := fs.String("t", "", "Name and optionally a tag in the 'name:tag' format")
	file := fs.String("f", "", "Name of the Docroidfile (Default is 'PATH/Docroidfile')")
	_ = fs.Parse(args)

	contextDir := "."
	if len(fs.Args()) > 0 {
		contextDir = fs.Args()[0]
	}

	builder := bundle.NewBuilder(store)
	_, err := builder.BuildImage(*file, contextDir, *tag, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Build failed: %v\n", err)
		os.Exit(1)
	}
}

func handleExport(store *app.Store, args []string) {
	args = rearrangeArgs(args)
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	outPath := fs.String("o", "", "Output file path (e.g. myapp.dpack)")
	_ = fs.Parse(args)

	if len(fs.Args()) < 1 {
		fmt.Println("Usage: docroid export [-o filename.dpack] <image_or_app_name>")
		os.Exit(1)
	}

	target := fs.Args()[0]

	// Check if it's an installed image first
	if img, _ := bundle.GetImage(store, target); img != nil {
		if err := bundle.ExportImage(store, target, *outPath, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Export image failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Otherwise try exporting an existing app
	if a, _ := store.FindApp(target); a != nil {
		if err := bundle.ExportApp(store, target, *outPath, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Export app failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	fmt.Fprintf(os.Stderr, "Error: '%s' is neither an installed image nor an existing app\n", target)
	os.Exit(1)
}

func handleImport(store *app.Store, args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	inFile := fs.String("i", "", "Input .dpack bundle archive")
	_ = fs.Parse(args)

	filePath := *inFile
	if filePath == "" && len(fs.Args()) > 0 {
		filePath = fs.Args()[0]
	}

	if filePath == "" {
		fmt.Println("Usage: docroid import -i <filename.dpack>")
		os.Exit(1)
	}

	_, err := bundle.ImportPackage(store, filePath, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Import failed: %v\n", err)
		os.Exit(1)
	}
}

func handleImages(store *app.Store, args []string) {
	images, err := bundle.ListImages(store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing images: %v\n", err)
		os.Exit(1)
	}

	w := utils.NewTableWriter(os.Stdout)
	fmt.Fprintln(w, "REPOSITORY\tTAG\tENGINE\tARCH\tCREATED\tCOMMAND")

	for _, img := range images {
		cmdStr := utils.TruncateString(img.Command+" "+strings.Join(img.Args, " "), 30)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			img.Name,
			img.Version,
			img.Engine,
			img.Architecture,
			utils.FormatTimeAgo(img.CreatedAt),
			cmdStr,
		)
	}
	_ = w.Flush()

	if len(images) == 0 {
		fmt.Println("No images found. Build one with 'docroid build -t name:tag .' or import with 'docroid import -i file.dpack'.")
	}
}

func handleRmi(store *app.Store, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: docroid rmi <image_name[:tag]>")
		os.Exit(1)
	}
	tag := args[0]
	if err := bundle.RemoveImage(store, tag); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to remove image: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Untagged and deleted image: %s\n", tag)
}

func handlePull(store *app.Store, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: docroid pull <image[:tag]>")
		fmt.Println("Examples:")
		fmt.Println("  docroid pull alpine:latest")
		fmt.Println("  docroid pull ghcr.io/owner/repo:tag")
		fmt.Println("  docroid pull https://example.com/rootfs.tar.gz")
		os.Exit(1)
	}
	ref := args[0]
	result, err := bundle.PullImage(store, ref, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Pull failed: %v\n", err)
		os.Exit(1)
	}
	if result.Digest != "" {
		fmt.Printf("Successfully pulled %s:%s (Digest: %s)\n", result.Name, result.Tag, result.Digest)
	} else {
		fmt.Printf("Successfully pulled and installed %s:%s\n", result.Name, result.Tag)
	}
}

func handlePush(store *app.Store, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: docroid push <image[:tag]>")
		fmt.Println("Example:")
		fmt.Println("  docroid push myorg/myapp:latest")
		fmt.Println("  docroid push registry.example.com/myapp:v1.0")
		os.Exit(1)
	}
	ref := args[0]
	digest, err := bundle.PushImage(store, ref, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Push failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully pushed image %s (Digest: %s)\n", ref, digest)
}


func rearrangeArgs(args []string) []string {
	var flags []string
	var nonFlags []string
	takesArg := map[string]bool{
		"-o": true, "-i": true, "-t": true, "-f": true,
		"-p": true, "-n": true, "-w": true, "-e": true,
		"-v": true, "--rootfs": true, "--restart": true,
		"--host": true, "--env-file": true, "--secret": true,
		"--cmd": true, "--image": true, "--drain": true, "--timeout": true,
		"--name": true, "--replicas": true, "--service-port": true,
		"--token": true, "--engine": true, "--selector": true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if len(nonFlags) > 0 {
			// Once command/binary is encountered, all subsequent arguments belong to the command
			nonFlags = append(nonFlags, arg)
			continue
		}
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if takesArg[arg] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			nonFlags = append(nonFlags, arg)
		}
	}
	return append(flags, nonFlags...)
}

func rearrangeClusterArgs(args []string) []string {
	var flags []string
	var nonFlags []string
	knownFlags := map[string]bool{
		"-p": true, "--port": true, "-n": true, "--name": true,
		"--replicas": true, "--service-port": true, "--engine": true,
		"--selector": true, "--token": true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if knownFlags[arg] {
			flags = append(flags, arg)
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else if strings.Contains(arg, "=") && (strings.HasPrefix(arg, "--name=") || strings.HasPrefix(arg, "--replicas=") || strings.HasPrefix(arg, "--service-port=") || strings.HasPrefix(arg, "-p=") || strings.HasPrefix(arg, "--port=") || strings.HasPrefix(arg, "--engine=")) {
			flags = append(flags, arg)
		} else {
			nonFlags = append(nonFlags, arg)
		}
	}
	return append(flags, nonFlags...)
}

func handleSecret(store *app.Store, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: docroid secret <command> [arguments...]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  set <KEY> <VALUE>  Store a secret value securely (0600 permissions)")
		fmt.Println("  get <KEY>          Retrieve a stored secret value")
		fmt.Println("  list, ls           List all stored secret keys")
		fmt.Println("  rm, delete <KEY>   Delete a stored secret")
		return
	}

	secStore, err := secrets.NewStore(store.BaseDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing secrets store: %v\n", err)
		os.Exit(1)
	}

	switch args[0] {
	case "set":
		if len(args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: docroid secret set <KEY> <VALUE>\n")
			os.Exit(1)
		}
		key := args[1]
		val := strings.Join(args[2:], " ")
		if err := secStore.Set(key, val); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving secret: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Secret '%s' stored successfully.\n", key)

	case "get":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: docroid secret get <KEY>\n")
			os.Exit(1)
		}
		key := args[1]
		val, err := secStore.Get(key)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(val)

	case "list", "ls":
		items, err := secStore.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing secrets: %v\n", err)
			os.Exit(1)
		}
		if len(items) == 0 {
			fmt.Println("No secrets stored.")
			return
		}
		w := utils.NewTableWriter(os.Stdout)
		fmt.Fprintln(w, "SECRET KEY\tVALUE (MASKED)")
		for k, v := range items {
			masked := strings.Repeat("•", len(v))
			if len(masked) > 12 {
				masked = strings.Repeat("•", 12)
			}
			if len(masked) == 0 {
				masked = "(empty)"
			}
			fmt.Fprintf(w, "%s\t%s\n", k, masked)
		}
		_ = w.Flush()

	case "rm", "delete":
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: docroid secret rm <KEY>\n")
			os.Exit(1)
		}
		key := args[1]
		if err := secStore.Delete(key); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting secret: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Secret '%s' deleted.\n", key)

	default:
		fmt.Fprintf(os.Stderr, "Unknown secret command '%s'. Run 'docroid secret' for usage.\n", args[0])
		os.Exit(1)
	}
}

func handleCompose(manager *app.AppManager, store *app.Store, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: docroid compose <up|down|ps|config|generate> [flags]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  up        Create and start all services defined in docroid-compose.yml")
		fmt.Println("  down      Stop and remove all services defined in docroid-compose.yml")
		fmt.Println("  ps        List status of compose services")
		fmt.Println("  config    Validate and view the resolved compose configuration")
		fmt.Println("  generate  Generate a docroid-compose.yml for an existing application")
		os.Exit(1)
	}

	sub := args[0]
	subArgs := args[1:]

	findComposeFile := func(specified string) (string, error) {
		if specified != "" {
			if _, err := os.Stat(specified); err == nil {
				return specified, nil
			}
			return "", fmt.Errorf("compose file not found: %s", specified)
		}
		candidates := []string{"docroid-compose.yml", "docroid-compose.yaml", "docker-compose.yml", "docker-compose.yaml"}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
		return "", fmt.Errorf("no docroid-compose.yml or docker-compose.yml found in current directory")
	}

	switch sub {
	case "up":
		fs := flag.NewFlagSet("compose up", flag.ExitOnError)
		file := fs.String("f", "", "Path to compose file (default: docroid-compose.yml)")
		detach := fs.Bool("d", true, "Run containers in the background")
		_ = fs.Parse(subArgs)

		path, err := findComposeFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			os.Exit(1)
		}

		cfg, err := compose.ParseYAML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing compose file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Deploying Docroid compose stack from '%s' (%d services)...\n", path, len(cfg.Services))
		launched, err := compose.Up(context.Background(), store, manager, cfg, *detach)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Compose up error: %v\n", err)
			os.Exit(1)
		}

		for _, a := range launched {
			portInfo := a.PortMapping
			if portInfo == "" && a.Port > 0 {
				portInfo = strconv.Itoa(a.Port)
			}
			if portInfo == "" {
				portInfo = "-"
			}
			fmt.Printf("✓ Service '%s' started (ID: %s, PID: %d, Port: %s)\n", a.Name, a.ID, a.PID, portInfo)
		}
		fmt.Println("All services running. Use 'docroid compose ps' or 'docroid ui' to monitor.")

	case "down":
		fs := flag.NewFlagSet("compose down", flag.ExitOnError)
		file := fs.String("f", "", "Path to compose file (default: docroid-compose.yml)")
		_ = fs.Parse(subArgs)

		path, err := findComposeFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			os.Exit(1)
		}

		cfg, err := compose.ParseYAML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing compose file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Stopping and removing services defined in '%s'...\n", path)
		if err := compose.Down(context.Background(), store, manager, cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping services: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ All compose services stopped and removed.")

	case "ps":
		fs := flag.NewFlagSet("compose ps", flag.ExitOnError)
		file := fs.String("f", "", "Path to compose file (default: docroid-compose.yml)")
		_ = fs.Parse(subArgs)

		path, err := findComposeFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			os.Exit(1)
		}

		cfg, err := compose.ParseYAML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		list, _ := compose.Status(store, cfg)
		if len(list) == 0 {
			fmt.Println("No active services running for this compose stack.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "SERVICE\tAPP ID\tSTATUS\tPID\tPORT\tCOMMAND")
		for _, a := range list {
			pInfo := a.PortMapping
			if pInfo == "" && a.Port > 0 {
				pInfo = strconv.Itoa(a.Port)
			}
			cmd := a.Command + " " + strings.Join(a.Args, " ")
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n", a.Name, a.ID, a.Status, a.PID, pInfo, utils.TruncateString(cmd, 30))
		}
		_ = w.Flush()

	case "config":
		fs := flag.NewFlagSet("compose config", flag.ExitOnError)
		file := fs.String("f", "", "Path to compose file (default: docroid-compose.yml)")
		_ = fs.Parse(subArgs)

		path, err := findComposeFile(*file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			os.Exit(1)
		}

		cfg, err := compose.ParseYAML(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
			os.Exit(1)
		}

		yamlOut, err := compose.ToYAML(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting config: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(yamlOut))

	case "generate":
		if len(subArgs) < 1 {
			fmt.Fprintf(os.Stderr, "Usage: docroid compose generate <app-name-or-id> [-o docroid-compose.yml]\n")
			os.Exit(1)
		}
		target := subArgs[0]
		targetApp, err := store.FindApp(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "App '%s' not found: %v\n", target, err)
			os.Exit(1)
		}

		var vols []string
		for _, v := range targetApp.Volumes {
			vols = append(vols, fmt.Sprintf("%s:%s", v.HostPath, v.ContainerPath))
		}

		yamlBytes, err := compose.GenerateComposeYAML(
			targetApp.Name,
			string(targetApp.Engine),
			targetApp.Rootfs,
			targetApp.Command,
			targetApp.Args,
			targetApp.PortMapping,
			targetApp.RestartPolicy,
			targetApp.WorkDir,
			targetApp.Env,
			vols,
			nil,
			true,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating compose file: %v\n", err)
			os.Exit(1)
		}

		outFile := "docroid-compose.yml"
		if len(subArgs) >= 3 && subArgs[1] == "-o" {
			outFile = subArgs[2]
		}

		if err := os.WriteFile(outFile, yamlBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outFile, err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated '%s' for application '%s'.\n", outFile, targetApp.Name)

	default:
		fmt.Fprintf(os.Stderr, "Unknown compose command '%s'. Run 'docroid compose' for usage.\n", sub)
		os.Exit(1)
	}
}

func handleSupervise(manager *app.AppManager, args []string) {
	if len(args) < 1 {
		return
	}
	appID := args[0]
	application, err := manager.Store().LoadApp(appID)
	if err != nil || application == nil || application.Status == app.StatusStopped {
		return
	}

	// Write supervisor PID for graceful cleanup
	supFile := filepath.Join(manager.Store().AppDir(appID), "supervisor.pid")
	_ = os.WriteFile(supFile, []byte(strconv.Itoa(os.Getpid())), 0644)
	defer os.Remove(supFile)

	// Start port proxy if host port differs from container port
	if application.PortMapping != "" {
		if mapping, err := netproxy.ParsePortMapping(application.PortMapping); err == nil && mapping.HostPort != mapping.ContainerPort {
			_ = netproxy.GetProxyManager().StartProxyWithControl(appID, mapping, manager.Store().AppDir(appID))
			defer netproxy.GetProxyManager().StopProxies(appID)
		}
	}

	// Start supervisor monitoring if restart policy is active
	if application.RestartPolicy != "" && application.RestartPolicy != "no" {
		var eng engine.Engine
		if application.Engine == app.EngineRootfs {
			eng = engine.NewRootfsEngine()
		} else {
			eng = engine.NewNativeEngine()
		}
		supervisor.GetSupervisor().Watch(application, manager.Store(), eng)
		defer supervisor.GetSupervisor().Cancel(appID)
	}

	// Wait for process termination, stop signal, or stopped state
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			return
		case <-ticker.C:
			current, err := manager.Store().LoadApp(appID)
			if err != nil || current == nil || current.Status == app.StatusStopped {
				return
			}
		}
	}
}

// Static Website CLI Handlers

func printStaticUsage() {
	fmt.Println("Usage: docroid static <command> [arguments...]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  serve <path_or_zip>  Host HTML/static files or a ZIP bundle")
	fmt.Println("                       Flags: -p <port> (default 8080), -d <domain> (vhost),")
	fmt.Println("                              -n <name>, --spa, --dir, -b / --detach")
	fmt.Println("  ps, list             List all hosted static websites and their statuses")
	fmt.Println("  start <name|id>      Start a stopped static website server")
	fmt.Println("  stop <name|id>       Stop a running static website server")
	fmt.Println("  rm <name|id>         Remove a static website and delete its files")
	fmt.Println("  files <name|id>      List all files inside a static website's public directory")
	fmt.Println("  export <name|id>     Export website public directory as a .zip bundle")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  docroid static serve ./build -p 8080 -d mysite.local")
	fmt.Println("  docroid static serve site.zip -p 3000 --spa -b")
	fmt.Println("  docroid static list")
}

func handleStatic(store *app.Store, args []string) {
	siteStore, err := staticsite.NewStore(store.BaseDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing static site store: %v\n", err)
		os.Exit(1)
	}

	if len(args) == 0 {
		printStaticUsage()
		return
	}

	sub := args[0]
	subArgs := args[1:]

	switch sub {
	case "help", "-h", "--help":
		printStaticUsage()

	case "ps", "list", "ls":
		sites, err := siteStore.ListSites()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing sites: %v\n", err)
			os.Exit(1)
		}
		if len(sites) == 0 {
			fmt.Println("No static websites currently hosted. Run 'docroid static serve <path>' to deploy one.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "SITE ID\tNAME\tPORT\tDOMAIN\tFILES\tSIZE\tSTATUS\tURL")
		for _, s := range sites {
			staticsite.IsSiteRunning(s, siteStore)
			dom := s.Domain
			if dom == "" {
				dom = "*"
			}
			urlStr := s.URL("localhost")
			statusStr := string(s.Status)
			if s.Status == staticsite.StatusRunning {
				statusStr = fmt.Sprintf("running (pid %d)", s.PID)
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%d\t%s\t%s\t%s\n",
				s.ID,
				s.Name,
				s.Port,
				dom,
				s.FileCount,
				utils.FormatBytes(s.TotalSize),
				statusStr,
				urlStr,
			)
		}
		_ = w.Flush()

	case "serve", "up":
		handleStaticServeCmd(siteStore, subArgs)

	case "start":
		if len(subArgs) == 0 {
			fmt.Println("Usage: docroid static start <name|id>")
			os.Exit(1)
		}
		target := subArgs[0]
		site, err := siteStore.FindSite(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := staticsite.StartSiteProcess(site.ID, siteStore); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start site: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Started static site '%s' on port %d (%s)\n", site.Name, site.Port, site.URL("localhost"))

	case "stop":
		if len(subArgs) == 0 {
			fmt.Println("Usage: docroid static stop <name|id>")
			os.Exit(1)
		}
		target := subArgs[0]
		site, err := siteStore.FindSite(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := staticsite.StopSiteProcess(site.ID, siteStore); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to stop site: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Stopped static site '%s' (port %d)\n", site.Name, site.Port)

	case "rm", "delete":
		if len(subArgs) == 0 {
			fmt.Println("Usage: docroid static rm <name|id>")
			os.Exit(1)
		}
		target := subArgs[0]
		site, err := siteStore.FindSite(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		_ = staticsite.StopSiteProcess(site.ID, siteStore)
		if err := siteStore.DeleteSite(site.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete site: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Deleted static site '%s' and cleaned up files\n", site.Name)

	case "files":
		if len(subArgs) == 0 {
			fmt.Println("Usage: docroid static files <name|id> [subpath]")
			os.Exit(1)
		}
		target := subArgs[0]
		subpath := ""
		if len(subArgs) > 1 {
			subpath = subArgs[1]
		}
		site, err := siteStore.FindSite(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		files, err := siteStore.ListSiteFiles(site.ID, subpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing files: %v\n", err)
			os.Exit(1)
		}
		if len(files) == 0 {
			fmt.Println("No files found.")
			return
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tSIZE\tMODIFIED")
		for _, f := range files {
			typeStr := f.MimeType
			sizeStr := utils.FormatBytes(f.Size)
			if f.IsDir {
				typeStr = "directory"
				sizeStr = "-"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", f.Name, typeStr, sizeStr, f.ModTime.Format("2006-01-02 15:04"))
		}
		_ = w.Flush()

	case "export":
		if len(subArgs) == 0 {
			fmt.Println("Usage: docroid static export <name|id> [-o out.zip]")
			os.Exit(1)
		}
		target := subArgs[0]
		outPath := ""
		for i := 1; i < len(subArgs); i++ {
			if subArgs[i] == "-o" && i+1 < len(subArgs) {
				outPath = subArgs[i+1]
				break
			}
		}
		site, err := siteStore.FindSite(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if outPath == "" {
			outPath = site.Name + ".zip"
		}
		f, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create %s: %v\n", outPath, err)
			os.Exit(1)
		}
		defer f.Close()
		if err := siteStore.ExportSiteZip(site.ID, f); err != nil {
			fmt.Fprintf(os.Stderr, "Export failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Exported site '%s' to '%s'\n", site.Name, outPath)

	default:
		// If argument looks like a path or flag, treat directly as serve
		handleStaticServeCmd(siteStore, args)
	}
}

func rearrangeStaticArgs(args []string) []string {
	var flags []string
	var nonFlags []string
	takesArg := map[string]bool{
		"-p": true, "--port": true,
		"-d": true, "--domain": true,
		"-n": true, "--name": true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if takesArg[arg] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++
				flags = append(flags, args[i])
			}
		} else {
			nonFlags = append(nonFlags, arg)
		}
	}
	return append(flags, nonFlags...)
}

func handleStaticServeCmd(siteStore *staticsite.Store, args []string) {
	args = rearrangeStaticArgs(args)
	fs := flag.NewFlagSet("static serve", flag.ExitOnError)
	port := fs.Int("p", 8080, "Port to listen on")
	fs.IntVar(port, "port", 8080, "Port to listen on")
	domain := fs.String("d", "", "Domain or virtual host (e.g. mysite.local or blank for wildcard)")
	fs.StringVar(domain, "domain", "", "Domain or virtual host")
	name := fs.String("n", "", "Custom site name")
	fs.StringVar(name, "name", "", "Custom site name")
	spa := fs.Bool("spa", false, "Enable SPA fallback (route all missing paths to index.html)")
	dirListing := fs.Bool("dir", false, "Enable directory listing if index.html is absent")
	detach := fs.Bool("b", false, "Run in background (detached)")
	fs.BoolVar(detach, "detach", false, "Run in background (detached)")
	fs.BoolVar(detach, "background", false, "Run in background (detached)")

	_ = fs.Parse(args)
	remaining := fs.Args()

	sourcePath := "."
	if len(remaining) > 0 {
		sourcePath = remaining[0]
	}

	fi, err := os.Stat(sourcePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: path '%s' does not exist: %v\n", sourcePath, err)
		os.Exit(1)
	}

	siteName := *name
	if siteName == "" {
		base := filepath.Base(sourcePath)
		base = strings.TrimSuffix(base, filepath.Ext(base))
		if base == "." || base == "/" {
			base = "site"
		}
		siteName = base
	}

	site, err := siteStore.CreateSite(siteName, *port, *domain, *spa, *dirListing)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create site: %v\n", err)
		os.Exit(1)
	}

	// Copy/extract files
	if fi.IsDir() {
		fileMap := make(map[string][]byte)
		_ = filepath.Walk(sourcePath, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				rel, _ := filepath.Rel(sourcePath, p)
				data, err := os.ReadFile(p)
				if err == nil {
					fileMap[rel] = data
				}
			}
			return nil
		})
		_ = siteStore.SaveUploadedFiles(site.ID, fileMap)
	} else if strings.HasSuffix(strings.ToLower(sourcePath), ".zip") {
		f, err := os.Open(sourcePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to open zip file: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		if err := siteStore.ExtractZipBundle(site.ID, f, fi.Size()); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to extract zip bundle: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Single file (e.g. index.html)
		data, err := os.ReadFile(sourcePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
			os.Exit(1)
		}
		_ = siteStore.SaveRawHTML(site.ID, string(data))
	}

	// Refresh stats
	site, _ = siteStore.LoadSite(site.ID)

	if *detach {
		if err := staticsite.StartSiteProcess(site.ID, siteStore); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start background server: %v\n", err)
			os.Exit(1)
		}
		site, _ = siteStore.LoadSite(site.ID)
		fmt.Printf("✓ Static site '%s' running in background (ID: %s, PID: %d)\n", site.Name, site.ID, site.PID)
		fmt.Printf("  • Local URL:  http://localhost:%d\n", site.Port)
		if site.Domain != "" && site.Domain != "*" {
			fmt.Printf("  • Domain URL: %s\n", site.URL(site.Domain))
		}
		fmt.Printf("  • Public Dir: %s (%d files, %s)\n", site.RootDir, site.FileCount, utils.FormatBytes(site.TotalSize))
		fmt.Printf("Run 'docroid static ps' or 'docroid static stop %s' to manage.\n", site.Name)
		return
	}

	// Foreground mode
	fmt.Printf("Docroid Static Server for '%s'\n", site.Name)
	fmt.Printf("=====================================\n")
	fmt.Printf("• Local URL:   %s\n", site.URL("localhost"))
	if site.Domain != "" && site.Domain != "*" {
		fmt.Printf("• Domain URL:  %s\n", site.URL(site.Domain))
	}
	fmt.Printf("• Files:       %d files (%s)\n", site.FileCount, utils.FormatBytes(site.TotalSize))
	fmt.Printf("• Public root: %s\n", site.RootDir)
	fmt.Println("\nPress Ctrl+C to stop the static web server.")

	if err := staticsite.ServeStandalone(site.ID, siteStore); err != nil {
		fmt.Fprintf(os.Stderr, "Server exited with error: %v\n", err)
		os.Exit(1)
	}
}

func handleStaticServe(store *app.Store, args []string) {
	fs := flag.NewFlagSet("_static-serve", flag.ExitOnError)
	siteID := fs.String("id", "", "Site ID to serve")
	_ = fs.Parse(args)

	if *siteID == "" {
		fmt.Fprintf(os.Stderr, "Error: missing --id flag\n")
		os.Exit(1)
	}

	siteStore, err := staticsite.NewStore(store.BaseDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := staticsite.ServeStandalone(*siteID, siteStore); err != nil {
		fmt.Fprintf(os.Stderr, "Error serving site: %v\n", err)
		os.Exit(1)
	}
}

var _ = io.EOF

// --- Distributed Cluster, Nodes & Pods Scaling Commands ---

func handleCluster(mgr *cluster.Manager, args []string) {
	if mgr == nil {
		fmt.Fprintf(os.Stderr, "Error: cluster manager not initialized\n")
		os.Exit(1)
	}

	if len(args) == 0 {
		fmt.Println("Usage: docroid cluster <init|join|status|leave>")
		return
	}

	subcmd := args[0]
	switch subcmd {
	case "init":
		fs := flag.NewFlagSet("cluster init", flag.ExitOnError)
		name := fs.String("name", "docroid-mesh", "Cluster mesh name")
		port := fs.Int("p", 9090, "Control-plane dashboard and API port")
		_ = fs.Parse(args[1:])

		addr := fmt.Sprintf(":%d", *port)
		cfg, err := mgr.InitCluster(*name, addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Cluster initialization failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Docroid Cluster '%s' initialized successfully!\n\n", cfg.ClusterName)
		fmt.Printf("  Role:          %s\n", cfg.Role)
		fmt.Printf("  Control Plane: %s\n", cfg.ControlPlaneAddress)
		fmt.Printf("  Join Token:    %s\n\n", cfg.JoinToken)
		fmt.Printf("To connect another device (Android Termux, Linux, macOS) to this cluster, run:\n")
		fmt.Printf("  docroid cluster join %s --token %s\n\n", cfg.ControlPlaneAddress, cfg.JoinToken)

	case "join":
		if len(args) < 2 {
			fmt.Println("Usage: docroid cluster join <master-addr> --token <token> [--name <node-name>]")
			os.Exit(1)
		}
		leaderAddr := args[1]
		fs := flag.NewFlagSet("cluster join", flag.ExitOnError)
		token := fs.String("token", "", "Cluster join token")
		nodeName := fs.String("name", "", "Custom device node name")
		_ = fs.Parse(args[2:])

		if *token == "" {
			fmt.Fprintf(os.Stderr, "Error: --token is required. Obtain it from 'docroid cluster status' on the control plane.\n")
			os.Exit(1)
		}

		cfg, err := mgr.JoinCluster(leaderAddr, *token, *nodeName, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to join cluster: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Successfully interconnected with cluster '%s'!\n", cfg.ClusterName)
		fmt.Printf("  Role:          Worker Node\n")
		fmt.Printf("  Control Plane: %s\n", cfg.ControlPlaneAddress)
		fmt.Printf("  Local Node ID: %s\n", cfg.LocalNodeID)

	case "status":
		status, err := mgr.GetClusterStatus()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading cluster status: %v\n", err)
			os.Exit(1)
		}

		if !status.Initialized {
			fmt.Println("Docroid node is currently in Standalone mode (not in a cluster).")
			fmt.Println("Run 'docroid cluster init' to start a new cluster, or 'docroid cluster join' to connect to one.")
			return
		}

		fmt.Printf("Cluster Mesh: %s\n", status.ClusterName)
		fmt.Printf("Role:         %s\n", status.Role)
		if status.LocalNode != nil {
			fmt.Printf("Local Node:   %s (%s)\n", status.LocalNode.Name, status.LocalNode.ID)
		}
		if status.Role == cluster.RoleControlPlane {
			fmt.Printf("Control Plane: %s\n", status.LeaderAddr)
			fmt.Printf("Join Token:    %s\n", status.JoinToken)
		} else {
			fmt.Printf("Connected To:  %s\n", status.LeaderAddr)
		}
		fmt.Printf("Nodes:        %d Ready / %d Total\n", status.ReadyNodes, status.TotalNodes)
		fmt.Printf("Pods:         %d Running / %d Total\n", status.RunningPods, status.TotalPods)
		fmt.Printf("Replicas:     %d Desired\n", status.TotalReplicas)

	case "leave":
		if err := mgr.LeaveCluster(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to leave cluster: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ Successfully left cluster mesh. Local pods terminated.")

	default:
		fmt.Fprintf(os.Stderr, "Unknown cluster command '%s'. Run 'docroid cluster' for usage.\n", subcmd)
		os.Exit(1)
	}
}

func handleNode(mgr *cluster.Manager, args []string) {
	if mgr == nil {
		fmt.Fprintf(os.Stderr, "Error: cluster manager not initialized\n")
		os.Exit(1)
	}

	subcmd := "ls"
	subargs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		subcmd = args[0]
		subargs = args[1:]
	}

	switch subcmd {
	case "ls", "list":
		nodes := mgr.ListNodes()
		if len(nodes) == 0 {
			fmt.Println("No interconnected nodes found. Run 'docroid cluster init' to initialize.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tSTATUS\tROLE\tAGE\tVERSION\tOS/ARCH\tPODS\tADDRESS")
		for _, n := range nodes {
			age := utils.FormatDuration(time.Since(n.JoinedAt))
			role := string(n.Role)
			if n.IsLocal {
				role += " (local)"
			}
			osArch := fmt.Sprintf("%s/%s", n.OS, n.Arch)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
				n.Name, n.Status, role, age, n.Version, osArch, n.AllocatedPods, n.Address)
		}
		_ = w.Flush()

	case "inspect":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid node inspect <node-id-or-name>")
			os.Exit(1)
		}
		target := subargs[0]
		nodes := mgr.ListNodes()
		var found *cluster.Node
		for _, n := range nodes {
			if n.ID == target || n.Name == target {
				found = n
				break
			}
		}
		if found == nil {
			fmt.Fprintf(os.Stderr, "Node '%s' not found\n", target)
			os.Exit(1)
		}

		fmt.Printf("Node: %s (%s)\n", found.Name, found.ID)
		fmt.Printf("  Role:           %s\n", found.Role)
		fmt.Printf("  Status:         %s\n", found.Status)
		fmt.Printf("  Address:        %s\n", found.Address)
		fmt.Printf("  OS / Arch:      %s / %s\n", found.OS, found.Arch)
		fmt.Printf("  Capacity:       %d Cores, %d MB RAM, Max %d Pods\n",
			found.Capacity.CPUCores, found.Capacity.MemoryMB, found.Capacity.MaxPods)
		fmt.Printf("  Allocated Pods: %d\n", found.AllocatedPods)
		fmt.Printf("  Last Heartbeat: %s ago\n", utils.FormatDuration(time.Since(found.LastHeartbeat)))

	case "rm", "remove":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid node rm <node-id>")
			os.Exit(1)
		}
		nodeID := subargs[0]
		if err := mgr.RemoveNode(nodeID); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove node: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Node '%s' removed from cluster.\n", nodeID)

	default:
		fmt.Fprintf(os.Stderr, "Unknown node command '%s'. Run 'docroid node ls' to view nodes.\n", subcmd)
		os.Exit(1)
	}
}

func handlePod(mgr *cluster.Manager, args []string) {
	if mgr == nil {
		fmt.Fprintf(os.Stderr, "Error: cluster manager not initialized\n")
		os.Exit(1)
	}

	subcmd := "ls"
	subargs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		subcmd = args[0]
		subargs = args[1:]
	}

	switch subcmd {
	case "ls", "list":
		pods := mgr.ListPods()
		if len(pods) == 0 {
			fmt.Println("No active pods in cluster. Run 'docroid pod run <cmd>' to deploy pods.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tPOD ID\tDEPLOYMENT\tDEVICE (NODE)\tSTATUS\tPORT\tRESTARTS\tAGE")
		for _, p := range pods {
			age := utils.FormatDuration(time.Since(p.CreatedAt))
			portStr := "—"
			if p.HostPort > 0 {
				portStr = fmt.Sprintf(":%d", p.HostPort)
			}
			dep := p.Deployment
			if dep == "" {
				dep = "standalone"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
				p.Name, p.ID, dep, p.NodeName, p.Status, portStr, p.RestartCount, age)
		}
		_ = w.Flush()

	case "run":
		args = rearrangeClusterArgs(subargs)
		fs := flag.NewFlagSet("pod run", flag.ExitOnError)
		name := fs.String("name", "", "Pod / Deployment name")
		replicas := fs.Int("replicas", 1, "Number of pod replicas to scale across nodes")
		port := fs.Int("p", 0, "Container / app port")
		servicePort := fs.Int("service-port", 0, "Cluster-wide service port (load balancer)")
		engine := fs.String("engine", "native", "Execution engine (native or rootfs)")
		_ = fs.Parse(args)

		cmdArgs := fs.Args()
		if len(cmdArgs) == 0 {
			fmt.Println("Usage: docroid pod run <command> [args...] [--name <name>] [--replicas N] [-p port] [--service-port SP]")
			os.Exit(1)
		}

		cmdStr := cmdArgs[0]
		var restArgs []string
		if len(cmdArgs) > 1 {
			restArgs = cmdArgs[1:]
		}

		depName := *name
		if depName == "" {
			depName = names.GenerateRandomName()
		}

		dep := &cluster.Deployment{
			Name:        depName,
			Namespace:   "default",
			Replicas:    *replicas,
			ServicePort: *servicePort,
			Template: cluster.PodSpec{
				Name:          depName,
				Command:       cmdStr,
				Args:          restArgs,
				Port:          *port,
				Engine:        *engine,
				RestartPolicy: "always",
			},
		}

		if err := mgr.CreateDeployment(dep); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create deployment: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✓ Deployment '%s' created! Scaling %d replicas across cluster nodes...\n", depName, *replicas)
		if *servicePort > 0 {
			fmt.Printf("  Cluster Service Load Balancer: http://0.0.0.0:%d\n", *servicePort)
		}
		fmt.Printf("Run 'docroid pod ls' to inspect running replicas.\n")

	case "scale":
		handleScale(mgr, subargs)

	case "logs":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid pod logs <pod-id> [-n lines]")
			os.Exit(1)
		}
		podID := subargs[0]
		fs := flag.NewFlagSet("pod logs", flag.ExitOnError)
		lines := fs.Int("n", 100, "Number of lines to show")
		_ = fs.Parse(subargs[1:])

		logs, err := mgr.GetPodLogs(podID, *lines)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch logs: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(logs)

	case "rm", "delete":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid pod rm <pod-id>")
			os.Exit(1)
		}
		podID := subargs[0]
		if err := mgr.DeletePod(podID); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete pod: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Pod '%s' terminated.\n", podID)

	default:
		fmt.Fprintf(os.Stderr, "Unknown pod command '%s'. Run 'docroid pod ls' to view pods.\n", subcmd)
		os.Exit(1)
	}
}

func handleScale(mgr *cluster.Manager, args []string) {
	if mgr == nil {
		fmt.Fprintf(os.Stderr, "Error: cluster manager not initialized\n")
		os.Exit(1)
	}

	if len(args) < 2 {
		fmt.Println("Usage: docroid scale <deployment-name> <replicas>")
		fmt.Println("Example: docroid scale my-api 4")
		os.Exit(1)
	}

	name := args[0]
	replicas, err := strconv.Atoi(args[1])
	if err != nil || replicas < 0 {
		fmt.Fprintf(os.Stderr, "Error: invalid replicas count '%s'\n", args[1])
		os.Exit(1)
	}

	fmt.Printf("Scaling deployment '%s' to %d replicas across cluster nodes...\n", name, replicas)
	if err := mgr.ScaleDeployment(name, replicas); err != nil {
		fmt.Fprintf(os.Stderr, "Scaling failed: %v\n", err)
		os.Exit(1)
	}

	time.Sleep(300 * time.Millisecond)
	dep, _ := mgr.GetDeployment(name)
	avail := 0
	if dep != nil {
		avail = dep.AvailableReplicas
	}
	fmt.Printf("✓ Deployment '%s' scaled successfully (Replicas: %d/%d Ready)\n", name, avail, replicas)
}

func handleDeployment(mgr *cluster.Manager, args []string) {
	if mgr == nil {
		fmt.Fprintf(os.Stderr, "Error: cluster manager not initialized\n")
		os.Exit(1)
	}

	subcmd := "ls"
	subargs := args
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		subcmd = args[0]
		subargs = args[1:]
	}

	switch subcmd {
	case "ls", "list":
		deps := mgr.ListDeployments()
		if len(deps) == 0 {
			fmt.Println("No deployments found. Run 'docroid pod run' to create a deployment.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NAME\tREPLICAS\tSERVICE PORT\tCOMMAND\tAGE")
		for _, d := range deps {
			age := utils.FormatDuration(time.Since(d.CreatedAt))
			svcPort := "—"
			if d.ServicePort > 0 {
				svcPort = fmt.Sprintf(":%d", d.ServicePort)
			}
			repStr := fmt.Sprintf("%d/%d", d.AvailableReplicas, d.Replicas)
			cmdStr := d.Template.Command + " " + strings.Join(d.Template.Args, " ")
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", d.Name, repStr, svcPort, cmdStr, age)
		}
		_ = w.Flush()

	case "rm", "delete":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid deployment rm <deployment-name>")
			os.Exit(1)
		}
		name := subargs[0]
		if err := mgr.DeleteDeployment(name); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to delete deployment: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Deployment '%s' and all its pods deleted.\n", name)

	default:
		fmt.Fprintf(os.Stderr, "Unknown deployment command '%s'. Run 'docroid deployment ls' for list.\n", subcmd)
		os.Exit(1)
	}
}

func handleNetwork(store *app.Store, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: docroid network <command> [arguments...]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  create      Create a virtual network (docroid network create <name>)")
		fmt.Println("  ls          List virtual networks")
		fmt.Println("  inspect     Display detailed network information")
		fmt.Println("  rm          Remove one or more networks")
		fmt.Println("  connect     Connect a container to a network")
		fmt.Println("  disconnect  Disconnect a container from a network")
		os.Exit(1)
	}

	mgr, err := network.NewNetworkManager(store.BaseDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing network manager: %v\n", err)
		os.Exit(1)
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "create":
		fs := flag.NewFlagSet("network create", flag.ExitOnError)
		driver := fs.String("d", "bridge", "Driver to manage the Network (bridge, host)")
		driverLong := fs.String("driver", "bridge", "Driver to manage the Network (bridge, host)")
		subnet := fs.String("subnet", "", "Subnet in CIDR format (e.g. 10.90.0.0/16)")
		_ = fs.Parse(subargs)
		rem := fs.Args()
		if len(rem) == 0 {
			fmt.Println("Error: network name is required. Usage: docroid network create <name>")
			os.Exit(1)
		}
		name := rem[0]
		d := *driver
		if *driverLong != "bridge" {
			d = *driverLong
		}
		net, err := mgr.CreateNetwork(name, network.DriverType(d), *subnet, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating network: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(net.ID)

	case "ls", "list":
		nets, err := mgr.ListNetworks()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing networks: %v\n", err)
			os.Exit(1)
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "NETWORK ID\tNAME\tDRIVER\tSCOPE\tSUBNET\tCONTAINERS")
		for _, n := range nets {
			shortID := n.ID
			if len(shortID) > 12 {
				shortID = shortID[:12]
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\n", shortID, n.Name, n.Driver, n.Scope, n.Subnet, len(n.Containers))
		}
		_ = w.Flush()

	case "inspect":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid network inspect <network-name-or-id>")
			os.Exit(1)
		}
		net, err := mgr.GetNetwork(subargs[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		data, _ := json.MarshalIndent([]*network.Network{net}, "", "  ")
		fmt.Println(string(data))

	case "rm", "delete":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid network rm <network-name-or-id>...")
			os.Exit(1)
		}
		for _, netName := range subargs {
			if err := mgr.RemoveNetwork(netName); err != nil {
				fmt.Fprintf(os.Stderr, "Error removing network '%s': %v\n", netName, err)
			} else {
				fmt.Println(netName)
			}
		}

	case "connect":
		fs := flag.NewFlagSet("network connect", flag.ExitOnError)
		var aliases stringSlice
		fs.Var(&aliases, "alias", "Add network-scoped alias for the container")
		_ = fs.Parse(subargs)
		rem := fs.Args()
		if len(rem) < 2 {
			fmt.Println("Usage: docroid network connect [--alias <alias>] <network> <container>")
			os.Exit(1)
		}
		netName := rem[0]
		containerNameOrID := rem[1]

		appInst, err := store.FindApp(containerNameOrID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding container '%s': %v\n", containerNameOrID, err)
			os.Exit(1)
		}

		ep, err := mgr.Connect(netName, appInst.ID, appInst.Name, appInst.Port, aliases)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error connecting container to network: %v\n", err)
			os.Exit(1)
		}
		appInst.Network = netName
		for _, a := range aliases {
			if !contains(appInst.NetworkAliases, a) {
				appInst.NetworkAliases = append(appInst.NetworkAliases, a)
			}
		}
		_ = store.SaveApp(appInst)
		fmt.Printf("✓ Connected container '%s' to network '%s' (IP: %s)\n", appInst.Name, netName, ep.IPv4Address)

	case "disconnect":
		if len(subargs) < 2 {
			fmt.Println("Usage: docroid network disconnect <network> <container>")
			os.Exit(1)
		}
		netName := subargs[0]
		containerNameOrID := subargs[1]

		appInst, _ := store.FindApp(containerNameOrID)
		targetID := containerNameOrID
		if appInst != nil {
			targetID = appInst.ID
		}

		if err := mgr.Disconnect(netName, targetID); err != nil {
			fmt.Fprintf(os.Stderr, "Error disconnecting container: %v\n", err)
			os.Exit(1)
		}
		if appInst != nil && appInst.Network == netName {
			appInst.Network = "bridge"
			_ = store.SaveApp(appInst)
		}
		fmt.Printf("✓ Disconnected container '%s' from network '%s'\n", containerNameOrID, netName)

	default:
		fmt.Fprintf(os.Stderr, "Unknown network command '%s'. Run 'docroid network' for usage.\n", subcmd)
		os.Exit(1)
	}
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func handleTunnel(tunnelMgr *tunnel.Manager, manager *app.AppManager, store *app.Store, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: docroid tunnel <command> [arguments...]")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  start     Connect any port or Docroid app through Cloudflare Tunnel (docroid tunnel start -p 8080)")
		fmt.Println("  ls        List all configured tunnels with live status and public URLs")
		fmt.Println("  stop      Stop an active tunnel (docroid tunnel stop <id|name>)")
		fmt.Println("  restart   Restart a tunnel (docroid tunnel restart <id|name>)")
		fmt.Println("  rm        Delete a tunnel configuration (docroid tunnel rm <id|name>)")
		fmt.Println("  logs      View real-time or recent logs for a tunnel (docroid tunnel logs <id|name>)")
		fmt.Println("  status    Inspect system cloudflared binary installation and health")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  docroid tunnel start -p 8080                    # Direct quick tunnel exposing port 8080")
		fmt.Println("  docroid tunnel start --app quote-service -d     # Expose Docroid container in background")
		fmt.Println("  docroid tunnel start -p 3000 --token <CF_TOKEN> # Zero Trust named tunnel")
		fmt.Println("  docroid tunnel ls                               # List active tunnels and public URLs")
		return
	}

	subcmd := args[0]
	subargs := args[1:]

	switch subcmd {
	case "start", "run":
		fs := flag.NewFlagSet("tunnel start", flag.ExitOnError)
		port := fs.Int("p", 0, "Target port to expose (e.g. 8080)")
		target := fs.String("t", "", "Target URL or address (e.g. 8080, 127.0.0.1:8080, http://localhost:3000)")
		appName := fs.String("app", "", "Name or ID of running Docroid app to expose")
		name := fs.String("n", "", "Friendly name for this tunnel")
		token := fs.String("token", "", "Cloudflare Zero Trust Tunnel token")
		hostname := fs.String("hostname", "", "Custom hostname (for Cloudflare DNS)")
		detach := fs.Bool("d", false, "Run tunnel in background (detached)")
		autoRestart := fs.Bool("restart", false, "Automatically restart tunnel if cloudflared crashes")
		_ = fs.Parse(subargs)

		targetPort := *port
		targetStr := *target
		var appID string
		resolvedAppName := *appName

		if resolvedAppName != "" {
			appInst, err := store.FindApp(resolvedAppName)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: app '%s' not found: %v\n", resolvedAppName, err)
				os.Exit(1)
			}
			appID = appInst.ID
			resolvedAppName = appInst.Name
			if targetPort == 0 && targetStr == "" {
				if appInst.Port > 0 {
					targetPort = appInst.Port
				} else if appInst.PortMapping != "" {
					if pm, err := netproxy.ParsePortMapping(appInst.PortMapping); err == nil {
						targetPort = pm.HostPort
					}
				}
			}
		}

		if targetPort == 0 && targetStr == "" && len(fs.Args()) > 0 {
			targetStr = fs.Args()[0]
		}

		mode := tunnel.ModeQuick
		if *token != "" {
			mode = tunnel.ModeToken
		}

		opts := tunnel.TunnelOptions{
			Name:        *name,
			AppID:       appID,
			AppName:     resolvedAppName,
			Target:      targetStr,
			TargetPort:  targetPort,
			Mode:        mode,
			Token:       *token,
			Hostname:    *hostname,
			AutoRestart: *autoRestart,
		}

		tun, err := tunnelMgr.CreateTunnel(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error configuring tunnel: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Starting Cloudflare Tunnel '%s' for target %s...\n", tun.Name, tun.Target)

		tun, err = tunnelMgr.Start(tun.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error starting tunnel: %v\n", err)
			os.Exit(1)
		}

		if *detach {
			fmt.Println("Connecting to Cloudflare edge network in background...")
			foundURL := ""
			for i := 0; i < 12; i++ {
				time.Sleep(500 * time.Millisecond)
				if cur, err := tunnelMgr.Store().GetTunnel(tun.ID); err == nil && cur.PublicURL != "" {
					foundURL = cur.PublicURL
					tun = cur
					break
				}
			}

			fmt.Println()
			fmt.Printf("✓ Tunnel '%s' (ID: %s) is running (PID: %d)\n", tun.Name, tun.ID, tun.PID)
			fmt.Printf("• Local Target: %s\n", tun.Target)
			if foundURL != "" {
				fmt.Printf("• Public URL:   %s\n", foundURL)
			} else if tun.Mode == tunnel.ModeToken {
				fmt.Println("• Cloudflare Zero Trust: Remote route active")
			} else {
				fmt.Printf("• Public URL:   Establishing connection (run 'docroid tunnel ls' in a few seconds)\n")
			}
			fmt.Println()
			fmt.Printf("View logs: docroid tunnel logs %s\n", tun.ID)
			fmt.Printf("Stop:      docroid tunnel stop %s\n", tun.ID)
			return
		}

		// Foreground mode
		fmt.Println("Connecting to Cloudflare edge network...")
		fmt.Println("Press Ctrl+C to disconnect.")
		fmt.Println()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		urlPrinted := false
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-sigChan:
				fmt.Println("\nStopping Cloudflare Tunnel...")
				_, _ = tunnelMgr.Stop(tun.ID)
				fmt.Println("✓ Tunnel disconnected.")
				return
			case <-ticker.C:
				cur, err := tunnelMgr.Store().GetTunnel(tun.ID)
				if err != nil {
					return
				}
				if !urlPrinted && cur.PublicURL != "" {
					urlPrinted = true
					fmt.Println("=================================================================")
					fmt.Printf("✓ PUBLIC CLOUDFLARE URL: %s\n", cur.PublicURL)
					fmt.Println("=================================================================")
					fmt.Printf("Routing directly to: %s\n", cur.Target)
					fmt.Println("Press Ctrl+C to close the tunnel.")
				}
				if cur.Status == tunnel.StatusStopped || cur.Status == tunnel.StatusError {
					if cur.LastError != "" {
						fmt.Fprintf(os.Stderr, "Tunnel exited: %s\n", cur.LastError)
					} else {
						fmt.Println("Tunnel stopped.")
					}
					return
				}
			}
		}

	case "ls", "list":
		tunnels, err := tunnelMgr.Store().ListTunnels()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing tunnels: %v\n", err)
			os.Exit(1)
		}

		if len(tunnels) == 0 {
			fmt.Println("No Cloudflare tunnels configured. Run 'docroid tunnel start -p <port>' or configure in Web UI.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tMODE\tTARGET\tPUBLIC URL\tSTATUS\tPID\tCREATED")
		for _, t := range tunnels {
			pub := t.PublicURL
			if pub == "" {
				if t.Mode == tunnel.ModeToken {
					pub = "(zero trust)"
				} else if t.Status == tunnel.StatusStarting {
					pub = "(connecting...)"
				} else {
					pub = "-"
				}
			}
			pidStr := "-"
			if t.PID > 0 {
				pidStr = strconv.Itoa(t.PID)
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				t.ID,
				t.Name,
				t.Mode,
				t.Target,
				pub,
				t.Status,
				pidStr,
				t.CreatedAt.Format("01/02 15:04"),
			)
		}
		_ = w.Flush()

	case "stop":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid tunnel stop <id|name>")
			os.Exit(1)
		}
		t, err := tunnelMgr.Store().FindTunnel(subargs[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if _, err := tunnelMgr.Stop(t.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Error stopping tunnel: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Tunnel '%s' (%s) stopped.\n", t.Name, t.ID)

	case "restart":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid tunnel restart <id|name>")
			os.Exit(1)
		}
		t, err := tunnelMgr.Store().FindTunnel(subargs[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Restarting tunnel '%s' (%s)...\n", t.Name, t.ID)
		if _, err := tunnelMgr.Restart(t.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Error restarting tunnel: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Tunnel '%s' restarted.\n", t.Name)

	case "rm", "remove", "delete":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid tunnel rm <id|name>")
			os.Exit(1)
		}
		t, err := tunnelMgr.Store().FindTunnel(subargs[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := tunnelMgr.Delete(t.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting tunnel: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Tunnel '%s' (%s) removed.\n", t.Name, t.ID)

	case "logs":
		if len(subargs) == 0 {
			fmt.Println("Usage: docroid tunnel logs <id|name> [-n lines]")
			os.Exit(1)
		}
		fs := flag.NewFlagSet("tunnel logs", flag.ExitOnError)
		lines := fs.Int("n", 50, "Number of lines to show")
		_ = fs.Parse(subargs[1:])

		t, err := tunnelMgr.Store().FindTunnel(subargs[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		logs, err := tunnelMgr.GetLogs(t.ID, *lines)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving logs: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(logs)

	case "status", "check":
		info := tunnel.GetCloudflaredInfo()
		fmt.Println("Cloudflare Tunnel Integration Status:")
		fmt.Println("=====================================")
		if info.Installed {
			fmt.Printf("• Binary:    Installed (%s)\n", info.Path)
			fmt.Printf("• Version:   %s\n", info.Version)
			tunnels, _ := tunnelMgr.Store().ListTunnels()
			active := 0
			for _, tun := range tunnels {
				if tun.Status == tunnel.StatusRunning || tun.Status == tunnel.StatusStarting {
					active++
				}
			}
			fmt.Printf("• Tunnels:   %d total (%d active)\n", len(tunnels), active)
		} else {
			fmt.Println("• Binary:    Not found")
			fmt.Printf("• Note:      %s\n", info.Error)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown tunnel command '%s'. Run 'docroid tunnel' for usage.\n", subcmd)
		os.Exit(1)
	}
}

