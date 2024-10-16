package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// executeCommand writes the command code to a temporary file and executes it
func executeCommand(commandName string) error {

    commandBlock, ok := commands[commandName]
    if !ok {
        return fmt.Errorf("command not found: %s", commandName)
    }

    launcher, ok := launchers[commandBlock.Lang]
    if !ok {
        return fmt.Errorf("launcher not found for language: %s", commandBlock.Lang)
    }

    // Write the command.code to the temporary file
    tmpFile, err := os.CreateTemp("", fmt.Sprintf("mdx-*.%s", launcher.extension))
    if err != nil {
        return fmt.Errorf("failed to create temporary file: %v", err)
    }
    defer os.Remove(tmpFile.Name()) // Clean up the file afterwards

    if _, err := tmpFile.Write([]byte(commandBlock.Code)); err != nil {
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
            }
        }
        return ast.WalkContinue, nil
    })
    return nil
}


