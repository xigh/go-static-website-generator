package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
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

	"gopkg.in/yaml.v3"
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
			tmpl := filepath.Join(tmpl, name)
			www := filepath.Join(www, name)

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
			res, err := parseMD(src)
			if err != nil {
				return err
			}

			os.MkdirAll(www, 0755)

			www := filepath.Join(www, name)
			www = strings.TrimSuffix(www, ".md") + ".htm"
			err = ioutil.WriteFile(www, []byte(res), 0644)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type HTML struct {
	Tags []string
	Text []string
}

type Front struct {
	Title *string
	Date  *string
	Image *string
}

type Page struct {
	Front   *Front
	Name    string
	Content template.HTML
}

var CR = []byte("\n")

func parseMD(src string) (string, error) {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		return "", err
	}

	var front Front

	lines := bytes.Split(buf, CR)
	for l, line := range lines {
		line = bytes.TrimSpace(line)
		if l > 0 {
			if !bytes.Equal(line, []byte("---")) {
				fmt.Printf("### %d: %s\n", l, line)
				continue
			}

			fmb := bytes.Join(lines[1:l-1], CR)
			err = yaml.Unmarshal(fmb, &front)
			if err != nil {
				return "", err
			}

			buf = bytes.Join(lines[l+1:], CR)
			break
		}

		if !bytes.Equal(line, []byte("---")) {
			break
		}
	}

	fmt.Printf("front: %#v\n", front)

	txt := text.NewReader(buf)
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		))
	doc := md.Parser().Parse(txt)

	html := &HTML{
		Tags: make([]string, 0),
	}

	nodes := doc.FirstChild()

	err = html.parseNodes(buf, "    ", nodes)
	if err != nil {
		return "", err
	}

	tmpl, err := template.ParseFiles("src/tmpl/index.htm")
	if err != nil {
		return "", err
	}

	page := &Page{
		Front:   &front,
		Name:    src,
		Content: template.HTML(strings.Join(html.Tags, "")),
	}

	var textBuf bytes.Buffer

	err = tmpl.Execute(&textBuf, page)
	if err != nil {
		return "", err
	}

	return textBuf.String(), nil
}

func (html *HTML) add_tag(tag string) {
	if len(html.Text) > 0 {
		html.Tags = append(html.Tags, strings.Join(html.Text, " "))
		html.Text = nil
	}
	html.Tags = append(html.Tags, tag)
}

func (html *HTML) add_text(text string) {
	html.Text = append(html.Text, text)
}

func (html *HTML) parseNodes(buf []byte, indent string, nodes ast.Node) error {
	var err error
	for nodes != nil {
		err = html.parseNode(buf, indent, nodes)
		if err != nil {
			return err
		}
		nodes = nodes.NextSibling()
	}
	return nil
}

func (html *HTML) parseNode(buf []byte, indent string, node ast.Node) error {
	switch n := node.(type) {
	case *ast.Text:
		text := string(n.Text(buf))
		text = strings.TrimSpace(text)
		if text != "" {
			fmt.Printf("%sText: %q\n", indent, text)
			html.add_text(text)
		}

	case *ast.ThematicBreak:
		fmt.Printf("%sThematicBreak\n", indent)

	case *ast.Heading:
		fmt.Printf("%sHeading %d\n", indent, n.Level)
		html.add_tag(fmt.Sprintf("<h%d>", n.Level))
		defer html.add_tag(fmt.Sprintf("</h%d>", n.Level))

	case *ast.Paragraph:
		fmt.Printf("%sParagraph\n", indent)
		html.add_tag("<p>")
		defer html.add_tag("</p>")

	case *ast.Blockquote:
		fmt.Printf("%sBlockquote\n", indent)
		html.add_tag("<blockquote>")
		defer html.add_tag("</blockquote>")

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
		html.add_tag("<ul>")
		defer html.add_tag("</ul>")

	case *ast.ListItem:
		fmt.Printf("%sListItem\n", indent)
		html.add_tag("<li>")
		defer html.add_tag("</li>")

	case *ast.Link:
		fmt.Printf("%sLink %q alt=%q\n", indent, n.Destination, n.Title)
		html.add_tag(fmt.Sprintf("<a href=%q>", n.Destination))
		defer html.add_tag("</a>")

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
			html.add_text(string(buf[start:stop]))
		}

	case *ast.AutoLink:
		label := string(n.Label(buf))
		text := string(n.Text(buf))
		fmt.Printf("%sAutoLink label=%q text=%q\n", indent, label, text)
		html.add_text(text)

	case *ast.Emphasis:
		fmt.Printf("%sEmphasis\n", indent)
		html.add_tag("<em>")
		defer html.add_tag("</em>")

	case *ast.CodeSpan:
		fmt.Printf("%sCodeSpan\n", indent)

	case *east.Table:
		fmt.Printf("%sTable\n", indent)
		html.add_tag("<table>")
		defer html.add_tag("</table>")

	case *east.TableHeader:
		fmt.Printf("%sTableHeader\n", indent)
		html.add_tag("<thead><tr>")
		defer html.add_tag("</tr></thead>")

	case *east.TableRow:
		fmt.Printf("%sTableRow\n", indent)
		html.add_tag("<tr>")
		defer html.add_tag("</tr>")

	case *east.TableCell:
		fmt.Printf("%sTableCell\n", indent)
		html.add_tag("<td>")
		defer html.add_tag("</td>")

	default:
		fmt.Printf("%s%#v", indent, n)
		return fmt.Errorf("not implemented")
	}

	return html.parseNodes(buf, indent+"    ", node.FirstChild())
}
