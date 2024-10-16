package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"text/template"

	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func executeCommand(commandName string, args ...string) error {
	logrus.Debug(fmt.Sprintf("Executing command %s with args %v", commandName, args))

	commandBlock, ok := commands[commandName]
	if !ok {
		return fmt.Errorf("command not found: %s", commandName)
	}

	launcher, ok := launchers[commandBlock.Lang]
	if !ok {
		return fmt.Errorf("launcher not found for language: %s", commandBlock.Lang)
	}

	// Create a map for the template arguments
	argMap := make(map[string]string)
	for i, arg := range args {
		argMap[fmt.Sprintf("arg%d", i+1)] = arg
	}

	// Validate that all placeholders in the template are provided in args and vice versa
	placeholderPattern := regexp.MustCompile(`{{\s*\.arg(\d+)\s*}}`)
	matches := placeholderPattern.FindAllStringSubmatch(commandBlock.Code, -1)

	placeholderSet := make(map[string]struct{})
	for _, match := range matches {
		placeholderSet[match[1]] = struct{}{}
	}

	for i := range args {
		argKey := fmt.Sprintf("%d", i+1)
		if _, ok := placeholderSet[argKey]; !ok {
			return fmt.Errorf("argument %d (\"%s\") is provided but not used in the template", i+1, args[i])
		}
	}

	for placeholder := range placeholderSet {
		argIndex := fmt.Sprintf("arg%s", placeholder)
		if _, ok := argMap[argIndex]; !ok {
			return fmt.Errorf("{{.arg%s}} is used in command \"%s\" but not provided in args", placeholder, commandName)
		}
	}

	// Parse and execute the template
	tmpl, err := template.New("command").Parse(commandBlock.Code)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	var renderedCode bytes.Buffer
	err = tmpl.Execute(&renderedCode, argMap)
	if err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Write the rendered code to the temporary file
	tmpFile, err := os.CreateTemp("", fmt.Sprintf("mdx-*.%s", launcher.extension))
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(renderedCode.Bytes()); err != nil {
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary file: %v", err)
	}

	cmd := exec.Command(launcher.cmd, tmpFile.Name())
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	return nil
}

func loadCommands(markdownFile string) error {
	source, err := os.ReadFile(markdownFile)
	if err != nil {
		return err
	}

	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var currentHeading string
	var foundCodeBlock bool

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := n.(*ast.Heading); ok && entering {
			if heading.Level <= 2 {
				currentHeading = string(heading.Text(source))
				logrus.Debug("Found heading:", currentHeading)
				foundCodeBlock = false
			}
		}

		if block, ok := n.(*ast.FencedCodeBlock); ok && entering && !foundCodeBlock {
			if currentHeading != "" {
				lang := string(block.Language(source))
				code := string(block.Text(source))

				commands[currentHeading] = CommandBlock{
					Lang:   lang,
					Code:   code,
					Config: make(map[string]string),
				}
				foundCodeBlock = true
				logrus.Debug(fmt.Sprintf("Found code block. Infostring: %s, Heading: %s", lang, currentHeading))
			}
		}
		return ast.WalkContinue, nil
	})
	return nil
}
