package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

// CommandBlock represents a parsed command block
type CommandBlock struct {
    Lang   string // the infostring from the code fence
    Code   string // the content of the code fence
	Args  []string // placeholder for the future
    Config map[string]string // placeholder for the future
}

// Global store for parsed commands
var commands = map[string]CommandBlock{}


type LauncherBlock struct {
	cmd string // The command to execute for the infostring above
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
    logrus.Debug("Added launchers:", addedLaunchers)
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



func main() {
	setLogLevel()
    // Define command-line flags
    // listAllCommandsFlag := flag.Bool("all", false, "List all commands, even if extension is not installed")
	flag.Parse()
	logrus.Debug("MDX started with parameters:", os.Args)
    // Check for subcommands
    if flag.NArg() > 0 {
        subcommand := flag.Arg(0)
        switch subcommand {
        // case "list":
        //     listCommands(*listAllCommandsFlag)
        //     return
		default:
			if flag.NArg() < 2 {
				fmt.Printf("Usage: mdx <markdown-file> [command] [args]")
				return
			}
			// Assume the first argument is a markdown file
            markdownFile := subcommand
			command_name := flag.Arg(1)

			loadLaunchers()
            // Load the commands from the markdown file into the global structure
            err := loadCommands(markdownFile)
			if err != nil {
                fmt.Print(err)
            }

			// Test whether command is in commands
			if _, ok := commands[command_name];ok {
				// Execute the command
				err := executeCommand(command_name)
				if err != nil {
					fmt.Printf("Error executing command: %v", err)
				}
			}else {
				fmt.Printf("Command not found in %s: %s", markdownFile, command_name, )
			}
			return
        }
    }
	fmt.Print("Usage: mdx <markdown-file> [command] [args] or mdx list [-all]")
}