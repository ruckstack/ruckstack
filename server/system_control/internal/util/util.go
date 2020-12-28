package util

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"github.com/shirou/gopsutil/v3/process"
	"io/ioutil"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

func ExpectNoError(err error) {
	if err != nil {
		fmt.Printf("Unexpected error %s", err)
		//panic(err)
		os.Exit(15)
	}
}

func ExecBash(bashCommand string) {
	command := exec.Command("bash", "-c", bashCommand)
	command.Dir = environment.ServerHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()
	if err := command.Run(); err != nil {
		panic(err)
	}
}

func GetAbsoluteName(object meta.Object) string {
	return object.GetNamespace() + "/" + object.GetName()
}

/**
Returns the process from the passed file. Returns nil if process is not running (or a zombie)
*/
func GetProcessFromFile(pidFilePath string) (*process.Process, error) {
	pidString, err := ioutil.ReadFile(pidFilePath)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("cannot read %s: %s", pidFilePath, err)
	}

	pid, err := strconv.Atoi(string(pidString))
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s: %s", pidFilePath, err)
	}

	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		ui.VPrintf("PID %d: %s", pid, err)
		return nil, nil
	}

	status, err := proc.Status()
	if err != nil {
		ui.Printf("error checking %d status: %s", pid, err)
	}
	if strings.Contains(strings.Join(status, " "), "zombie") {
		return nil, nil
	}

	return proc, nil
}

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
		ui.Printf("Sending SIGKILL to %s (PID %d)", name, proc.Pid)

		if err := proc.KillWithContext(ctx); err != nil {
			ui.Printf("Cannot kill %s (PID %d): %s", name, proc.Pid, err)
		}

		return
	}

	if waitGroup != nil {
		waitGroup.Add(1)
	}

	ui.Printf("Sending SIGTERM to %s (PID %d)...", name, proc.Pid)
	if err := proc.TerminateWithContext(ctx); err != nil {
		ui.Printf("error killing %s (PID %d): %s", name, proc.Pid, err)
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
