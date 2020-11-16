package ui

import (
	"bufio"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

var (
	logger       *log.Logger
	verbose      bool
	inputScanner *bufio.Scanner
)

var NotDirectoryCheck = func(input string) error {
	stat, err := os.Stat(input)
	if os.IsNotExist(err) {
		return nil
	} else {
		if !stat.IsDir() {
			return fmt.Errorf("%s is not a directory", input)
		}
	}
	return nil
}

type InputCheck func(string) error

func init() {
	logger = log.New(os.Stdout, "", 0)

	if os.Getenv("RUCKSTACK_VERBOSE") == "true" {
		SetVerbose(true)
	}

	inputScanner = bufio.NewScanner(os.Stdin)
}

func SetVerbose(value bool) {
	if value {
		logger.SetFlags(log.Ldate | log.Ltime)
		Println("Enabled verbose output")
	} else {
		logger.SetFlags(0)
	}
	verbose = value
}

func IsVerbose() bool {
	return verbose
}

func GetOutput() io.Writer {
	return logger.Writer()
}

func SetOutput(writer io.Writer) {
	logger.SetOutput(writer)
}

func Println(a ...interface{}) {
	logger.Println(a...)
}

/**
Prints the message only if verbose is enabled
*/
func VPrintln(a ...interface{}) {
	if verbose {
		Println(a...)
	}
}

func Printf(format string, a ...interface{}) {
	logger.Printf(format, a...)
}

/**
Prints the message only if verbose is enabled
*/
func VPrintf(format string, a ...interface{}) {
	if verbose {
		Printf(format, a...)
	}
}

/**
Prints the message and exits with an error
*/
func Fatal(v ...interface{}) {
	if verbose {
		debug.PrintStack()
	}

	logger.Fatal(v...)
}

/**
Prints the message and exits with an error
*/
func Fatalf(format string, a ...interface{}) {
	if verbose {
		debug.PrintStack()
	}
	logger.Fatalf(format, a...)
}

func MarkFlagsRequired(command *cobra.Command, requiredFlags ...string) {
	for _, requiredFlag := range requiredFlags {
		if err := command.MarkFlagRequired(requiredFlag); err != nil {
			Fatal(err)
		}
		command.Flag(requiredFlag).Usage = command.Flag(requiredFlag).Usage + " (required)"
	}
}

func MarkFlagsFilename(command *cobra.Command, filenameFlags ...string) {
	for _, requiredFlag := range filenameFlags {
		if err := command.MarkFlagFilename(requiredFlag); err != nil {
			Fatal(err)
		}
	}
}

func MarkFlagsDirname(command *cobra.Command, dirnameFlags ...string) {
	for _, requiredFlag := range dirnameFlags {
		if err := command.MarkFlagDirname(requiredFlag); err != nil {
			Fatal(err)
		}
	}
}

func PromptForString(prompt string, defaultValue string, matchers ...InputCheck) string {
	prompt += ": "

	if defaultValue != "" {
		prompt += "(Default '" + defaultValue + "')"
	}

	Println(prompt)
	inputScanner.Scan()
	input := inputScanner.Text()

	for _, matcher := range matchers {
		err := matcher(input)
		if err != nil {
			Println(err.Error())
			return PromptForString(prompt, defaultValue, matchers...)
		}
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func PromptForBoolean(prompt string, defaultValue *bool) bool {
	yString := "y"
	nString := "n"
	if defaultValue != nil {
		if *defaultValue {
			yString = "Y"
		} else {
			nString = "N"
		}
	}

	Printf("%s: [%s|%s]", prompt, yString, nString)
	inputScanner.Scan()
	input := inputScanner.Text()

	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		if defaultValue == nil {
			return PromptForBoolean(prompt, defaultValue)
		} else {
			return *defaultValue
		}
	}
	if input == "y" {
		return true
	} else if input == "n" {
		return false
	} else {
		Printf("Invalid value '%s'. Enter 'y' or 'n'", input)
		return PromptForBoolean(prompt, defaultValue)
	}
}

func StartProgressMonitor(message string) *spinner.Spinner {
	progressMonitor := spinner.New(spinner.CharSets[11], 100*time.Millisecond, spinner.WithWriter(os.Stderr))
	progressMonitor.Suffix = " " + message + "..."
	progressMonitor.FinalMSG = message + "...DONE\n"
	progressMonitor.Start()

	return progressMonitor
}
