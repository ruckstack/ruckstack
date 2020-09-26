package main

import (
	"github.com/ruckstack/ruckstack/builder/cli/cmd/commands"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_processArguments(t *testing.T) {
	pathToOut := "/path/to/out"
	pathToProject := "/path/to/project"
	autoMkDirs = false

	tests := []struct {
		name       string
		args       []string
		parsedArgs map[string]string
		newArgs    []string
		env        []string
		mountCount int
	}{
		{
			name:       "Empty Args",
			args:       []string{},
			parsedArgs: map[string]string{},
			newArgs:    []string{},
			env:        []string{},
			mountCount: 1,
		},
		{
			name: "Stand along flag",
			args: []string{"--dangling"},
			parsedArgs: map[string]string{
				"--dangling": "",
			},
			newArgs:    []string{"--dangling"},
			env:        []string{},
			mountCount: 1,
		},
		{
			name:    "Complex processing",
			args:    []string{"--a", "1", "--b", "--out", pathToOut, "--c", "2", "test-command", "--d", "3", "--e", "--project", pathToProject, "--f", "4"},
			newArgs: []string{"--a", "1", "--b", "--out", "/data/out", "--c", "2", "test-command", "--d", "3", "--e", "--project", "/data/project", "--f", "4"},
			parsedArgs: map[string]string{
				"--a":       "1",
				"--b":       "",
				"--out":     pathToOut,
				"--c":       "2",
				"--d":       "3",
				"--e":       "",
				"--project": pathToProject,
				"--f":       "4",
			},
			env: []string{
				"WRAPPED_OUT=" + pathToOut,
				"WRAPPED_PROJECT=" + pathToProject,
			},
			mountCount: 3,
		},
		{
			name: "Handles default values",
			args: []string{"new-project", "--type", "example"},
			parsedArgs: map[string]string{
				"--type": "example",
				"--out":  ".",
			},
			newArgs:    []string{"new-project", "--type", "example", "--out", "/data/out"},
			env:        []string{"WRAPPED_OUT=."},
			mountCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedArgs, args, env, mounts := processArguments(tt.args)
			assert.ElementsMatch(t, parsedArgs, tt.parsedArgs)
			assert.ElementsMatch(t, args, tt.newArgs)
			assert.ElementsMatch(t, env, tt.env)
			assert.Equal(t, tt.mountCount, len(mounts))
		})
	}
}

/**
makes sure the commandDefaults map matches the RootCommand reality
*/
func Test_commandDefaults(t *testing.T) {
	flagsToCheck := []string{"out", "project"}

	rootCommand := commands.RootCmd
	for _, command := range rootCommand.Commands() {
		for _, checkFlag := range flagsToCheck {
			t.Run("Default is correct "+command.Name()+"."+checkFlag, func(t *testing.T) {
				flag := command.Flag(checkFlag)
				mapArgs := commandDefaults[command.Name()]

				if flag == nil {
					if mapArgs != nil {
						_, exists := mapArgs[checkFlag]
						assert.False(t, exists)
					}
				} else {
					if assert.NotNil(t, mapArgs) {
						mapDefault, exists := mapArgs[checkFlag]
						assert.True(t, exists)
						assert.Equal(t, flag.DefValue, mapDefault)
					}
				}
			})
		}
	}

	for commandName, commandArgs := range commandDefaults {
		for _, realCommand := range rootCommand.Commands() {
			if realCommand.Name() != commandName {
				continue
			}

			for commandArg, commandValue := range commandArgs {
				t.Run("No extra info "+commandName+"."+commandArg, func(t *testing.T) {
					realFlag := realCommand.Flag(commandArg)
					assert.NotNil(t, realFlag)
					assert.Equal(t, realFlag.DefValue, commandValue)
				})
			}
		}
	}

}
