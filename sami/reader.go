package sami

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/vimeo/caps"
	"golang.org/x/net/html"
)

type Reader struct {
	lines []caps.CaptionContent
}

func (r *Reader) Detect(content []byte) bool {
	return r.DetectString(string(content))
}

func (*Reader) DetectString(content string) bool {
	return strings.Contains(strings.ToLower(content), "<sami")
}

func (r *Reader) ReadString(content string) (*caps.CaptionSet, error) {
	if strings.Contains(strings.ToLower(content), "<html") {
		return nil, fmt.Errorf("sami file seems to be an html file")
	}
	metadata, err := NewParser().Feed(content)
	if err != nil {
		return nil, err
	}
	samiDoc := goquery.NewDocumentFromNode(metadata.Root)
	cs := caps.NewCaptionSet()
	cs.Styles = metadata.Styles
	for lang := range metadata.Languages {
		c := r.translateLang(lang, samiDoc)
		cs.SetCaptions(lang, c)
	}
	if cs.IsEmpty() {
		return nil, fmt.Errorf("empty caption file")
	}
	return cs, nil
}

func (r *Reader) translateLang(lang string, doc *goquery.Document) []*caps.Caption {
	var (
		cs  []*caps.Caption
		ms  float64
		err error
	)
	doc.Find(fmt.Sprintf("p[lang|=%s]", lang)).Each(func(_ int, s *goquery.Selection) {
		ms, err = strconv.ParseFloat(s.Parent().AttrOr("start", ""), 64)
		if err != nil {
			h, _ := s.Parent().Html()
			log.Printf("found p tag with parent sync tag with invalid start attr: %s", h)
			return
		}
		start := ms * 1000
		end := float64(0)
		last := len(cs) - 1
		if len(cs) > 0 && cs[last].End == nil {
			end = ms * 1000
			cs[last].End = &end
		}
		r.lines = []caps.CaptionContent{}
		n := s.Nodes[0]
		r.translateTag(n)
		_, styles := r.translateAttrs(n)
		if len(r.lines) > 0 {
			cs = append(cs, &caps.Caption{
				Start: &start,
				End:   &end,
				Nodes: r.lines[:],
				Style: *styles,
			})
		}
	})
	last := len(cs) - 1
	if len(cs) > 0 && cs[last].End == nil {
		end := (ms + 4000) * 1000
		cs[last].End = &end
	}
	return cs
}

func (r *Reader) translateTag(n *html.Node) {
	switch n.Data {
	case "br":
		r.lines = append(r.lines, caps.NewLineBreak())
	case "i":
		r.lines = append(r.lines, caps.NewCaptionStyle(true, caps.StyleProps{Italics: true}))
		child := n.FirstChild
		for child != nil {
			r.translateTag(child)
			child = child.NextSibling
		}
		r.lines = append(r.lines, caps.NewCaptionStyle(false, caps.StyleProps{Italics: true}))
	case "span", "div":
		r.translateSpan(n)
	case "p":
		fallthrough
	default:
		if n.FirstChild == nil && n.Type == html.TextNode {
			if t := strings.TrimSpace(text(n)); t != "" {
				r.lines = append(r.lines, caps.NewCaptionText(t))
			}
		} else {
			child := n.FirstChild
			for child != nil {
				r.translateTag(child)
				child = child.NextSibling
			}
		}
	}
}

func text(n *html.Node) string {
	var output func(*bytes.Buffer, *html.Node)
	output = func(buf *bytes.Buffer, n *html.Node) {
		switch n.Type {
		case html.TextNode:
			buf.WriteString(n.Data)
		case html.CommentNode:
		default:
			for child := n.FirstChild; child != nil; child = child.NextSibling {
				output(buf, child)
			}
		}
	}
	var buf bytes.Buffer
	output(&buf, n)
	return buf.String()
}
func (r *Reader) translateSpan(n *html.Node) {
	if len(n.Attr) > 0 {
		_, styleProps := r.translateAttrs(n)
		r.lines = append(r.lines, caps.NewCaptionStyle(true, *styleProps))
		child := n.FirstChild
		for child != nil {
			r.translateTag(child)
			child = child.NextSibling
		}
		r.lines = append(r.lines, caps.NewCaptionStyle(false, *styleProps))
		return
	}
	child := n.FirstChild
	for child != nil {
		r.translateTag(child)
		child = child.NextSibling
	}
}

func (r *Reader) translateAttrs(n *html.Node) (string, *caps.StyleProps) {
	props := caps.StyleProps{}
	lang := caps.DefaultLang
	for _, attr := range n.Attr {
		if attr.Key == "class" || attr.Key == "id" {
			props.Class = attr.Val
			continue
		}
		if attr.Key == "style" {
			for _, dec := range strings.Split(attr.Val, ";") {
				prop := strings.Split(dec, ":")
				if len(prop) < 2 {
					log.Printf("invalid inline style declaration: %+v", prop)
					continue
				}
				switch prop[0] {
				case "text-align":
					props.TextAlign = prop[1]
				case "font-family":
					props.FontFamily = prop[1]
				case "font-size":
					props.FontSize = prop[1]
				case "font-style":
					if strings.TrimSpace(prop[1]) == "italics" {
						props.Italics = true
					}
				case "color":
					props.Color = prop[1]
				case "lang":
					lang = prop[1]
				}
			}
		}
	}
	return lang, &props
}
