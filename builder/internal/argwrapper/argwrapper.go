package argwrapper

import (
	"fmt"
	"os"
	"strings"
)

/**
Check for a WRAPPED_* environment variable that was set by ruckstack wrapper and return that if it was set.
Otherwise, return the fallbackValue
*/
func GetOriginalValue(key string, fallbackValue string) string {
	env := os.Getenv("WRAPPED_" + strings.ToUpper(key))
	if env == "" {
		return fallbackValue
	} else {
		return env
	}
}

/**
Adds a WRAPPED_* environment variable to the passed array
*/
func SaveOriginalValue(key string, value string, envVariables []string) []string {
	return append(envVariables, fmt.Sprintf("WRAPPED_%s=%s", strings.ToUpper(key), value))
}
