package sami

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

// Parse returns the root of parsed SAMI tree that was read from the io.Reader.
//
// It doesn't implement the SAMI spec and its basically a super simple tag parser,
// watered down from golang.org/x/net/html, using its tokenizer.
// The main reason we had to create this is just so we can have an html-like parser that
// doesn't enforce HTML5 spec restrictions and semantics(e.g. requiring
// the document to have an html root with head/body, html5 semantics, etc).
// Since most golang pkgs for parsing/querying/editing html do parsing using golang.org/x/net/html,
// they all end up messing with the resulting SAMI tree, which can lead to
// weird/unexpected results, so this fixes that.
//
// The input is assumed to be UTF-8 encoded.
func Parse(r io.Reader) (*html.Node, error) {
	p := &parser{
		tokenizer: html.NewTokenizer(r),
		root: &html.Node{
			Type: html.DocumentNode,
		},
	}
	// this is default value already, but lets force it anyways
	p.tokenizer.AllowCDATA(false)
	if err := p.parse(); err != nil {
		return nil, err
	}
	return p.root, nil
}

type parser struct {
	// tokenizer provides the tokens for the parser.
	tokenizer *html.Tokenizer
	// tok is the most recently read token.
	tok html.Token
	// root is the root element of sami tree.
	root *html.Node
	// The stack of open elements
	oe nodeStack
}

// top method returns the current open tag on the stack or the top most one(the root parent)
func (p *parser) top() *html.Node {
	if n := p.oe.top(); n != nil {
		return n
	}
	return p.root
}

// addChild adds a child node n to the top element, and pushes n onto the stack
// of open elements if it is an element node.
func (p *parser) addChild(n *html.Node) {
	p.top().AppendChild(n)
	if n.Type == html.ElementNode {
		p.oe = append(p.oe, n)
	}
}

// addText adds text to the preceding node if it is a text node, or else it
// calls addChild with a new text node.
func (p *parser) addText(text string) {
	if text == "" {
		return
	}
	t := p.top()
	if n := t.LastChild; n != nil && n.Type == html.TextNode {
		n.Data += text
		return
	}
	p.addChild(&html.Node{
		Type: html.TextNode,
		Data: text,
	})
}

// addElement adds a child element based on the current token.
func (p *parser) addElement() {
	p.addChild(&html.Node{
		Type:     html.ElementNode,
		DataAtom: p.tok.DataAtom,
		Data:     p.tok.Data,
		Attr:     p.tok.Attr,
	})
}

// parseCurrentToken runs the current token through the parsing routines
// until it is consumed.
func (p *parser) parseCurrentToken() {
	switch p.tok.Type {
	case html.TextToken:
		d := p.tok.Data
		d = strings.ReplaceAll(d, "\x00", "")
		if d == "" {
			return
		}
		p.addText(d)
	case html.SelfClosingTagToken:
		p.addElement()
		p.oe.pop()
	case html.StartTagToken:
		p.addElement()
	case html.EndTagToken:
		p.oe.pop()
	}
}

func (p *parser) parse() error {
	var err error
	for err != io.EOF {
		p.tokenizer.Next()
		p.tok = p.tokenizer.Token()
		if p.tok.Type == html.ErrorToken {
			err = p.tokenizer.Err()
			if err != nil && err != io.EOF {
				return err
			}
		}
		p.parseCurrentToken()
	}
	return nil
}

// nodeStack is a stack of nodes.
type nodeStack []*html.Node

// pop pops the stack.
func (s *nodeStack) pop() *html.Node {
	if len(*s) == 0 {
		return nil
	}
	i := len(*s)
	n := (*s)[i-1]
	*s = (*s)[:i-1]
	return n
}

// top returns the most recently pushed node, or nil if s is empty.
func (s *nodeStack) top() *html.Node {
	if i := len(*s); i > 0 {
		return (*s)[i-1]
	}
	return nil
}
