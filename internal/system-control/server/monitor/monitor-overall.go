package monitor

import (
	"fmt"
	"github.com/ruckstack/ruckstack/internal/system-control/util"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"time"
)

func watchOverall() {
	monitorFile := filepath.Join(util.InstallDir(), "logs", "monitor.status")

	//start with saying it's not healthy
	log.Println("HEALTH: System is not healthy (starting up)")

	for true {
		monitorStatus := fmt.Sprintf("Monitor status at %s\n", time.Now().Format(time.RubyDate))
		monitorStatus += "---------------------------------------------\n"

		monitorStatus += "\nWARNINGS:\n"
		if len(knownWarnings) == 0 {
			monitorStatus += "No warnings\n"
		} else {
			// sort keys
			var keys []string
			for k := range knownWarnings {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				monitorStatus += fmt.Sprintf("%s: %s\n", key, knownWarnings[key])
			}
		}
		monitorStatus += "No known problems\n"

		monitorStatus += "\nERRORS:\n"
		if len(knownProblems) == 0 {
			monitorStatus += "No known errors\n"

			if !ServerStatus.SystemReady {
				log.Println("HEALTH: System is healthy")
			}
			ServerStatus.SystemReady = true
		} else {
			monitorStatus += "UNHEALTHY -- Known problems:\n"

			// sort keys
			var keys []string
			for k := range knownProblems {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				monitorStatus += fmt.Sprintf("%s: %s\n", key, knownProblems[key])
			}

			if ServerStatus.SystemReady {
				log.Println("HEALTH: System is not healthy")
			}
			ServerStatus.SystemReady = false
		}

		err := ioutil.WriteFile(monitorFile, []byte(monitorStatus), 0644)
		util.Check(err)

		time.Sleep(10 * time.Second)
	}
}
