package ui

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestPromptForString(t *testing.T) {
	type args struct {
		prompt       string
		defaultValue string
		matchers     []InputCheck
	}
	tests := []struct {
		name         string
		args         args
		input        string
		want         string
		wantedPrompt string
	}{
		{
			name:         "Can get simple string",
			input:        " Passed value  \nAnd later stuff",
			want:         "Passed value",
			wantedPrompt: "Test String: ",
			args: args{
				prompt: "Test String",
			},
		},
		{
			name:  "Asks until value passes",
			input: "UPPER\nlower",
			want:  "lower",
			args: args{
				prompt: "Test String",
				matchers: []InputCheck{
					func(input string) error {
						if strings.ToLower(input) != input {
							return fmt.Errorf("only lower case")
						}
						return nil
					},
				},
			},
		},
		{
			name:         "Takes default",
			input:        "\nAnd later stuff",
			want:         "def_val",
			wantedPrompt: "Test String: (Default 'def_val')",
			args: args{
				prompt:       "Test String",
				defaultValue: "def_val",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := new(bytes.Buffer)
			SetOutput(output)

			inputScanner = bufio.NewScanner(bytes.NewBufferString(tt.input))

			got := PromptForString(tt.args.prompt, tt.args.defaultValue, tt.args.matchers...)
			assert.Equal(t, tt.want, got)
			assert.Contains(t, output.String(), tt.wantedPrompt)
		})
	}
}

func TestPromptForBoolean(t *testing.T) {
	trueValue := true

	type args struct {
		prompt       string
		defaultValue *bool
	}
	var tests = []struct {
		name         string
		args         args
		input        string
		want         bool
		wantedPrompt string
	}{
		{
			name:         "Can get y",
			input:        " y \nAnd later stuff",
			want:         true,
			wantedPrompt: "Test boolean: [y|n]",
			args: args{
				prompt: "Test boolean",
			},
		},
		{
			name:         "Can get N",
			input:        " N \nAnd later stuff",
			want:         false,
			wantedPrompt: "Test boolean: [y|n]",
			args: args{
				prompt: "Test boolean",
			},
		},
		{
			name:         "Default shown",
			input:        " N \nAnd later stuff",
			want:         false,
			wantedPrompt: "Test boolean: [Y|n]",
			args: args{
				prompt:       "Test boolean",
				defaultValue: &trueValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := new(bytes.Buffer)
			SetOutput(output)

			inputScanner = bufio.NewScanner(bytes.NewBufferString(tt.input))

			got := PromptForBoolean(tt.args.prompt, tt.args.defaultValue)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, strings.TrimSpace(output.String()), tt.wantedPrompt)
		})
	}
}
