package commands

import (
	"github.com/go-playground/assert/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"testing"
)

func TestFormattingCorrect(t *testing.T) {
	testCommandFormatting(t, RootCmd)
}

func testCommandFormatting(t *testing.T, command *cobra.Command) {
	t.Run("Testing command "+command.Name(), func(t *testing.T) {
		assert.MatchRegex(t, command.Use, "^[a-z-]+$")
		assert.MatchRegex(t, command.Short, "^[A-Z].*")

		if command.Long != "" {
			assert.MatchRegex(t, command.Long, "^[A-Z].*")
		}

		command.Flags().VisitAll(func(flag *pflag.Flag) {
			assert.MatchRegex(t, flag.Name, "^[a-z-]+$")
			assert.MatchRegex(t, flag.Usage, "^[A-Z].*")
		})

		for _, subCommand := range command.Commands() {
			testCommandFormatting(t, subCommand)
		}
	})
}
