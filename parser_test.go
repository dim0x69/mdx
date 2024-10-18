package main

import (
	"reflect"
	"testing"
)

func TestExtractCommandAndDepsFromHeading(t *testing.T) {
	tests := []struct {
		heading      string
		expectedCmd  string
		expectedDeps []string
	}{
		{
			heading:      "[commandName](dep1 dep2 dep3)",
			expectedCmd:  "commandName",
			expectedDeps: []string{"dep1", "dep2", "dep3"},
		},
		{
			heading:      "# This is a heading [commandName](dep1 dep2 dep3) with some text",
			expectedCmd:  "commandName",
			expectedDeps: []string{"dep1", "dep2", "dep3"},
		},
		{
			heading:      "[commandName]()",
			expectedCmd:  "commandName",
			expectedDeps: []string{},
		},
		{
			heading:      "[commandName]",
			expectedCmd:  "",
			expectedDeps: nil,
		},
		{
			heading:      "[commandName](dep1)",
			expectedCmd:  "commandName",
			expectedDeps: []string{"dep1"},
		},
		{
			heading:      "NoCommand",
			expectedCmd:  "",
			expectedDeps: nil,
		},
		{
			heading:      "[commandName](   dep1   dep2   dep3   )",
			expectedCmd:  "commandName",
			expectedDeps: []string{"dep1", "dep2", "dep3"},
		},
		{
			heading:      "[commandName](dep1, dep2, dep3)",
			expectedCmd:  "commandName",
			expectedDeps: []string{"dep1,", "dep2,", "dep3"},
		},
	}

	for _, test := range tests {
		cmd, deps := extractCommandAndDepsFromHeading(test.heading)
		if cmd != test.expectedCmd {
			t.Errorf("extractCommandAndDepsFromHeading(%q) = %q; want %q", test.heading, cmd, test.expectedCmd)
		}
		if !reflect.DeepEqual(deps, test.expectedDeps) {
			t.Errorf("extractCommandAndDepsFromHeading(%q) = %v; want %v", test.heading, deps, test.expectedDeps)
		}
	}
}
