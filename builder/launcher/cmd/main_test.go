package main

import (
	"github.com/ruckstack/ruckstack/builder/internal/environment"
	"gotest.tools/assert"
	"testing"
)

func Test_processArguments(t *testing.T) {
	pathToOut := environment.TempPath("path/to/out")
	pathToProject := environment.TempPath("path/to/project")

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
			mountCount: 0,
		},
		{
			name: "Stand along flag",
			args: []string{"--dangling"},
			parsedArgs: map[string]string{
				"--dangling": "",
			},
			newArgs:    []string{"--dangling"},
			env:        []string{},
			mountCount: 0,
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
				"WRAPPED_OUT_ABS=" + pathToOut,
				"WRAPPED_PROJECT=" + pathToProject,
				"WRAPPED_PROJECT_ABS=" + pathToProject,
			},
			mountCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedArgs, args, env, mounts := processArguments(tt.args)
			assert.DeepEqual(t, parsedArgs, tt.parsedArgs)
			assert.DeepEqual(t, args, tt.newArgs)
			assert.DeepEqual(t, env, tt.env)
			assert.Equal(t, tt.mountCount, len(mounts))
		})
	}
}
