package ui

import (
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"runtime/debug"
)

var (
	logger  *log.Logger
	verbose bool
)

func init() {
	logger = log.New(os.Stdout, "", 0)
}

func SetVerbose(value bool) {
	if value {
		logger.SetFlags(log.Ldate | log.Ltime)
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

func Print(a ...interface{}) {
	logger.Print(a...)
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
