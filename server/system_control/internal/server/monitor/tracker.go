package monitor

import (
	"strings"
)

type Tracker struct {
	Name  string
	Check func(*Tracker)

	currentProblems map[string]string
	seenProblems    map[string]bool

	currentWarnings map[string]string
	seenWarnings    map[string]bool

	updateChannel chan *Tracker

	receivedData bool
}

func (tracker *Tracker) FoundProblem(component string, problemKey string, description string) {
	tracker.receivedData = true

	problemKey = component + ":" + problemKey

	tracker.seenProblems[problemKey] = true

	existingDesc, problemExists := tracker.currentProblems[problemKey]
	if !problemExists || existingDesc != description {
		message := problemKey
		if description != "" {
			message += " -- " + description
		}
		logger.Println("PROBLEM: " + message)

		tracker.currentProblems[problemKey] = description

		tracker.updateChannel <- tracker
	}

}

func (tracker *Tracker) ResolveProblem(component string, problemKey string, resolvedMessage string) {
	tracker.receivedData = true

	problemKey = component + ":" + problemKey

	_, problemExists := tracker.currentProblems[problemKey]
	if problemExists {
		delete(tracker.currentProblems, problemKey)
		logger.Println("RESOLVED: " + resolvedMessage)

		tracker.updateChannel <- tracker
	} else {
		if !tracker.seenProblems[problemKey] {
			logger.Println("RESOLVED: " + resolvedMessage)
			tracker.seenProblems[problemKey] = true

			tracker.updateChannel <- tracker
		}
	}
}

func (tracker *Tracker) FoundWarning(component string, warningKey string, description string) {
	tracker.receivedData = true

	warningKey = component + ":" + warningKey

	tracker.seenWarnings[warningKey] = true

	existingDesc, warningExists := tracker.currentWarnings[warningKey]
	if !warningExists || existingDesc != description {
		message := warningKey
		if description != "" {
			message += " -- " + description
		}
		logger.Println("WARNING: " + message)
		tracker.currentWarnings[warningKey] = description

		tracker.updateChannel <- tracker
	}

}

func (tracker *Tracker) ResolveWarning(component string, warningKey string, resolvedMessage string) {
	tracker.receivedData = true

	warningKey = component + ":" + warningKey

	_, warningExists := tracker.currentWarnings[warningKey]
	if warningExists {
		delete(tracker.currentWarnings, warningKey)
		logger.Println("RESOLVED: " + resolvedMessage)

		tracker.updateChannel <- tracker
	} else {
		if !tracker.seenWarnings[warningKey] {
			logger.Println("RESOLVED: " + resolvedMessage)

			tracker.seenWarnings[warningKey] = true
			tracker.updateChannel <- tracker
		}
	}
}

func (tracker *Tracker) Log(message string) {
	logger.Println(message)
}

func (tracker *Tracker) Logf(format string, v interface{}) {
	logger.Printf(format, v)
}

func (tracker *Tracker) ResolveComponent(component string) {
	for _, trackerMap := range []map[string]string{tracker.currentProblems, tracker.currentWarnings} {
		for key, _ := range trackerMap {
			if strings.HasPrefix(key, component+":") {
				delete(trackerMap, key)
			}
		}

	}

	for _, trackerMap := range []map[string]bool{tracker.seenProblems, tracker.seenWarnings} {
		for key, _ := range trackerMap {
			if strings.HasPrefix(key, component+":") {
				delete(trackerMap, key)
			}
		}
	}

	tracker.updateChannel <- tracker
}

func (tracker *Tracker) IsHealthy() bool {
	if !tracker.receivedData {
		return false
	}

	return len(tracker.currentProblems) == 0
}
