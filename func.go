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

func isExecutableInPath(candidates []string) string {
	for _, cmd := range candidates {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}

// CommandBlock represents a parsed command block
type CommandBlock struct {
	Lang     string         // the infostring from the code fence
	Code     string         // the content of the code fence
	Args     []string       // placeholder for the future
	Filename string         // the filename of the markdown file
	Config   map[string]any // placeholder for the future
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

	// Add more binaries as needed

	// Print added launchers
	logrus.Debug("Added launchers: ", addedLaunchers)
}

func executeCommand(commandName string, args ...string) error {
	logrus.Debug(fmt.Sprintf("Executing command %s with args %v", commandName, args))

	commandBlock, ok := commands[commandName]
	if !ok {
		return fmt.Errorf("command not found: %s", commandName)
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

	launcher, ok := launchers[commandBlock.Lang]
	if !ok {
		return fmt.Errorf("launcher not found for language: %s", commandBlock.Lang)
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

	if !commandBlock.Config["shebang"].(bool) {
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

func loadCommands(markdownFile string) error {
	source, err := os.ReadFile(markdownFile)
	if err != nil {
		return err
	}

	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var currentHeadingCommand string
	var foundCodeBlock bool
	var currentHeading string

	return ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := n.(*ast.Heading); ok && entering {
			if heading.Level >= 2 {
				currentHeading = string(heading.Text(source))
				currentHeadingCommand = extractInlineCodeFromHeading(heading, source)
				logrus.Debug(fmt.Sprintf("Found heading: '%s' and command: '%s'", currentHeading, currentHeadingCommand))
				foundCodeBlock = false
			}
		}

		if block, ok := n.(*ast.FencedCodeBlock); ok && entering && !foundCodeBlock {

			lang := string(block.Language(source))
			code := string(block.Text(source))

			code_shebang := false
			// Check for shebang
			if len(code) >= 2 && code[:2] == "#!" {
				code_shebang = true
			}

			// ignore code blocks for languages which have neither infostring nor shebang defined.
			// Notify the user
			if lang == "" && !code_shebang {
				if _, launcherExists := launchers[lang]; !launcherExists {
					logrus.Debug(fmt.Sprintf("no launcher defined for infostring: '%s'. Ignoring command '%s' in '%s'", lang, currentHeadingCommand, markdownFile))
					return ast.WalkContinue, nil
				}
			}

			if currentHeading == "" {
				return ast.WalkStop, fmt.Errorf("no heading found for code block in file: '%s'", markdownFile)
			}
			if currentHeadingCommand == "" {
				return ast.WalkStop, fmt.Errorf("no inline code found in heading: %s", currentHeading)
			}

			// ignore code blocks which have no infostring and no shebang
			if lang == "" && !code_shebang {
				logrus.Debug(fmt.Sprintf("No infostring and no shebang defined for command '%s' in '%s'. Ignoring command.", currentHeadingCommand, markdownFile))
				return ast.WalkContinue, nil
			}

			// notify the user if both language and shebang are defined
			if lang != "" && code_shebang {
				logrus.Warn(fmt.Sprintf("Both language and shebang defined for command '%s' in '%s'. The shebang will be used!", currentHeadingCommand, markdownFile))
			}

			if currentHeadingCommand != "" {
				if _, exists := commands[currentHeadingCommand]; exists {
					return ast.WalkStop, fmt.Errorf("duplicate command found: '%s' was already defined in '%s'", currentHeadingCommand, commands[currentHeadingCommand].Filename)
				}
				commandBlock := CommandBlock{
					Lang:     lang,
					Code:     code,
					Filename: markdownFile,
					Config:   make(map[string]any),
				}

				commandBlock.Config["shebang"] = code_shebang
				commands[currentHeadingCommand] = commandBlock
				foundCodeBlock = true
				logrus.Debug(fmt.Sprintf("Found code block. Infostring: '%s', Command: '%s'", lang, currentHeadingCommand))
			}
		}
		return ast.WalkContinue, nil
	})
}

func extractInlineCodeFromHeading(heading *ast.Heading, source []byte) string {
	for child := heading.FirstChild(); child != nil; child = child.NextSibling() {
		if codeSpan, ok := child.(*ast.CodeSpan); ok {
			return string(codeSpan.Text(source))
		}
	}
	return ""
}
