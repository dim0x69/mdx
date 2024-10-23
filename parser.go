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

func loadCommands(markdownFile string, commands map[string]CommandBlock) error {
	/*
		The search strategy is as follows. We start at the beginning of the document, parse the Markdown file into an AST and walk the tree:

		1 We search for a heading. (findHeadingWalker)
		2 If we find a heading, we walk all siblings of the heading and call praseCodeBlock for all FencedCodeBlock nodes.
		  praseCodeBlock extacts the code from the code block, updates the currentCommandBlock and appends the code block to the currentCommandBlock.CodeBlocks.
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

	var currentCommandBlock CommandBlock

	praseCodeBlock := func(n ast.Node) error {

		if block, ok := n.(*ast.FencedCodeBlock); ok {

			lang := string(block.Language(source))
			code := string(block.Text(source))

			if code == "" {
				logrus.Warn(fmt.Sprintf("Empty code block found for command '%s' in '%s'.", currentCommandBlock.Name, markdownFile))
				return nil
			}
			code_shebang := false
			if len(code) >= 2 && code[:2] == "#!" {
				code_shebang = true
			}

			if lang == "" && !code_shebang {
				logrus.Warn(fmt.Sprintf("Found Code Block with no infostring and no shebang defined for command '%s' in '%s'. Ignoring", currentCommandBlock.Name, markdownFile))
				return nil
			}

			if lang != "" && code_shebang {
				logrus.Warn(fmt.Sprintf("Both language and shebang defined for command '%s' in '%s'. The shebang will be used!", currentCommandBlock.Name, markdownFile))
			}

			codeBlock := CodeBlock{
				Lang: lang,
				Code: code,
				Meta: make(map[string]any),
			}
			codeBlock.Meta["shebang"] = code_shebang

			currentCommandBlock.CodeBlocks = append(currentCommandBlock.CodeBlocks, codeBlock)
			logrus.Debug(fmt.Sprintf("Wrote new code block. Infostring: '%s', Command: '%s'", lang, currentCommandBlock.Name))
		}

		return nil
	}

	findHeadingWalker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if heading, ok := n.(*MdxHeading); ok && entering {
			currentCommandBlock = CommandBlock{}
			currentCommandBlock.Filename = markdownFile
			currentCommandBlock.Meta = make(map[string]any)
			currentCommandBlock.CodeBlocks = []CodeBlock{}
			currentCommandBlock.Name = heading.commandName
			currentCommandBlock.Dependencies = heading.deps

			if _, exists := commands[currentCommandBlock.Name]; exists {
				return ast.WalkStop, fmt.Errorf("%w: '%s' was already defined in '%s'", ErrDuplicateCommand, currentCommandBlock.Name, commands[currentCommandBlock.Name].Filename)
			}

			logrus.Debug(fmt.Sprintf("Found heading with command: '%s' and dependencies: %v", currentCommandBlock.Name, currentCommandBlock.Dependencies))
			// findCodeBlocksWalker will extract the code blocks below this heading
			// and append them to the currentCommandBlock.CodeBlocks

			for sibling := heading.NextSibling(); sibling != nil; sibling = sibling.NextSibling() {
				if _, ok := sibling.(*ast.Heading); ok {
					break
				}
				if _, ok := sibling.(*MdxHeading); ok {
					break
				}
				if _, ok := sibling.(*ast.FencedCodeBlock); ok {
					err = praseCodeBlock(sibling)
				}
				if err != nil {
					return ast.WalkStop, err
				}
			}

			if len(currentCommandBlock.CodeBlocks) > 0 {
				commands[currentCommandBlock.Name] = currentCommandBlock
			} else {
				logrus.Debug(fmt.Sprintf("No code blocks found for command '%s' in '%s'.", currentCommandBlock.Name, markdownFile))
			}

		}
		return ast.WalkContinue, nil
	}

	return ast.Walk(doc, findHeadingWalker)
}
