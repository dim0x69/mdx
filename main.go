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
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

/*
getMarkdownFilePaths returns a list of markdown files to load commands from.
The order of precedence is:
1. The file flag
2. The MDX_FILE_DIR environment variable
3. The MDX_FILE_PATH environment variable
4. The current working directory
*/
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

func listCommands(commands map[string]CommandBlock) {
	fmt.Println("Available commands:")
	for name, command := range commands {
		if len(command.Dependencies) > 0 {
			fmt.Printf("%s: %s)\n", name, command.Dependencies)
		} else {
			fmt.Println(name)
		}
	}
}

func main() {
	setLogLevel()
	fileFlag := flag.String("file", "", "Specify a markdown file")
	fileFlagShort := flag.String("f", "", "Specify a markdown file (shorthand)")
	listFlag := flag.Bool("list", false, "list commands")
	listFlagShort := flag.Bool("l", false, "list commands (shorthand)")
	flag.Parse()

	if *fileFlagShort != "" {
		fileFlag = fileFlagShort
	}

	if *listFlagShort {
		listFlag = listFlagShort
	}

	logrus.Debug("MDX started with parameters:", os.Args)

	// Check for subcommands
	if flag.NArg() < 1 && !*listFlag {
		errorExit("Usage: mdx [-file <markdown-file>] [-list] <command> [args]")
	}

	commandName := flag.Arg(0)
	commandArgs := []string{}
	if flag.NArg() > 1 {
		commandArgs = flag.Args()[1:]
	}

	loadLaunchers()

	var commands = map[string]CommandBlock{}

	mdFiles := getMarkdownFilePaths(*fileFlag)
	for _, mdFile := range mdFiles {
		logrus.Debug(fmt.Sprintf("Loading file %s", mdFile))
		err := loadCommands(mdFile, commands)
		if err != nil {
			errorExit("Error loading commands from %s: %v", mdFile, err)
		}
	}

	if *listFlag {
		listCommands(commands)
		os.Exit(0)
	}

	if command, ok := commands[commandName]; ok {
		err := executeCommandBlock(commands, &command, commandArgs...)
		if err != nil {
			errorExit("Error executing command: %v", err)
		}
	} else {
		errorExit("Command not found: %s", commandName)
	}
}
