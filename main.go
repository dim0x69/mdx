package main

import (
	"flag"
	"log"
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

func loadLaunchers(){
	launchers["sh"] = LauncherBlock{cmd: "sh", extension: "sh"}
}

func main() {
    // Define command-line flags
    // listAllCommandsFlag := flag.Bool("all", false, "List all commands, even if extension is not installed")
	flag.Parse()

    // Check for subcommands
    if flag.NArg() > 0 {
        subcommand := flag.Arg(0)
        switch subcommand {
        // case "list":
        //     listCommands(*listAllCommandsFlag)
        //     return
		default:
			if flag.NArg() < 2 {
				log.Fatal("Usage: mdx <markdown-file> [command] [args]")
				return
			}
			// Assume the first argument is a markdown file
            markdownFile := subcommand
			command_name := flag.Arg(1)

			loadLaunchers()
            // Load the commands from the markdown file into the global structure
            err := loadCommands(markdownFile)
			if err != nil {
                log.Fatal(err)
            }

			// Test whether command is in commands
			if _, ok := commands[command_name];ok {
				// Execute the command
				err := executeCommand(command_name)
				if err != nil {
					log.Fatalf("Error executing command: %v", err)
				}
			}else {
				log.Fatalf("Command not found in %s: %s", markdownFile, command_name, )
			}
        }
    }
	log.Fatal("Usage: mdx <markdown-file> [command] [args] or mdx list [-all]")
}