package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

func captureOutput(f func() error) (string, error) {
	var buf bytes.Buffer
	stdout := os.Stdout
	stderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	err := f()

	w.Close()
	io.Copy(&buf, r)
	os.Stdout = stdout
	os.Stderr = stderr

	return buf.String(), err
}

func TestExecuteExecuteCommandBlock_ValidCodeBlockExecution(t *testing.T) {
	launchers = map[string]LauncherBlock{"sh": {"sh", "sh"}, "bash": {"sh", "sh"}}
	commands := make(map[string]CommandBlock)

	commands["test"] = CommandBlock{

		CodeBlocks: []CodeBlock{
			{
				Lang: "sh",
				Code: `echo "Hello, {{.arg1}}"`,
				Meta: map[string]interface{}{"shebang": false},
			},
			{
				Lang: "sh",
				Code: `echo -n "Hello"`,
				Meta: map[string]interface{}{"shebang": false},
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
	})

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
				Lang: "sh",
				Code: `echo -n "!"`,
				Meta: map[string]interface{}{"shebang": false},
			},
		},
		Dependencies: []string{"dep1"},
		Meta:         map[string]interface{}{},
	}
	commands["dep1"] = CommandBlock{
		CodeBlocks: []CodeBlock{
			{
				Lang: "sh",
				Code: `echo -n "World"`,
				Meta: map[string]interface{}{"shebang": false},
			},
		},
		Dependencies: []string{"dep2"},
		Meta:         map[string]interface{}{},
	}
	commands["dep2"] = CommandBlock{
		CodeBlocks: []CodeBlock{
			{
				Lang: "sh",
				Code: `echo -n "Hello "`,
				Meta: map[string]interface{}{"shebang": false},
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
	})

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
		Lang: "sh",
		Code: `echo "Hello, {{.arg1}}"`,
		Meta: map[string]interface{}{"shebang": false},
	}
	args := []string{"World"}
	var wantErr error = nil

	output, err := captureOutput(func() error {
		return executeCodeBlock(&codeBlock, args...)
	})

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
		Lang: "sh",
		Code: `echo "Hello, {{.arg1}}" > file.txt && cat ${PWD}/file.txt && rm file.txt`,
		Meta: map[string]interface{}{"shebang": false},
	}
	args := []string{"World"}
	var wantErr error = nil

	output, err := captureOutput(func() error {
		return executeCodeBlock(&codeBlock, args...)
	})

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
		Lang: "sh",
		Code: "#!/bin/sh" + "\n" + `echo "Hello, {{.arg1}}"`,
		Meta: map[string]interface{}{"shebang": true},
	}
	args := []string{"World"}
	var wantErr error = nil

	output, err := captureOutput(func() error {
		return executeCodeBlock(&codeBlock, args...)
	})

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
		Lang: "sh",
		Code: `echo "Hello, {{.arg1}}"`,
		Meta: map[string]interface{}{"shebang": false},
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
		Lang: "sh",
		Code: `echo "Hello, {{.arg1}}"`,
		Meta: map[string]interface{}{"shebang": false},
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
		Lang: "sh",
		Code: `echo "Hello, {{.arg1"`, // Missing closing braces
		Meta: map[string]interface{}{"shebang": false},
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
		Lang: "unknown",
		Code: `echo "Hello, {{.arg1}}"`,
		Meta: map[string]interface{}{"shebang": false},
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
	})

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
