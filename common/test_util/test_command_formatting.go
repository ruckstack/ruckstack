package test_util

import (
	"github.com/go-playground/assert/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"testing"
)

/**
Ensures that command names and parameters are formatted in a consistent way.
All CLIs should run this against their root command
*/
func RunCommandFormattingTest(t *testing.T, command *cobra.Command) {
	t.Run("Testing command "+command.Name(), func(t *testing.T) {
		assert.MatchRegex(t, command.Short, "^[A-Z].*") //starts with a capital letter

		if command.Long != "" {
			assert.MatchRegex(t, command.Long, "(?m)^[A-Z].*") //starts with a capital letter
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
