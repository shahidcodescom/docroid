//go:build !windows

package metrics

import (
	"os"
	"syscall"
)

func checkProcessAlive(p *os.Process) error {
	return p.Signal(syscall.Signal(0))
}
