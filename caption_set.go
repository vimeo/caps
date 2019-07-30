package gocaption

import "strings"

type kind int

const (
	text kind = iota
	style
	lineBreak
)

type captionNode interface {
	Kind() kind
	GetContent() string
}

type captionText struct {
	content string
}

func (captionText) Kind() kind {
	return text
}

func (c captionText) GetContent() string {
	return c.content
}

type captionStyle struct {
	content string
	start   int
}

func (captionStyle) Kind() kind {
	return style
}

func (c captionStyle) GetContent() string {
	return c.content
}

type captionBreak struct{}

func (captionBreak) Kind() kind {
	return lineBreak
}

func (c captionBreak) GetContent() string {
	return "\n"
}

type Caption struct {
	start int
	end   int
	nodes []captionNode
}

func newCaption() Caption {
	return Caption{
		start: 0,
		end:   0,
		nodes: []captionNode{},
	}
}

func (c Caption) IsEmpty() bool {
	return len(c.nodes) == 0
}

func (c Caption) GetText() string {
	var content strings.Builder
	for _, node := range c.nodes {
		if node.Kind() != style {
			content.WriteString(node.GetContent())
		}
	}
	return content.String()
}

type CaptionSet struct {
	styles   map[string]string
	captions map[string][]Caption
}

func (c CaptionSet) SetCaptions(lang string, captions []Caption) {
	c.captions[lang] = captions
}

func (c CaptionSet) GetLanguages() []string {
	keys := []string{}
	for k := range c.captions {
		keys = append(keys, k)
	}
	return keys
}

func (c CaptionSet) GetCaptions(lang string) []Caption {
	captions, ok := c.captions[lang]
	if !ok {
		return []Caption{}
	}
	return captions
}

func (c CaptionSet) AddStyle(id, style string) {
	c.styles[id] = style
}

func (c CaptionSet) GetStyle(id string) string {
	style, ok := c.styles[id]
	if !ok {
		return ""
	}
	return style
}

func (c CaptionSet) GetStyles() []string {
	values := []string{}
	for _, v := range c.styles {
		values = append(values, v)
	}
	return values
}
