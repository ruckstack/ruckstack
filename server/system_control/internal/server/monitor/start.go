package monitor

import (
	"context"
	"fmt"
	"github.com/ruckstack/ruckstack/common/ui"
	"github.com/ruckstack/ruckstack/server/system_control/internal/environment"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger

var ServerStatus = struct {
	SystemReady bool
}{}

var trackers = []*Tracker{}
var trackerUpdateChannel = make(chan *Tracker)

var shutdownInProgress = false
var monitorContext context.Context

func Start(ctx context.Context) error {
	monitorContext = ctx

	ui.Println("Starting monitor...")

	logFile, err := os.OpenFile(filepath.Join(environment.ServerHome, "logs", "monitor.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening webserver.log: %s", err)
	}

	logger = log.New(logFile, "", log.LstdFlags)

	logger.Println("---- Starting monitor ---")

	go func() {
		logger.Printf("HEALTH: System is unhealthy")

		for updatedTracker := range trackerUpdateChannel {
			if shutdownInProgress {
				return
			}

			logger.Printf("Updated tracker %s", updatedTracker.Name)

			saveMonitorStatus()

			foundUnhealthy := false
			for _, tracker := range trackers {
				if !tracker.IsHealthy() {
					foundUnhealthy = true
					break
				}
			}
			nowSystemReady := !foundUnhealthy

			if ServerStatus.SystemReady != nowSystemReady {
				if ServerStatus.SystemReady {
					logger.Printf("HEALTH: System is unhealthy")
					ServerStatus.SystemReady = false
				} else {
					logger.Printf("HEALTH: System is now healthy")
					ServerStatus.SystemReady = true
				}
			}
		}
	}()

	return nil
}

func saveMonitorStatus() {
	saveData := []PersistedTracker{}
	for _, tracker := range trackers {
		saveData = append(saveData, PersistedTracker{
			Name:            tracker.Name,
			CurrentProblems: tracker.currentProblems,
			CurrentWarnings: tracker.currentWarnings,
		})
	}

	monitorStatusFile, err := os.OpenFile(environment.ServerHome+"/logs/monitor.status", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Printf("Error opening monitor.status: %s", err)
		return
	}

	encoder := yaml.NewEncoder(monitorStatusFile)
	if err := encoder.Encode(saveData); err != nil {
		logger.Printf("Error writing monitor.status: %s", err)
	}
	_ = monitorStatusFile.Close()
}

func Add(tracker *Tracker) {
	tracker.Context = monitorContext
	tracker.currentProblems = map[string]string{}
	tracker.seenProblems = map[string]bool{}

	tracker.currentWarnings = map[string]string{}
	tracker.seenWarnings = map[string]bool{}
	tracker.updateChannel = trackerUpdateChannel

	logger.Printf("Starting monitor for %s...", tracker.Name)
	go func() { tracker.Check(tracker) }()

	trackers = append(trackers, tracker)
}

//func watchOverall() {
//	monitorFile := filepath.Join(environment.ServerHome, "logs", "monitor.status")
//
//	//start with saying it's not healthy
//	log.Println("HEALTH: System is not healthy (starting up)")
//
//	for true {
//		monitorStatus := fmt.Sprintf("Monitor status at %s\n", time.Now().Format(time.RubyDate))
//		monitorStatus += "---------------------------------------------\n"
//
//		monitorStatus += "\nWARNINGS:\n"
//		if len(currentWarnings) == 0 {
//			monitorStatus += "No warnings\n"
//		} else {
//			// sort keys
//			var keys []string
//			for k := range currentWarnings {
//				keys = append(keys, k)
//			}
//			sort.Strings(keys)
//
//			for _, key := range keys {
//				monitorStatus += fmt.Sprintf("%s: %s\n", key, currentWarnings[key])
//			}
//		}
//		monitorStatus += "No known problems\n"
//
//		monitorStatus += "\nERRORS:\n"
//		if len(currentProblems) == 0 {
//			monitorStatus += "No known errors\n"
//
//			if !ServerStatus.SystemReady {
//				log.Println("HEALTH: System is healthy")
//			}
//			ServerStatus.SystemReady = true
//		} else {
//			monitorStatus += "UNHEALTHY -- Known problems:\n"
//
//			// sort keys
//			var keys []string
//			for k := range currentProblems {
//				keys = append(keys, k)
//			}
//			sort.Strings(keys)
//
//			for _, key := range keys {
//				monitorStatus += fmt.Sprintf("%s: %s\n", key, currentProblems[key])
//			}
//
//			if ServerStatus.SystemReady {
//				log.Println("HEALTH: System is not healthy")
//			}
//			ServerStatus.SystemReady = false
//		}
//
//		err := ioutil.WriteFile(monitorFile, []byte(monitorStatus), 0644)
//		if err != nil {
//			fmt.Printf("ERROR: %s", err)
//			return
//		}
//
//		time.Sleep(10 * time.Second)
//	}
//}

type PersistedTracker struct {
	Name            string
	CurrentProblems map[string]string
	CurrentWarnings map[string]string
}
