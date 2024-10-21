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

func getMarkdownFilePaths(fileFlag string) []string {
	var mdFiles []string
	if fileFlag != "" {
		logrus.Debug("using file flag to find markdown files")
		mdFiles = []string{fileFlag}

	} else if mdxFileDir := os.Getenv("MDX_FILE_DIR"); mdxFileDir != "" {
		logrus.Debug("using MDX_FILE_DIR")
		var err error
		mdFiles, err = filepath.Glob(filepath.Join(mdxFileDir, "*.md"))
		logrus.Debug(fmt.Sprintf("Searching for markdown files in %s", mdxFileDir))
		if err != nil {
			errorExit("Error searching for markdown files in %s: %v", mdxFileDir, err)
		}
	} else if mdxFilePath := os.Getenv("MDX_FILE_PATH"); mdxFilePath != "" {
		logrus.Debug("using MDX_FILE_PATH")
		var err error
		mdFiles = []string{mdxFilePath}
		logrus.Debug(fmt.Sprintf("Searching in markdown file %s", mdxFilePath))
		if err != nil {
			errorExit("Error searching for markdown files in %s: %v", mdxFilePath, err)
		}
	} else {
		logrus.Debug("using CWD to find markdown files")
		var err error
		mdFiles, err = filepath.Glob("*.md")
		if err != nil {
			errorExit("Error searching for markdown files: %v", err)
		}
	}

	if len(mdFiles) == 0 {
		errorExit("No markdown files found")
	}
	return mdFiles
}

func main() {
	setLogLevel()

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

	// Function to load commands from a list of markdown files
	loadCommandsFromFiles := func(mdFiles []string) {
		for _, mdFile := range mdFiles {
			logrus.Debug(fmt.Sprintf("Loading file %s", mdFile))
			err := loadCommands(mdFile)
			if err != nil {
				errorExit("Error loading commands from %s: %v", mdFile, err)
			}
		}
	}

	mdFiles := getMarkdownFilePaths(*fileFlag)
	loadCommandsFromFiles(mdFiles)

	// Execute command
	if _, ok := commands[commandName]; ok {
		err := executeCommandBlock(commands[commandName], commandArgs...)
		if err != nil {
			errorExit("Error executing command: %v", err)
		}
	} else {
		errorExit("Command not found: %s", commandName)
	}
}
