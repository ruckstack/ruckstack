package util

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"os"
)

var validate = validator.New()

func Check(err error) {
	if err != nil {
		panic(err)
	}
}

func CheckWithMessage(err error, message string, messageParams ...interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, message, messageParams...)
		fmt.Fprintln(os.Stderr)

		Check(err)
	}
}

func Validate(obj interface{}) error {
	return validate.Struct(obj)
}
