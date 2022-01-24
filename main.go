package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var (
	www  = flag.String("www", "www", "set target directory")
	src  = flag.String("src", "src/www", "set source directory")
	tmpl = flag.String("tmpl", "src/tmpl", "set template directory")
)

func main() {
	flag.Parse()

	fmt.Printf("generating website\n")

	err := process(*www, *src, *tmpl)
	if err != nil {
		log.Fatal(err)
	}
}

func process(www, src, tmpl string) error {
	infos, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, info := range infos {
		name := info.Name()
		src := filepath.Join(src, name)

		if info.IsDir() {
			www := filepath.Join(www, name)
			tmpl := filepath.Join(tmpl, name)

			err = process(www, src, tmpl)
			if err != nil {
				return err
			}
			continue
		}

		ext := filepath.Ext(name)
		switch ext {
		case ".md":
			fmt.Printf("\n%s:\n", src)
			err := parseMD(src)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func parseMD(src string) error {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	txt := text.NewReader(buf)
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		))
	doc := md.Parser().Parse(txt)
	nodes := doc.FirstChild()
	err = parseNodes(buf, "    ", nodes)
	if err != nil {
		return err
	}
	return nil
}

func parseNodes(buf []byte, indent string, nodes ast.Node) error {
	var err error
	for nodes != nil {
		err = parseNode(buf, indent, nodes)
		if err != nil {
			return err
		}
		nodes = nodes.NextSibling()
	}
	return nil
}

func parseNode(buf []byte, indent string, node ast.Node) error {
	switch n := node.(type) {
	case *ast.Text:
		text := string(n.Text(buf))
		fmt.Printf("%sText: %q\n", indent, strings.TrimSpace(text))

	case *ast.ThematicBreak:
		fmt.Printf("%sThematicBreak\n", indent)

	case *ast.Heading:
		fmt.Printf("%sHeading %d\n", indent, n.Level)

	case *ast.Paragraph:
		fmt.Printf("%sParagraph\n", indent)

	case *ast.Blockquote:
		fmt.Printf("%sBlockquote\n", indent)

	case *ast.FencedCodeBlock:
		fmt.Printf("%sFencedCodeBlock %q\n", indent, n.Language(buf))
		lines := n.Lines()
		for l := 0; l < lines.Len(); l++ {
			start := lines.At(l).Start
			stop := lines.At(l).Stop
			fmt.Printf("%s    %d: %q\n", indent, l, buf[start:stop])
		}

	case *ast.List:
		fmt.Printf("%sList\n", indent)

	case *ast.ListItem:
		fmt.Printf("%sListItem\n", indent)

	case *ast.Link:
		fmt.Printf("%sLink\n", indent)

	case *ast.TextBlock:
		fmt.Printf("%sTextBlock\n", indent)

	case *ast.HTMLBlock:
		fmt.Printf("%sHTMLBlock\n", indent)
		lines := n.Lines()
		for l := 0; l < lines.Len(); l++ {
			start := lines.At(l).Start
			stop := lines.At(l).Stop
			fmt.Printf("%s    %d: %q\n", indent, l, buf[start:stop])
		}

	case *ast.RawHTML:
		fmt.Printf("%sRawHTML\n", indent)
		segments := n.Segments
		for l := 0; l < segments.Len(); l++ {
			start := segments.At(l).Start
			stop := segments.At(l).Stop
			fmt.Printf("%s    %d: %q\n", indent, l, buf[start:stop])
		}

	case *ast.AutoLink:
		fmt.Printf("%sAutoLink\n", indent)

	case *ast.Emphasis:
		fmt.Printf("%sEmphasis\n", indent)

	case *ast.CodeSpan:
		fmt.Printf("%sCodeSpan\n", indent)

	case *east.Table:
		fmt.Printf("%sTable\n", indent)

	case *east.TableHeader:
		fmt.Printf("%sTableHeader\n", indent)

	case *east.TableRow:
		fmt.Printf("%sTableRow\n", indent)

	case *east.TableCell:
		fmt.Printf("%sTableCell\n", indent)

	default:
		fmt.Printf("%s%#v", indent, n)
		return fmt.Errorf("not implemented")
	}

	return parseNodes(buf, indent+"    ", node.FirstChild())
}
