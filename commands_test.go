package main

import (
	"strings"
	"testing"
)

func TestParse1(t *testing.T) {
	commands = map[string]CommandBlock{}

	loadCommands("tests/test1.md")
	if len(commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(commands))
	}
	if _, ok := commands["simple_echo"]; !ok {
		t.Fatalf("expected command simple_echo to be present")
	}
	// test if the command is parsed correctly
	if commands["simple_echo"].Lang != "sh" {
		t.Fatalf("expected lang sh, got %s", commands["simple_echo"].Lang)
	}
	if strings.TrimSpace(commands["simple_echo"].Code) != "echo \"{{.arg1}} {{.arg2}}\"" {
		t.Fatalf("expected code echo hello, got \"%s\"", commands["simple_echo"].Code)
	}

}
