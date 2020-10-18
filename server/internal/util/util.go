package util

import (
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/internal/environment"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"os/exec"
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
