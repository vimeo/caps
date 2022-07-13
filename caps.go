package caps

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

const DefaultLang = "en-US"

type CaptionReader interface {
	Read([]byte) (*CaptionSet, error)
	Detect([]byte) bool
}

type CaptionWriter interface {
	Write(*CaptionSet) ([]byte, error)
}

type CaptionContent interface {
	Text() bool
	Style() bool
	LineBreak() bool
	Content() string
}

type isNot struct{}

func (isNot) Text() bool {
	return false
}

func (isNot) LineBreak() bool {
	return false
}

func (isNot) Style() bool {
	return false
}

type CaptionText struct {
	content string
	isNot
}

func (CaptionText) Text() bool {
	return true
}

func NewCaptionText(text string) CaptionContent {
	return CaptionText{text, isNot{}}
}

func (c CaptionText) Content() string {
	return c.content
}

type CaptionStyle struct {
	Props StyleProps
	Start bool
	isNot
}

func (c CaptionStyle) Style() bool {
	return true
}

func (c CaptionStyle) Content() string {
	return c.Props.String()
}

func NewCaptionStyle(start bool, style StyleProps) CaptionContent {
	return CaptionStyle{style, start, isNot{}}
}

type CaptionLineBreak struct{ isNot }

func (c CaptionLineBreak) LineBreak() bool {
	return true
}

func (c CaptionLineBreak) Content() string {
	return "\n"
}

func NewLineBreak() CaptionContent {
	return CaptionLineBreak{isNot{}}
}

const defaultStyleID = "default"

// FIXME This is a simple placeholder for style types, this can be better represented
// but I need to implement more caption types first(this was written with just dfxp)
type StyleProps struct {
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

func (s StyleProps) String() string {
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

func DefaultStyleProps() StyleProps {
	return StyleProps{Color: "white", FontFamily: "monospace", FontSize: "1c"}
}

type CaptionSet struct {
	Styles   map[string]StyleProps
	Captions map[string][]*Caption
}

func NewCaptionSet() *CaptionSet {
	return &CaptionSet{
		Styles:   map[string]StyleProps{},
		Captions: map[string][]*Caption{},
	}
}

func (c CaptionSet) SetCaptions(lang string, captions []*Caption) {
	c.Captions[lang] = captions
}

func (c CaptionSet) Languages() []string {
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

func (c CaptionSet) AddStyle(style StyleProps) {
	c.Styles[style.ID] = style
}

func (c CaptionSet) GetStyle(id string) StyleProps {
	if style, ok := c.Styles[id]; ok {
		return style
	}
	return DefaultStyleProps()
}

func (c CaptionSet) GetStyles() []StyleProps {
	values := []StyleProps{}
	for _, v := range c.Styles {
		values = append(values, v)
	}
	return values
}

func SplitLines(s string) []string {
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}
