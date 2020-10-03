package test_util

import (
	"github.com/go-playground/assert/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"strings"
	"testing"
)

/**
Ensures that command names and parameters are formatted in a consistent way.
All CLIs should run this against their root command
*/
func RunCommandFormattingTest(t *testing.T, command *cobra.Command) {
	t.Run("Testing command "+command.Name(), func(t *testing.T) {
		if !strings.HasPrefix(command.Use, "/tmp") {
			//tests for installer rootCommand set the Use by default to be the test name
			assert.MatchRegex(t, command.Use, "^[a-z-]+$")
		}
		assert.MatchRegex(t, command.Short, "^[A-Z].*")

		if command.Long != "" {
			assert.MatchRegex(t, command.Long, "^[A-Z].*")
		}

		command.Flags().VisitAll(func(flag *pflag.Flag) {
			assert.MatchRegex(t, flag.Name, "^[a-z-]+$")
			assert.MatchRegex(t, flag.Usage, "^[A-Z].*")
		})

		for _, subCommand := range command.Commands() {
			RunCommandFormattingTest(t, subCommand)
		}
	})
}
