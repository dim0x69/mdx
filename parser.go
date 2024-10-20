package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type MdxHeading struct {
	ast.BaseBlock
	commandName string
	deps        []string
}

func (b *MdxHeading) Dump(source []byte, level int) {
	ast.DumpHelper(b, source, level, nil, nil)
}

func (b *MdxHeading) Kind() ast.NodeKind {
	return KindMdxHeading
}

var KindMdxHeading = ast.NewNodeKind("MdxHeading")

type MdxHeadingParser struct{}

func (p *MdxHeadingParser) Trigger() []byte {
	return []byte{'#'}
}

func (p *MdxHeadingParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	commandName, deps := extractCommandAndDepsFromHeading(string(line))

	if commandName == "" {
		logrus.Debug(fmt.Sprintf("No command found in heading: %s", line))
		return nil, parser.NoChildren
	}

	reader.Advance(len(line))
	return &MdxHeading{commandName: commandName, deps: deps}, parser.NoChildren
}

func (p *MdxHeadingParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (p *MdxHeadingParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
}

func (p *MdxHeadingParser) CanInterruptParagraph() bool {
	return false
}

func (p *MdxHeadingParser) CanAcceptIndentedLine() bool {
	return false
}

/*
extractCommandAndDepsFromHeading extracts the command name and dependencies from the given heading and source.
The command name is extracted from the link text and the dependencies are extracted from the link destination.

[commandName](dep1 dep2 dep3) => commandName, [dep1, dep2, dep3]
*/

func extractCommandAndDepsFromHeading(heading string) (string, []string) {

	// NOTE: goldmark does not support parsing links inside of a heading.
	// We have to use a regular expression to extract the command name and dependencies.
	re := regexp.MustCompile(`\[([^\]]+)\]\((?:([^)]*))?\)`)
	matches := re.FindStringSubmatch(heading)
	if len(matches) < 2 {
		return "", nil
	}

	// matches[1] is the text inside the square brackets
	// matches[2] is the URL inside the parentheses (if present)
	text := matches[1]
	commandName := strings.TrimSpace(text)
	var url string
	if len(matches) > 2 {
		url = matches[2]
		depsString := strings.TrimSpace(url)
		deps := strings.Fields(depsString)
		return commandName, deps
	}

	return commandName, []string{}
}

func loadCommands(markdownFile string) error {
	/*
		The search strategy is as follows. We start at the beginning of the document, parse the Markdown file into an AST and walk the tree:

		1 We search for a heading. (findHeadingWalker)
		2 If we find a heading, we call the findCodeBlocksWalker with the NextSibling of the Heading.
		  findCodeBlocksWalker which extracts the commands from all code blocks below this heading.
		  findCodeBlocksWalker runs until it reaches the next heading.
		3 Goto 1.
	*/

	// TODO: load all commands

	source, err := os.ReadFile(markdownFile)
	if err != nil {
		return err
	}

	md := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithBlockParsers(
				util.Prioritized(&MdxHeadingParser{}, 0),
			),
		),
	)
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var currentCommandName string
	var currentCommandDeps []string

	findCodeBlocksWalker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {

		if _, ok := n.(*ast.Heading); ok && !entering {
			return ast.WalkStop, nil
		}
		if _, ok := n.(*MdxHeading); ok && !entering {
			return ast.WalkStop, nil
		}

		if block, ok := n.(*ast.FencedCodeBlock); ok && entering {

			lang := string(block.Language(source))
			code := string(block.Text(source))

			if code == "" {
				logrus.Warn(fmt.Sprintf("Empty code block found for command '%s' in '%s'.", currentCommandName, markdownFile))
				return ast.WalkContinue, nil
			}
			code_shebang := false
			// Check for shebang
			if len(code) >= 2 && code[:2] == "#!" {
				code_shebang = true
			}

			// return an error for code blocks which have no infostring and no shebang
			if lang == "" && !code_shebang {
				logrus.Warn(fmt.Sprintf("No infostring and no shebang defined for command '%s' in '%s'.", currentCommandName, markdownFile))
				return ast.WalkStop, ErrNoInfostringOrShebang
			}

			// notify the user if both language and shebang are defined
			if lang != "" && code_shebang {
				logrus.Warn(fmt.Sprintf("Both language and shebang defined for command '%s' in '%s'. The shebang will be used!", currentCommandName, markdownFile))
			}

			commandBlock := CommandBlock{
				Lang:         lang,
				Code:         code,
				Dependencies: currentCommandDeps,
				Filename:     markdownFile,
				Meta:         make(map[string]any),
			}

			commandBlock.Meta["shebang"] = code_shebang
			commands[currentCommandName] = commandBlock
			logrus.Debug(fmt.Sprintf("Wrote code block. Infostring: '%s', Command: '%s'", lang, currentCommandName))
		}

		return ast.WalkContinue, nil
	}

	findHeadingWalker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := n.(*MdxHeading); ok && entering {
			currentCommandName = heading.commandName
			currentCommandDeps = heading.deps

			if _, exists := commands[currentCommandName]; exists {
				return ast.WalkStop, fmt.Errorf("%w: '%s' was already defined in '%s'", ErrDuplicateCommand, currentCommandName, commands[currentCommandName].Filename)
			}

			logrus.Debug(fmt.Sprintf("Found heading: '%s' with command: '%s' and dependencies: %v", string(heading.Text(source)), currentCommandName, currentCommandDeps))
			err = ast.Walk(heading.NextSibling(), findCodeBlocksWalker)
			if err != nil {
				return ast.WalkStop, err
			}
		}
		return ast.WalkContinue, nil
	}

	return ast.Walk(doc, findHeadingWalker)
}
