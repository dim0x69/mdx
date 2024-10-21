package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetMarkdownFilePaths_FileFlag(t *testing.T) {
	mdFiles := getMarkdownFilePaths("test.md")
	if !reflect.DeepEqual(mdFiles, []string{"test.md"}) {
		t.Errorf("Expected %v, but got %v", []string{"test.md"}, mdFiles)
	}
}

func TestGetMarkdownFilePaths_MDXFileDir(t *testing.T) {
	os.Setenv("MDX_FILE_DIR", "TestGetMarkdownFilePaths_MDXFileDir")
	defer os.Unsetenv("MDX_FILE_DIR")

	// Create test directory and files
	os.Mkdir("TestGetMarkdownFilePaths_MDXFileDir", 0755)
	defer os.RemoveAll("TestGetMarkdownFilePaths_MDXFileDir")
	os.Create("TestGetMarkdownFilePaths_MDXFileDir/test1.md")
	os.Create("TestGetMarkdownFilePaths_MDXFileDir/test2.md")

	mdFiles := getMarkdownFilePaths("")
	expectedFiles, _ := filepath.Glob("TestGetMarkdownFilePaths_MDXFileDir/*.md")
	if !reflect.DeepEqual(mdFiles, expectedFiles) {
		t.Errorf("Expected %v, but got %v", expectedFiles, mdFiles)
	}
}

func TestGetMarkdownFilePaths_MDXFilePath(t *testing.T) {
	os.Setenv("MDX_FILE_PATH", "test.md")
	defer os.Unsetenv("MDX_FILE_PATH")

	mdFiles := getMarkdownFilePaths("")
	if !reflect.DeepEqual(mdFiles, []string{"test.md"}) {
		t.Errorf("Expected %v, but got %v", []string{"test.md"}, mdFiles)
	}
}

func TestGetMarkdownFilePaths_EnvAllDefined(t *testing.T) {
	os.Setenv("MDX_FILE_DIR", "testdir")
	defer os.Unsetenv("MDX_FILE_DIR")

	// Create test directory and files
	os.Mkdir("testdir", 0755)
	defer os.RemoveAll("testdir")
	os.Create("testdir/test1.md")
	os.Create("testdir/test2.md")

	os.Setenv("MDX_FILE_PATH", "test3.md")
	defer os.Unsetenv("MDX_FILE_PATH")

	mdFiles := getMarkdownFilePaths("")
	if !reflect.DeepEqual(mdFiles, []string{"testdir/test1.md", "testdir/test2.md"}) {
		t.Errorf("Expected %v, but got %v", []string{"testdir/test1.md", "testdir/test2.md"}, mdFiles)
	}
}

func TestGetMarkdownFilePaths_AllDefined(t *testing.T) {
	os.Setenv("MDX_FILE_DIR", "testdir")
	defer os.Unsetenv("MDX_FILE_DIR")

	// Create test directory and files
	os.Mkdir("testdir", 0755)
	defer os.RemoveAll("testdir")
	os.Create("testdir/test1.md")
	os.Create("testdir/test2.md")

	os.Setenv("MDX_FILE_PATH", "test3.md")
	defer os.Unsetenv("MDX_FILE_PATH")

	mdFiles := getMarkdownFilePaths("ff.md")
	if !reflect.DeepEqual(mdFiles, []string{"ff.md"}) {
		t.Errorf("Expected %v, but got %v", []string{"ff.md"}, mdFiles)
	}

}
