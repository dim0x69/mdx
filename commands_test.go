package main

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type FileParseTest struct {
	filePath     string
	expectedCmds map[string]CommandBlock
	expectedErr  error
}

func RunFileParseTest(t *testing.T, tt *FileParseTest) {
	t.Run(tt.filePath, func(t *testing.T) {
		commands = map[string]CommandBlock{}
		err := loadCommands(tt.filePath)

		if (err != nil || tt.expectedErr != nil) && !errors.Is(err, tt.expectedErr) {
			t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			return
		}
		if (err != nil || tt.expectedErr != nil) && errors.Is(err, tt.expectedErr) {
			return
		}

		if len(commands) != len(tt.expectedCmds) {
			t.Fatalf("expected %d commands, got %d", len(tt.expectedCmds), len(commands))
		}

		for name, expectedCmd := range tt.expectedCmds {
			actualCmd, ok := commands[name]
			if !ok {
				t.Fatalf("expected command %s to be present", name)
			}

			for i, expectedCodeBlock := range expectedCmd.CodeBlocks {
				actualCodeBlock := actualCmd.CodeBlocks[i]
				if actualCodeBlock.Lang != expectedCodeBlock.Lang {
					t.Fatalf("expected lang %s, got %s", expectedCodeBlock.Lang, actualCodeBlock.Lang)
				}

				if strings.TrimSpace(actualCodeBlock.Code) != strings.TrimSpace(expectedCodeBlock.Code) {
					t.Fatalf("expected code %s, got %s", expectedCodeBlock.Code, actualCodeBlock.Code)
				}

				if !reflect.DeepEqual(actualCodeBlock.Meta, expectedCodeBlock.Meta) {
					t.Fatalf("expected meta %v, got %v", expectedCodeBlock.Meta, actualCodeBlock.Meta)
				}
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
func TestOneCommandWithDeps(t *testing.T) {
	test := &FileParseTest{
		filePath: "tests/test1.md",
		expectedCmds: map[string]CommandBlock{
			"simple_echo": {
				CodeBlocks: []CodeBlock{{
					Lang: "sh",
					Code: "echo \"{{.arg1}} {{.arg2}}\"",
					Meta: map[string]interface{}{"shebang": false},
				}},
				Dependencies: []string{"dep1", "dep2"},
				Meta:         map[string]interface{}{},
			},
		},
		expectedErr: nil,
	}
	RunFileParseTest(t, test)
}

func TestTwoCommands(t *testing.T) {
	test := &FileParseTest{
		filePath: "tests/two_commands.md",
		expectedCmds: map[string]CommandBlock{
			"simple_echo1": {
				CodeBlocks: []CodeBlock{{
					Lang: "sh",
					Code: "code1",
					Meta: map[string]interface{}{"shebang": false},
				}},
				Dependencies: []string{"dep1"},
				Meta:         map[string]interface{}{},
			},
			"simple_echo2": {
				CodeBlocks: []CodeBlock{{
					Lang: "sh",
					Code: "code2",
					Meta: map[string]interface{}{"shebang": false},
				}},
				Dependencies: []string{"dep1", "dep2"},
				Meta:         map[string]interface{}{},
			},
		},
		expectedErr: nil,
	}
	RunFileParseTest(t, test)
}

func TestParseShebang(t *testing.T) {
	test := &FileParseTest{
		filePath: "tests/test2.md",
		expectedCmds: map[string]CommandBlock{
			"simple_echo": {
				CodeBlocks: []CodeBlock{{
					Lang: "",
					Code: "#!/my/python\nprint(blubb)",
					Meta: map[string]interface{}{"shebang": true},
				}},
				Dependencies: []string{},
				Meta:         map[string]interface{}{},
			},
		},
		expectedErr: nil,
	}
	RunFileParseTest(t, test)
}

func TestTwoCommandHaveSameName(t *testing.T) {
	test := &FileParseTest{
		filePath:     "tests/err_same_command_name.md",
		expectedCmds: nil,
		expectedErr:  ErrDuplicateCommand,
	}
	RunFileParseTest(t, test)
}

func TestNoShebangNoInfostringDefined(t *testing.T) {
	test := &FileParseTest{
		filePath:     "tests/err_no_shebang_no_infostring.md",
		expectedCmds: nil,
		expectedErr:  ErrNoInfostringOrShebang,
	}
	RunFileParseTest(t, test)

}

func TestNoCodeInCodeFence(t *testing.T) {
	test := &FileParseTest{
		filePath:     "tests/no_code_in_codefence.md",
		expectedCmds: nil,
		expectedErr:  nil,
	}
	RunFileParseTest(t, test)

}
