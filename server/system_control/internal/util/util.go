package util

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/pkg/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/config"
	"github.com/shirou/gopsutil/v3/process"
	"io/ioutil"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
	command.Dir = config.ServerHome
	command.Stdout = ui.GetOutput()
	command.Stderr = ui.GetOutput()
	if err := command.Run(); err != nil {
		panic(err)
	}
}

func GetAbsoluteName(object meta.Object) string {
	return object.GetNamespace() + "/" + object.GetName()
}

/*
*
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
