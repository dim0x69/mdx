package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func setLogLevel() {
	logLevel := os.Getenv("MDX_LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.WarnLevel)
	}
}

func errorExit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {
	setLogLevel()

	// Define command-line flags
	fileFlag := flag.String("file", "", "Specify a markdown file")
	flag.Parse()

	logrus.Debug("MDX started with parameters:", os.Args)

	// Check for subcommands
	if flag.NArg() < 1 {
		errorExit("Usage: mdx [-file <markdown-file>] <command> [args]\n")
	}

	commandName := flag.Arg(0)
	commandArgs := []string{}
	if flag.NArg() > 1 {
		commandArgs = flag.Args()[1:]
	}

	loadLaunchers()

	// Load commands from the specified markdown file or all markdown files in the current directory
	if *fileFlag != "" {
		err := loadCommands(*fileFlag)
		if err != nil {
			errorExit("Error loading commands from %s: %v", *fileFlag, err)
		}
	} else {
		mdFiles, err := filepath.Glob("*.md")
		if err != nil {
			errorExit("Error searching for markdown files: %v", err)
		}
		if len(mdFiles) == 0 {
			errorExit("No markdown files found in the current directory")
		}
		for _, mdFile := range mdFiles {
			logrus.Debug(fmt.Sprintf("Loading file %s", mdFile))
			err := loadCommands(mdFile)
			if err != nil {
				errorExit("Error loading commands from %s: %v", mdFile, err)
			}
		}
	}

	// Test whether command is in commands
	if _, ok := commands[commandName]; ok {
		// Execute the command
		err := executeCommand(commandName, commandArgs...)
		if err != nil {
			errorExit("Error executing command: %v", err)
		}
	} else {
		errorExit("Command not found: %s", commandName)
	}
}
