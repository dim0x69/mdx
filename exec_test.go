package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func captureOutput(f func() error, captureStderr bool) (string, error) {
	var buf bytes.Buffer
	stdout := os.Stdout
	stderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	if captureStderr {
		os.Stderr = w
	}

	err := f()

	w.Close()
	io.Copy(&buf, r)
	os.Stdout = stdout
	if captureStderr {
		os.Stderr = stderr
	}

	return buf.String(), err
}

func TestExecuteExecuteCommandBlock_ValidCodeBlockExecution(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}
	commands := make(map[string]CommandBlock)

	commands["test"] = CommandBlock{

		CodeBlocks: []CodeBlock{
			{
				Lang:   "sh",
				Code:   `echo "Hello, {{.arg1}}"`,
				Config: ConfigBlock{SheBang: false},
			},
			{
				Lang:   "sh",
				Code:   `echo -n "Hello"`,
				Config: ConfigBlock{SheBang: false},
			},
		},
		Dependencies: []string{},
		Meta:         map[string]interface{}{},
	}
	args := []string{"World"}
	var wantErr error = nil

	commandBlock := commands["test"]
	output, err := captureOutput(func() error {
		return executeCommandBlock(commands, &commandBlock, args...)
	}, false)

	expectedOutput := "Hello, World\nHello"
	if output != expectedOutput {
		t.Errorf("executeCodeBlock() output = %v, expectedOutput %v", output, expectedOutput)
	}

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteExecuteCommandBlock_ValidCodeBlockExecutionTwoLayersDependencies(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}
	commands := make(map[string]CommandBlock)

	commands["test"] = CommandBlock{
		CodeBlocks: []CodeBlock{
			{
				Lang:   "sh",
				Code:   `echo -n "!"`,
				Config: ConfigBlock{SheBang: false},
			},
		},
		Dependencies: []string{"dep1"},
		Meta:         map[string]interface{}{},
	}
	commands["dep1"] = CommandBlock{
		CodeBlocks: []CodeBlock{
			{
				Lang:   "sh",
				Code:   `echo -n "World"`,
				Config: ConfigBlock{SheBang: false},
			},
		},
		Dependencies: []string{"dep2"},
		Meta:         map[string]interface{}{},
	}
	commands["dep2"] = CommandBlock{
		CodeBlocks: []CodeBlock{
			{
				Lang:   "sh",
				Code:   `echo -n "Hello "`,
				Config: ConfigBlock{SheBang: false},
			},
		},
		Dependencies: []string{},
		Meta:         map[string]interface{}{},
	}
	args := []string{}
	var wantErr error = nil

	commandBlock := commands["test"]
	output, err := captureOutput(func() error {
		return executeCommandBlock(commands, &commandBlock, args...)
	}, false)

	expectedOutput := "Hello World!"
	if output != expectedOutput {
		t.Errorf("executeCodeBlock() output = %v, expectedOutput %v", output, expectedOutput)
	}

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_ValidCodeBlockExecution(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}
	codeBlock := CodeBlock{
		Lang:   "sh",
		Code:   `echo "Hello, {{.arg1}}"`,
		Config: ConfigBlock{SheBang: false},
	}
	args := []string{"World"}
	var wantErr error = nil

	output, err := captureOutput(func() error {
		return executeCodeBlock(&codeBlock, args...)
	}, false)

	expectedOutput := "Hello, World\n"
	if output != expectedOutput {
		t.Errorf("executeCodeBlock() output = %v, expectedOutput %v", output, expectedOutput)
	}

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_ValidCodeBlockExecution_CWD(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}
	codeBlock := CodeBlock{
		Lang:   "sh",
		Code:   `echo "Hello, {{.arg1}}" > file.txt && cat ${PWD}/file.txt && rm file.txt`,
		Config: ConfigBlock{SheBang: false},
	}
	args := []string{"World"}
	var wantErr error = nil

	output, err := captureOutput(func() error {
		return executeCodeBlock(&codeBlock, args...)
	}, false)

	expectedOutput := "Hello, World\n"
	if output != expectedOutput {
		t.Errorf("executeCodeBlock() output = %v, expectedOutput %v", output, expectedOutput)
	}

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_ValidCodeBlockExecution_SheBang(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}
	codeBlock := CodeBlock{
		Lang:   "sh",
		Code:   "#!/bin/sh" + "\n" + `echo "Hello, {{.arg1}}"`,
		Config: ConfigBlock{SheBang: true},
	}
	args := []string{"World"}
	var wantErr error = nil

	output, err := captureOutput(func() error {
		return executeCodeBlock(&codeBlock, args...)
	}, false)

	expectedOutput := "Hello, World\n"
	if output != expectedOutput {
		t.Errorf("executeCodeBlock() output = %v, expectedOutput %v", output, expectedOutput)
	}

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_MissingArgument(t *testing.T) {
	codeBlock := CodeBlock{
		Lang:   "sh",
		Code:   `echo "Hello, {{.arg1}}"`,
		Config: ConfigBlock{SheBang: false},
	}
	args := []string{}
	wantErr := ErrArgUsedInTemplateNotProvided

	err := executeCodeBlock(&codeBlock, args...)

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_UnusedArgument(t *testing.T) {
	codeBlock := CodeBlock{
		Lang:   "sh",
		Code:   `echo "Hello, {{.arg1}}"`,
		Config: ConfigBlock{SheBang: false},
	}
	args := []string{"World", "Extra"}
	wantErr := ErrArgProvidedButNotUsed

	err := executeCodeBlock(&codeBlock, args...)

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_TemplateParsingError(t *testing.T) {
	codeBlock := CodeBlock{
		Lang:   "sh",
		Code:   `echo "Hello, {{.arg1"`, // Missing closing braces
		Config: ConfigBlock{SheBang: false},
	}
	args := []string{"World"}
	wantErr := ErrArgProvidedButNotUsed

	err := executeCodeBlock(&codeBlock, args...)

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_LauncherNotDefined(t *testing.T) {
	codeBlock := CodeBlock{
		Lang:   "unknown",
		Code:   `echo "Hello, {{.arg1}}"`,
		Config: ConfigBlock{SheBang: false},
	}
	args := []string{"World"}
	wantErr := ErrNoLauncherDefined

	err := executeCodeBlock(&codeBlock, args...)

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_DependencyMissing(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}

	args := []string{}
	wantErr := ErrDependencyNotFound
	commands := map[string]CommandBlock{}
	loadCommands("tests/test_dependency_missing.md", commands)
	commandBlock := commands["cmd1"]
	output, err := captureOutput(func() error {
		return executeCommandBlock(commands, &commandBlock, args...)
	}, false)

	// This test would output Hello, if the availability of all deps is not validated before execution.
	expectedOutput := ""
	if output != expectedOutput {
		t.Errorf("executeCodeBlock() output = %v, expectedOutput %v", output, expectedOutput)
	}

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_ConfigFailExecution(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}

	args := []string{}
	var wantErr error = ErrCodeBlockExecFailed
	commands := map[string]CommandBlock{}
	loadCommands("tests/test_config_fail.md", commands)
	commandBlock := commands["cmd"]
	output, err := captureOutput(func() error {
		return executeCommandBlock(commands, &commandBlock, args...)
	}, false)

	// This test would output Hello, if the availability of all deps is not validated before execution.
	expectedOutput := "hello"
	if output != expectedOutput {
		t.Errorf("executeCodeBlock() output = %v, expectedOutput %v", output, expectedOutput)
	}

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestExecuteCodeBlock_ConfigIgnoreError(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}

	args := []string{}
	var wantErr error = nil
	commands := map[string]CommandBlock{}
	loadCommands("tests/test_config_ignore.md", commands)
	commandBlock := commands["cmd"]
	output, err := captureOutput(func() error {
		return executeCommandBlock(commands, &commandBlock, args...)
	}, false)

	// This test would output Hello, if the availability of all deps is not validated before execution.
	expectedOutput := "hello world"
	if output != expectedOutput {
		t.Errorf("executeCodeBlock() output = %v, expectedOutput %v", output, expectedOutput)
	}

	if wantErr != nil {
		if !errors.Is(err, wantErr) {
			t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
		}
	} else if err != nil {
		t.Errorf("executeCodeBlock() error = %v, wantErr %v", err, wantErr)
	}
}

func TestLoadLaunchers(t *testing.T) {
	err := loadLaunchers()
	for _, launcher := range launchers {
		if !strings.HasPrefix(launcher.cmd, "/") {
			t.Errorf("launcher.Cmd = %v, expected to start with '/'", launcher.cmd)
		}
	}

	if _, ok := launchers["sh"]; !ok {
		t.Errorf("Expected launcher for 'sh' to be defined")
	}

	if _, ok := launchers["py"]; !ok {
		t.Errorf("Expected launcher for 'py' to be defined")
	}
	if _, ok := launchers["env"]; !ok {
		t.Errorf("Expected launcher for 'env' to be defined")
	}

	if err != nil {
		t.Errorf("loadLaunchers() error = %v, wantErr %v", err, nil)
	}
}
