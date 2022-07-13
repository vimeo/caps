package sami

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	cssparser "github.com/aymerick/douceur/parser"
	"github.com/vimeo/caps"
	"golang.org/x/net/html"
)

type Parser struct {
	selfClosingToken bool
	root             *html.Node
	sami             string
	line             string
	styles           map[string]caps.StyleProps
	langs            map[string]struct{}
	lastElement      *html.Tokenizer
	name2codepoint   map[string]int
}

func NewParser() *Parser {
	return &Parser{
		sami:           "",
		line:           "",
		styles:         map[string]caps.StyleProps{},
		langs:          map[string]struct{}{},
		lastElement:    nil,
		name2codepoint: map[string]int{"apos": 0x0027},
	}
}

type ParsedSAMI struct {
	Root      *html.Node
	Styles    map[string]caps.StyleProps
	Languages map[string]struct{}
}

func (p *Parser) Feed(data string) (*ParsedSAMI, error) {
	if strings.HasPrefix(data, "<html") {
		return nil, fmt.Errorf("SAMI file seems to be an HTML file")
	}
	if strings.Contains(data, "no closed captioning available") {
		return nil, fmt.Errorf("SAMI file contains 'no closed captioning available'")
	}
	data = strings.ReplaceAll(data, "<i/>", "<i>")
	data = strings.ReplaceAll(data, ";>", ">")
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	st, err := p.parseCSS(doc)
	if err != nil {
		return nil, err
	}
	p.styles = st
	p.addLangs(doc)
	return &ParsedSAMI{
		Root:      doc.Find("sami").Nodes[0],
		Styles:    p.styles,
		Languages: p.langs,
	}, err
}

func (p *Parser) findLang(n *html.Node) string {
	for _, attr := range n.Attr {
		attrName := strings.ToLower(attr.Key)
		if attrName == "lang" && attr.Val != "" {
			return attr.Val
		}
		if attrName == "class" {
			if l := p.styles[attr.Val].Lang; l != "" {
				return l
			}
		}
	}
	return caps.DefaultLang
}

func (p *Parser) addLangs(doc *goquery.Document) {
	for _, n := range doc.Find("p").Nodes {
		lang := p.findLang(n)
		n.Attr = append(n.Attr, html.Attribute{
			Key: "lang",
			Val: lang,
		})
		p.langs[lang] = struct{}{}
	}
}

var (
	supportedStyleProps = map[string]struct{}{
		"text-align":  {},
		"font-family": {},
		"font-size":   {},
		"color":       {},
		"lang":        {},
	}
)

func (p *Parser) parseCSS(doc *goquery.Document) (map[string]caps.StyleProps, error) {
	sheet, err := cssparser.Parse(doc.Find("style").Text())
	if err != nil {
		return nil, err
	}
	st := map[string]caps.StyleProps{}
	for _, r := range sheet.Rules {
		props := map[string]string{}
		for _, prop := range r.Declarations {
			if _, ok := supportedStyleProps[prop.Property]; ok {
				props[prop.Property] = prop.Value
			}
		}
		if len(props) > 0 {
			sp := caps.DefaultStyleProps()
			for k, v := range props {
				switch k {

				case "text-align":
					sp.TextAlign = v
				case "font-family":
					sp.FontFamily = v
				case "font-size":
					sp.FontSize = v
				case "font-style":
					if strings.TrimSpace(v) == "italics" {
						sp.Italics = true
					}
				case "color":
					sp.Color = v
				case "lang":
					sp.Lang = v
				}
			}
			st[p.selector(r.Selectors)] = sp
		}
	}

	// TODO:
	//cv = cssutils_css.ColorValue(prop.value)
	//	# Code for RGB to hex conversion comes from
	//# http://bit.ly/1kwfBnQ
	//	new_style[u'color'] = u"#%02x%02x%02x" % (
	//	    cv.red, cv.green, cv.blue)

	return st, nil
}

func (p *Parser) selector(selectors []string) string {
	if len(selectors) == 0 {
		return ""
	}
	// NOTE: python only makes this change in the first selector, so
	// we're ignoring any other selectors from the list.
	// TODO: maybe SAMI spec only support simple selectors by id and
	// class? that would explain why python only changes the first one.
	s := selectors[0]
	if s[0] == '#' || s[0] == '.' {
		s = s[1:]
	}
	return strings.ToLower(s)
}
