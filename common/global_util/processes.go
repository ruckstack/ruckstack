package global_util

import (
	"context"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/shirou/gopsutil/v3/process"
	"k8s.io/apimachinery/pkg/util/wait"
	"sync"
	"syscall"
	"time"
)

/**
Attempts to terminate the process. If it does not terminate in the given time, kills it.
*/
func ShutdownProcess(proc *process.Process, terminateTimeout time.Duration, waitGroup *sync.WaitGroup, ctx context.Context) {
	name, err := proc.Name()
	if err != nil {
		ui.VPrintf("Cannot determine name for PID %d: %s", proc.Pid, err)
		name = "UNKNOWN"
	}
	if terminateTimeout == 0 {
		ui.Printf("Stopping %s (PID %d)", name, proc.Pid)

		if err := proc.KillWithContext(ctx); err != nil {
			ui.Printf("Cannot kill %s (PID %d): %s", name, proc.Pid, err)
		}

		return
	}

	if waitGroup != nil {
		waitGroup.Add(1)
	}

	ui.Printf("Gracefully stopping %s (PID %d)...", name, proc.Pid)
	if err := proc.TerminateWithContext(ctx); err != nil {
		ui.Printf("error terminating %s (PID %d): %s", name, proc.Pid, err)
	}

	go func() {
		if waitGroup != nil {
			defer waitGroup.Done()
		}
		err := wait.PollImmediate(time.Second, terminateTimeout, func() (bool, error) {
			err := proc.SendSignal(syscall.Signal(0))

			if err != nil {
				ui.VPrintf("%s (PID %d): %s", name, proc.Pid, err)
				return true, nil
			}

			return false, nil
		})

		if err != nil {
			ui.Printf("Timeout waiting for %s (PID %d) to terminate", name, proc.Pid)
			ui.Printf("Sending SIGKILL to %s (PID %d)", name, proc.Pid)

			if err := proc.KillWithContext(ctx); err != nil {
				ui.Printf("Cannot kill %s (PID %d): %s", name, proc.Pid, err)
			}
		}
	}()
}
