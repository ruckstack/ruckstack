package commands

import (
	"github.com/ruckstack/ruckstack/common/test_util"
	"testing"
)

func TestFormattingCorrect(t *testing.T) {
	test_util.RunCommandFormattingTest(t, rootCmd)
}
