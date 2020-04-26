package caps

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

type DFXPHead struct {
	Style  dfxpStyle  `xml:"styling>style"`
	Layout dfxpRegion `xml:"layout>region"`
}

type dfxpRegion struct {
	XMLName         xml.Name `xml:"region"`
	ID              string   `xml:"xml:id,attr"`
	TTSTextAlign    string   `xml:"tts:textAlign,attr,omitempty"`
	TTSDisplayAlign string   `xml:"tts:displayAlign,attr,omitempty"`
}

func dfxpDefaultRegion() dfxpRegion {
	return dfxpRegion{
		ID:              "bottom",
		TTSTextAlign:    "center",
		TTSDisplayAlign: "after",
	}
}

type pContent string

func (c pContent) MarshalText() ([]byte, error) {
	fmt.Println("MarshalText called", c)
	return []byte(c), nil
}

type dfxpP struct {
	XMLName xml.Name  `xml:"p"`
	Begin   string    `xml:"begin,attr"`
	End     string    `xml:"end,attr"`
	StyleID string    `xml:"style,attr"`
	Content string    `xml:",innerxml"`
	Span    *dfxpSpan `xml:",omitempty"`
}

const brTag = "<br/>"

func NewDFXPp(caption *Caption, s string) dfxpP {
	start := caption.FormatStart()
	end := caption.FormatEnd()
	line := ""
	hasSpan := false
	var sp dfxpSpan
	for _, node := range caption.Nodes {
		if node.Kind() == text {
			buf := bytes.Buffer{}
			xml.Escape(&buf, []byte(node.GetContent()))
			line += buf.String()
		} else if node.Kind() == lineBreak {
			line += "<br/>"
		} else if node.Kind() == style {
			sp = NewDFXPspan(line, NewDFXPStyle(node.(captionStyle).Style))
			hasSpan = true
		}
		hasSpan = false
	}
	if hasSpan {
		return dfxpP{
			Begin:   start,
			End:     end,
			StyleID: s,
			Span:    &sp,
		}
	}

	return dfxpP{
		Begin:   start,
		End:     end,
		StyleID: s,
		Content: line,
	}
}

type dfxpLang struct {
	XMLName xml.Name `xml:"div"`
	Lang    string   `xml:"xml:lang,attr"`
	Ps      []dfxpP
}

type dfxpBody struct {
	XMLName xml.Name `xml:"body"`
	Langs   []dfxpLang
}

type DFXPBaseMarkup struct {
	XMLName    xml.Name `xml:"tt"`
	TtXMLLang  string   `xml:"xml:lang,attr" default:"en"`
	TtXMLns    string   `xml:"xmlns,attr" default:"http://www.w3.org/ns/ttml"`
	TtXMLnsTTS string   `xml:"xmlns:tts,attr" default:"http://www.w3.org/ns/ttml#styling"`
	Head       DFXPHead `xml:"head"`
	Body       dfxpBody `xml:"body"`
}

func NewDFXPBaseMarkup() DFXPBaseMarkup {
	return DFXPBaseMarkup{
		TtXMLLang:  "en",
		TtXMLns:    "http://www.w3.org/ns/ttml",
		TtXMLnsTTS: "http://www.w3.org/ns/ttml#styling",
	}
}

type DFXPWriter struct {
	pStyle   bool
	openSpan bool
}

type dfxpSpan struct {
	XMLName xml.Name `xml:"span"`
	Text    string   `xml:",chardata"`
	dfxpStyle
}

func NewDFXPspan(s string, style dfxpStyle) dfxpSpan {
	return dfxpSpan{xml.Name{}, s, style}
}

type dfxpStyle struct {
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

func dfxpDefaultStyle() dfxpStyle {
	return dfxpStyle{
		ID:            "default",
		TTSColor:      "white",
		TTSFontFamily: "monospace",
		TTSFontSize:   "1c",
	}
}

func NewDFXPStyle(style Style) dfxpStyle {
	fontStyle := ""
	if style.Italics {
		fontStyle = "italic"
	}
	fontWeight := ""
	if style.Bold {
		fontWeight = "bold"
	}
	return dfxpStyle{
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

//TODO implement XML Marshal grabing values from embeded Style.

func NewDFXPWriter() DFXPWriter {
	return DFXPWriter{
		false,
		false,
	}
}

// TODO: rewrite all _recreate from python's DFXPWriter class

func (w DFXPWriter) Write(captions *CaptionSet) (DFXPBaseMarkup, error) {
	st := dfxpDefaultStyle()
	for _, style := range captions.GetStyles() {
		st = NewDFXPStyle(style)
	}
	sid := st.ID
	base := NewDFXPBaseMarkup()
	base.Head = DFXPHead{Style: st, Layout: dfxpDefaultRegion()}
	for _, lang := range captions.GetLanguages() {
		divLang := dfxpLang{Lang: lang, Ps: []dfxpP{}}
		for _, c := range captions.GetCaptions(lang) {
			if c.Style.ID != "" {
				sid = c.Style.ID
			}
			divLang.Ps = append(divLang.Ps, NewDFXPp(c, sid))
		}
		base.Body.Langs = append(base.Body.Langs, divLang)
	}
	return base, nil
}
