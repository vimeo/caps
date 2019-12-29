package caps

import (
	"fmt"
	"strings"
)

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

func CreateText(text string) captionNode {
	return captionText{text}
}

func CreateBreak() captionNode {
	return captionBreak{}
}

func (captionText) Kind() kind {
	return text
}

func (c captionText) GetContent() string {
	return c.content
}

type captionStyle struct {
	content map[string]string
	start   bool
}

func (captionStyle) Kind() kind {
	return style
}

func (c captionStyle) GetContent() string {
	rawContent := ""
	for k, v := range c.content {
		rawContent += fmt.Sprintf("%s: %s\n", k, v)
	}
	return rawContent
}

func CreateStyle(start bool, content map[string]string) captionNode {
	return captionStyle{content, start}
}

type captionBreak struct{}

func (captionBreak) Kind() kind {
	return lineBreak
}

func (c captionBreak) GetContent() string {
	return "\n"
}

type Caption struct {
	Start  int
	End    int
	Nodes  []captionNode
	Styles map[string]string
}

func NewCaption(start, end int, nodes []captionNode, styles map[string]string) Caption {
	return Caption{
		start,
		end,
		nodes,
		styles,
	}
}

func DefaultCaption() Caption {
	return Caption{
		Start: 0,
		End:   0,
		Nodes: []captionNode{},
	}
}

func (c Caption) IsEmpty() bool {
	return len(c.Nodes) == 0
}

func (c Caption) GetText() string {
	var content strings.Builder
	for _, node := range c.Nodes {
		if node.Kind() != style {
			content.WriteString(node.GetContent())
		}
	}
	return content.String()
}

type CaptionSet struct {
	Styles   map[string]map[string]string
	Captions map[string][]*Caption
}

func NewCaptionSet() *CaptionSet {
	return &CaptionSet{
		Styles:   map[string]map[string]string{},
		Captions: map[string][]*Caption{},
	}
}

func (c CaptionSet) SetCaptions(lang string, captions []*Caption) {
	c.Captions[lang] = captions
}

func (c CaptionSet) GetLanguages() []string {
	keys := []string{}
	for k := range c.Captions {
		keys = append(keys, k)
	}
	return keys
}

func (c CaptionSet) IsEmpty() bool {
	for _, captions := range c.Captions {
		if len(captions) != 0 {
			return false
		}
	}
	return true
}

func (c CaptionSet) GetCaptions(lang string) []*Caption {
	captions, ok := c.Captions[lang]
	if !ok {
		return []*Caption{}
	}
	return captions
}

func (c CaptionSet) AddStyle(id string, style map[string]string) {
	c.Styles[id] = style
}

func (c CaptionSet) GetStyle(id string) map[string]string {
	if style, ok := c.Styles[id]; ok {
		return style
	}
	return map[string]string{}
}

func (c CaptionSet) GetStyles() []map[string]string {
	values := []map[string]string{}
	for _, v := range c.Styles {
		values = append(values, v)
	}
	return values
}
