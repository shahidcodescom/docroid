package engine

import (
	"context"
	"io"
	"os"
	"os/exec"

	"docroid/pkg/app"
)

// RunConfig contains parameters required to launch an application
type RunConfig struct {
	App         *app.App
	Store       *app.Store
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	Detached    bool
	Interactive bool
	TTY         bool
	User        string
}

// Engine defines the interface for running applications
type Engine interface {
	Name() app.EngineType
	Prepare(cfg *RunConfig) (*exec.Cmd, error)
	Run(ctx context.Context, cfg *RunConfig) (*os.Process, error)
}
