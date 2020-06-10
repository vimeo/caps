package dfxp

import (
	"bytes"
	"encoding/xml"
	"strings"

	"github.com/thiagopnts/caps"
)

func NewReader() caps.CaptionReader {
	return &Reader{
		framerate:  "30",
		multiplier: []int{1, 1},
		timebase:   "media",
		nodes:      []caps.CaptionContent{},
	}
}

func NewWriter() caps.CaptionWriter {
	return Writer{
		false,
		false,
	}
}

type Head struct {
	Style  Style  `xml:"styling>style"`
	Layout Region `xml:"layout>region"`
}

type Region struct {
	XMLName         xml.Name `xml:"region"`
	ID              string   `xml:"xml:id,attr"`
	TTSTextAlign    string   `xml:"tts:textAlign,attr,omitempty"`
	TTSDisplayAlign string   `xml:"tts:displayAlign,attr,omitempty"`
}

func DefaultRegion() Region {
	return Region{
		ID:              "bottom",
		TTSTextAlign:    "center",
		TTSDisplayAlign: "after",
	}
}

type Paragraph struct {
	XMLName xml.Name `xml:"p"`
	Begin   string   `xml:"begin,attr"`
	End     string   `xml:"end,attr"`
	StyleID string   `xml:"style,attr"`
	Content string   `xml:",innerxml"`
	Span    *Span    `xml:",omitempty"`
}

const brTag = "<br/>"

func NewParagraph(caption *caps.Caption, s string) Paragraph {
	start := caption.FormatStart()
	end := caption.FormatEnd()
	line := ""
	var sp *Span

	for _, node := range caption.Nodes {
		if node.IsText() && sp == nil {
			buf := bytes.Buffer{}
			xml.Escape(&buf, []byte(node.GetContent()))
			str := buf.String()
			str = strings.ReplaceAll(str, `&#39;`, `'`)
			str = strings.ReplaceAll(str, `&#34;`, `"`)
			str = strings.ReplaceAll(str, `&#xA;`, ``)
			line += str
		} else if node.IsLineBreak() && sp == nil {
			line += "<br/>"
		} else if node.IsStyle() && sp == nil {
			sp = NewSpan(line, NewStyle(node.(caps.CaptionStyle).Style))
		} else if sp != nil {
			// FIXME do all the strings.ReplaceAll here too
			line += node.GetContent()
			sp.Text += line
		}
	}
	if sp != nil {
		return Paragraph{
			Begin:   start,
			End:     end,
			StyleID: s,
			Span:    sp,
		}
	}

	return Paragraph{
		Begin:   start,
		End:     end,
		StyleID: s,
		Content: line,
	}
}

type Lang struct {
	XMLName xml.Name `xml:"div"`
	Lang    string   `xml:"xml:lang,attr"`
	Ps      []Paragraph
}

type Body struct {
	XMLName xml.Name `xml:"body"`
	Langs   []Lang
}

type BaseMarkup struct {
	XMLName    xml.Name `xml:"tt"`
	TtXMLLang  string   `xml:"xml:lang,attr" default:"en"`
	TtXMLns    string   `xml:"xmlns,attr" default:"http://www.w3.org/ns/ttml"`
	TtXMLnsTTS string   `xml:"xmlns:tts,attr" default:"http://www.w3.org/ns/ttml#styling"`
	Head       Head     `xml:"head"`
	Body       Body     `xml:"body"`
}

func NewBaseMarkup() BaseMarkup {
	return BaseMarkup{
		TtXMLLang:  "en",
		TtXMLns:    "http://www.w3.org/ns/ttml",
		TtXMLnsTTS: "http://www.w3.org/ns/ttml#styling",
	}
}

type Span struct {
	XMLName xml.Name `xml:"span"`
	Text    string   `xml:",chardata"`
	Style
}

func NewSpan(s string, style Style) *Span {
	return &Span{xml.Name{}, s, style}
}

type Style struct {
	XMLName       xml.Name `xml:"style"`
	ID            string   `xml:"xml:id,attr,omitempty"`
	TTSTextAlign  string   `xml:"tts:textAlign,attr,omitempty"`
	TTSFontStyle  string   `xml:"tts:fontStyle,attr,omitempty"`
	TTSFontFamily string   `xml:"tts:fontFamily,attr,omitempty"`
	TTSFontSize   string   `xml:"tts:fontSize,attr,omitempty"`
	TTSFontWeight string   `xml:"tts:fontweight,attr,omitempty"`
	TTSColor      string   `xml:"tts:color,attr,omitempty"`
	// FIXME this is never parsed to Style
	TTSDisplayAlign string `xml:"tts:displayAlign,attr,omitempty"`
}

func DefaultStyle() Style {
	return Style{
		ID:            "default",
		TTSColor:      "white",
		TTSFontFamily: "monospace",
		TTSFontSize:   "1c",
	}
}

func NewStyle(style caps.StyleProps) Style {
	fontStyle := ""
	if style.Italics {
		fontStyle = "italic"
	}
	fontWeight := ""
	if style.Bold {
		fontWeight = "bold"
	}
	return Style{
		ID:            style.ID,
		TTSTextAlign:  style.TextAlign,
		TTSFontStyle:  fontStyle,
		TTSFontFamily: style.FontFamily,
		TTSFontSize:   style.FontSize,
		TTSColor:      style.Color,
		TTSFontWeight: fontWeight,
		// FIXME this is never parsed to Style
		TTSDisplayAlign: "",
	}
}
