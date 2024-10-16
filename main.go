package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// CommandBlock represents a parsed command block
type CommandBlock struct {
	Lang     string            // the infostring from the code fence
	Code     string            // the content of the code fence
	Args     []string          // placeholder for the future
	Filename string            // the filename of the markdown file
	Config   map[string]string // placeholder for the future
}

// Global store for parsed commands
var commands = map[string]CommandBlock{}

type LauncherBlock struct {
	cmd       string // The command to execute for the infostring above
	extension string // The file extension for the language
}

var launchers = map[string]LauncherBlock{}

func isExecutableInPath(candidates []string) string {
	for _, cmd := range candidates {
		if _, err := exec.LookPath(cmd); err == nil {
			return cmd
		}
	}
	return ""
}

func loadLaunchers() {
	addedLaunchers := []string{}

	if cmd := isExecutableInPath([]string{"sh"}); cmd != "" {
		launchers["sh"] = LauncherBlock{cmd: cmd, extension: "sh"}
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
		errorExit("Usage: mdx [-file <markdown-file>] <command> [args]")
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
