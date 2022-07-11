package dfxp

import (
	"encoding/xml"

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

type Paragraph struct {
	XMLName xml.Name `xml:"p"`
	Begin   string   `xml:"begin,attr"`
	End     string   `xml:"end,attr"`
	StyleID string   `xml:"style,attr"`
	Content string   `xml:",innerxml"`
	Span    *Span    `xml:",omitempty"`
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

type Span struct {
	XMLName xml.Name `xml:"span"`
	Text    string   `xml:",chardata"`
	Style
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
