package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/sirupsen/logrus"
)

func isExecutableInPath(candidates []string) string {
	for _, cmd := range candidates {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}

type CodeBlock struct {
	Lang string         // the infostring from the code fence
	Code string         // the content of the code fence
	Meta map[string]any // contains metadata for the code block

}

// CommandBlock represents a heading, which contains one to multiple code fences.
type CommandBlock struct {
	Name         string         // the name of the command, same as the key in the commands map
	Dependencies []string       // commands to execute before this command
	CodeBlocks   []CodeBlock    // the code fences below the heading
	Filename     string         // the filename of the markdown file
	Meta         map[string]any // placeholder for the future
}

// Global store for parsed commands
var commands = map[string]CommandBlock{}

type LauncherBlock struct {
	cmd       string // The command to execute for the infostring above
	extension string // The file extension for the language
}

// global storage for launchers
// the key is the infostring from the code fence
var launchers = map[string]LauncherBlock{}

func loadLaunchers() {
	addedLaunchers := []string{}

	if cmd := isExecutableInPath([]string{"sh"}); cmd != "" {
		launchers["sh"] = LauncherBlock{cmd: cmd, extension: "sh"}
		launchers["bash"] = LauncherBlock{cmd: cmd, extension: "sh"}
		addedLaunchers = append(addedLaunchers, cmd)
	}

	if cmd := isExecutableInPath([]string{"bash"}); cmd != "" {
		launchers["bash"] = LauncherBlock{cmd: cmd, extension: "bash"}
		addedLaunchers = append(addedLaunchers, cmd)
	}

	pythonCandidates := []string{"python", "python3"}
	if cmd := isExecutableInPath(pythonCandidates); cmd != "" {
		launchers["python"] = LauncherBlock{cmd: cmd, extension: "py"}
		addedLaunchers = append(addedLaunchers, cmd)
	}

	logrus.Debug("Added launchers: ", addedLaunchers)
}

func executeCommandBlock(commandBlock *CommandBlock, args ...string) error {
	for _, dep := range commandBlock.Dependencies {
		if _, ok := commands[dep]; !ok {
			return fmt.Errorf("%w: %s", ErrDependencyNotFound, dep)
		}
		dependency := commands[dep]
		if err := executeCommandBlock(&dependency, args...); err != nil {
			logrus.Debug(fmt.Sprintf("Executing command %s with args %v", dependency.Name, args))
			return err
		}
	}

	for i, codeBlock := range commandBlock.CodeBlocks {
		logrus.Debug(fmt.Sprintf("Executing Code Block #%d", i))

		if i == 0 {
			if err := executeCodeBlock(&codeBlock, args...); err != nil {
				return err
			}
		} else {
			if err := executeCodeBlock(&codeBlock); err != nil {
				return err
			}
		}
	}

	return nil

}
func executeCodeBlock(codeBlock *CodeBlock, args ...string) error {

	// Create a map for the template arguments
	argMap := make(map[string]string)
	for i, arg := range args {
		argMap[fmt.Sprintf("arg%d", i+1)] = arg
	}

	// Validate that all placeholders in the template are provided in args and vice versa
	placeholderPattern := regexp.MustCompile(`{{\s*\.arg(\d+)\s*}}`)
	matches := placeholderPattern.FindAllStringSubmatch(codeBlock.Code, -1)

	placeholderSet := make(map[string]struct{})
	for _, match := range matches {
		placeholderSet[match[1]] = struct{}{}
	}

	for i := range args {
		argKey := fmt.Sprintf("%d", i+1)
		if _, ok := placeholderSet[argKey]; !ok {
			return fmt.Errorf("%w: argument: %d (\"%s\")", ErrArgProvidedButNotUsed, i+1, args[i])
		}
	}

	for placeholder := range placeholderSet {
		argIndex := fmt.Sprintf("arg%s", placeholder)
		if _, ok := argMap[argIndex]; !ok {
			return fmt.Errorf("%w: {{.arg%s}}", ErrArgUsedInTemplateNotProvided, placeholder)
		}
	}

	tmpl, err := template.New("command").Parse(codeBlock.Code)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	var renderedCode bytes.Buffer
	err = tmpl.Execute(&renderedCode, argMap)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	launcher, ok := launchers[codeBlock.Lang]
	if !ok {
		return fmt.Errorf("%w: %s", ErrNoLauncherDefined, codeBlock.Lang)
	}

	// Write the rendered code to the temporary file
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("mdx-*.%s", launcher.extension))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	// Set the permissions of the temporary file to 755
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to set permissions on temporary file: %v", err)
	}

	defer os.Remove(tmpFile.Name())

	if !codeBlock.Meta["shebang"].(bool) {
		if _, err := tmpFile.Write([]byte(fmt.Sprintf("#!/usr/bin/env %s\n", launcher.cmd))); err != nil {
			return fmt.Errorf("failed to write to temporary file: %v", err)
		}

	}
	if _, err := tmpFile.Write(renderedCode.Bytes()); err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %v", err)
	}

	cmd := exec.Command(tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Dir = os.Getenv("PWD")
	logrus.Debug(fmt.Sprintf("Executing command in directory: %s", cmd.Dir))

	if err := cmd.Run(); err != nil {
		content, readErr := os.ReadFile(tmpFile.Name())
		if readErr != nil {
			return fmt.Errorf("failed to execute command: %v, and failed to read temporary file: %v", err, readErr)
		}
		fmt.Printf("Content of tmpFile:\n%s\n", content)
		return fmt.Errorf("failed to execute command: %v", err)
	}
	return nil
}
