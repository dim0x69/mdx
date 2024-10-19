package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestParseGeneric(t *testing.T) {
	tests := []struct {
		filePath     string
		expectedCmds map[string]CommandBlock
		expectedErr  error
	}{
		{
			filePath: "tests/test1.md",
			expectedCmds: map[string]CommandBlock{
				"simple_echo": {
					Lang:         "sh",
					Code:         "echo \"{{.arg1}} {{.arg2}}\"",
					Dependencies: []string{"dep1", "dep2"},
					Meta:         map[string]interface{}{"shebang": false},
				},
			},
			expectedErr: nil,
		},
		{
			filePath: "tests/test2.md",
			expectedCmds: map[string]CommandBlock{
				"simple_echo": {
					Lang:         "",
					Code:         "#!/my/python\nprint(blubb)",
					Dependencies: []string{},
					Meta:         map[string]interface{}{"shebang": true},
				},
			},
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			commands = map[string]CommandBlock{}
			err := loadCommands(tt.filePath)

			if (err != nil || tt.expectedErr != nil) && !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}

			if len(commands) != len(tt.expectedCmds) {
				t.Fatalf("expected %d commands, got %d", len(tt.expectedCmds), len(commands))
			}

			for name, expectedCmd := range tt.expectedCmds {
				actualCmd, ok := commands[name]
				if !ok {
					t.Fatalf("expected command %s to be present", name)
				}

				if actualCmd.Lang != expectedCmd.Lang {
					t.Fatalf("expected lang %s, got %s", expectedCmd.Lang, actualCmd.Lang)
				}

				if strings.TrimSpace(actualCmd.Code) != strings.TrimSpace(expectedCmd.Code) {
					t.Fatalf("expected code %s, got %s", expectedCmd.Code, actualCmd.Code)
				}

				if !reflect.DeepEqual(actualCmd.Dependencies, expectedCmd.Dependencies) {
					t.Fatalf("expected dependencies %v, got %v", expectedCmd.Dependencies, actualCmd.Dependencies)
				}

				if !reflect.DeepEqual(actualCmd.Meta, expectedCmd.Meta) {
					t.Fatalf("expected meta %v, got %v", expectedCmd.Meta, actualCmd.Meta)
				}
			}
		})
	}
}
