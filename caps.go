package caps

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Kind int

const (
	Text Kind = iota
	CapStyle
	LineBreak
)

type CaptionNode interface {
	Kind() Kind
	GetContent() string
}

type CaptionText struct {
	content string
}

func CreateText(text string) CaptionNode {
	return CaptionText{text}
}

func CreateBreak() CaptionNode {
	return captionBreak{}
}

func (CaptionText) Kind() Kind {
	return Text
}

func (c CaptionText) GetContent() string {
	return c.content
}

type CaptionStyle struct {
	Style Style
	Start bool
}

func (CaptionStyle) Kind() Kind {
	return CapStyle
}

func (c CaptionStyle) GetContent() string {
	return c.Style.String()
}

func CreateCaptionStyle(start bool, style Style) CaptionNode {
	return CaptionStyle{style, start}
}

type captionBreak struct{}

func (captionBreak) Kind() Kind {
	return LineBreak
}

func (c captionBreak) GetContent() string {
	return "\n"
}

type Caption struct {
	Start int
	End   int
	Nodes []CaptionNode
	Style Style
}

const defaultStyleID = "default"

// FIXME This is a simple placeholder for style types, this can be better represented
// but I need to implement more caption types first(this was written with just the dfxp)
type Style struct {
	ID         string
	Class      string
	TextAlign  string
	FontFamily string
	FontSize   string
	Color      string
	Italics    bool
	Bold       bool
	Underline  bool
}

func (s Style) String() string {
	return fmt.Sprintf(`
	class: %s\n
	text-align: %s\n
	font-family: %s\n
	font-size: %s\n
	color: %s\n
	italics: %s\n
	bold: %s\n
	underline: %s\n
	`,
		s.Class,
		s.TextAlign,
		s.FontFamily,
		s.FontSize,
		s.Color,
		strconv.FormatBool(s.Italics),
		strconv.FormatBool(s.Bold),
		strconv.FormatBool(s.Underline),
	)
}

func DefaultStyle() Style {
	return Style{Color: "white", FontFamily: "monospace", FontSize: "1c"}
}

func NewCaption(start, end int, nodes []CaptionNode, style Style) Caption {
	return Caption{
		start,
		end,
		nodes,
		style,
	}
}

func DefaultCaption() Caption {
	return Caption{
		Start: 0,
		End:   0,
		Nodes: []CaptionNode{},
	}
}

func (c Caption) IsEmpty() bool {
	return len(c.Nodes) == 0
}

func (c Caption) GetText() string {
	var content strings.Builder
	for _, node := range c.Nodes {
		if node.Kind() != CapStyle {
			content.WriteString(node.GetContent())
		}
	}
	return content.String()
}
func (c Caption) FormatStartWithSeparator(sep string) string {
	return formatTimestamp(c.Start, sep)
}

func (c Caption) FormatStart() string {
	return formatTimestamp(c.Start, ".")
}

func (c Caption) FormatEndWithSeparator(sep string) string {
	return formatTimestamp(c.End, sep)
}

func (c Caption) FormatEnd() string {
	return formatTimestamp(c.End, ".")
}

func formatTimestamp(value int, sep string) string {
	value /= 1000
	seconds := math.Mod(float64(value)/1000, 60)
	minutes := (value / (1000 * 60)) % 60
	hours := (value / (1000 * 60 * 60) % 24)
	timestamp := fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, seconds)
	if sep != "." {
		return strings.ReplaceAll(timestamp, ".", sep)
	}
	return timestamp
}

type CaptionSet struct {
	Styles   map[string]Style
	Captions map[string][]*Caption
}

func NewCaptionSet() *CaptionSet {
	return &CaptionSet{
		Styles:   map[string]Style{},
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

func (c CaptionSet) AddStyle(style Style) {
	c.Styles[style.ID] = style
}

func (c CaptionSet) GetStyle(id string) Style {
	if style, ok := c.Styles[id]; ok {
		return style
	}
	return DefaultStyle()
}

func (c CaptionSet) GetStyles() []Style {
	values := []Style{}
	for _, v := range c.Styles {
		values = append(values, v)
	}
	return values
}
