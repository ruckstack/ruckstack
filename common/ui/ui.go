package ui

import (
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
