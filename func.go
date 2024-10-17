package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
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
	Lang         string         // the infostring from the code fence
	Code         string         // the content of the code fence
	Dependencies []string       // to execute before this command
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

	if !commandBlock.Meta["shebang"].(bool) {
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
	/*
		The search strategy is as follows. We start at the beginning of the document, parse the Markdown file into an AST and walk the tree:

		1 We search for a heading. (findHeadingWalker)
		2 If we find a heading, we call the findCodeBlocksWalker with the NextSibling of the Heading.
		  findCodeBlocksWalker which extracts the commands from all code blocks below this heading.
		  findCodeBlocksWalker runs until it reaches the next heading.
		3 Goto 1.
	*/

	// TODO: load all commands

	source, err := os.ReadFile(markdownFile)
	if err != nil {
		return err
	}

	md := goldmark.New()
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var currentCommandName string
	var currentCommandDeps []string

	findCodeBlocksWalker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {

		if _, ok := n.(*ast.Heading); ok && !entering {
			return ast.WalkStop, nil
		}

		if block, ok := n.(*ast.FencedCodeBlock); ok && entering {

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
					logrus.Debug(fmt.Sprintf("no launcher defined for infostring: '%s'. Ignoring command '%s' in '%s'", lang, currentCommandName, markdownFile))
					return ast.WalkContinue, nil
				}
			}

			// ignore code blocks which have no infostring and no shebang
			if lang == "" && !code_shebang {
				logrus.Debug(fmt.Sprintf("No infostring and no shebang defined for command '%s' in '%s'. Ignoring command.", currentCommandName, markdownFile))
				return ast.WalkContinue, nil
			}

			// notify the user if both language and shebang are defined
			if lang != "" && code_shebang {
				logrus.Warn(fmt.Sprintf("Both language and shebang defined for command '%s' in '%s'. The shebang will be used!", currentCommandName, markdownFile))
			}

			commandBlock := CommandBlock{
				Lang:         lang,
				Code:         code,
				Dependencies: currentCommandDeps,
				Filename:     markdownFile,
				Meta:         make(map[string]any),
			}

			commandBlock.Meta["shebang"] = code_shebang
			commands[currentCommandName] = commandBlock
			logrus.Debug(fmt.Sprintf("Wrote code block. Infostring: '%s', Command: '%s'", lang, currentCommandName))
		}

		return ast.WalkContinue, nil
	}

	findHeadingWalker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := n.(*ast.Heading); ok && entering {
			commandName, dependencies := extractCommandAndDepsFromHeading(string(heading.Text(source)))
			if commandName == "" {
				logrus.Debug(fmt.Sprintf("No command found in heading: '%s'. Skipping.", string(heading.Text(source))))
				return ast.WalkContinue, nil
			}
			currentCommandName = commandName
			currentCommandDeps = dependencies

			if _, exists := commands[currentCommandName]; exists {
				return ast.WalkStop, fmt.Errorf("duplicate command found: '%s' was already defined in '%s'", currentCommandName, commands[currentCommandName].Filename)
			}

			logrus.Debug(fmt.Sprintf("Found heading: '%s' with command: '%s' and dependencies: %v", string(heading.Text(source)), currentCommandName, currentCommandDeps))
			err = ast.Walk(heading.NextSibling(), findCodeBlocksWalker)
			if err != nil {
				return ast.WalkStop, nil
			}
		}
		return ast.WalkContinue, nil
	}

	return ast.Walk(doc, findHeadingWalker)
}

// extractCommandAndDepsFromHeading extracts the command name and dependencies from the given heading and source.
// The command name is extracted from the link text and the dependencies are extracted from the link destination.
// [commandName](dep1 dep2 dep3) => commandName, [dep1, dep2, dep3]
func extractCommandAndDepsFromHeading(heading string) (string, []string) {

	// NOTE: goldmark does not support parsing links inside of a heading.
	// We have to use a regular expression to extract the command name and dependencies.

	re := regexp.MustCompile(`\[\s*(.*?)\s*\]\s*\((.*?)\)`)
	matches := re.FindStringSubmatch(heading)
	if len(matches) < 3 {
		return "", nil
	}

	commandName := strings.TrimSpace(matches[1])
	depsString := strings.TrimSpace(matches[2])
	deps := strings.Fields(depsString)

	return commandName, deps
}
