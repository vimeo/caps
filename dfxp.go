package caps

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
)

const defaultLanguageCode = "en-US"

type DFXPReader struct {
	framerate  string
	multiplier []int
	timebase   string
	nodes      []captionNode
}

func NewDFXPReader() DFXPReader {
	return DFXPReader{
		framerate:  "30",
		multiplier: []int{1, 1},
		timebase:   "media",
		nodes:      []captionNode{},
	}
}

func (DFXPReader) Detect(content string) bool {
	return strings.Contains(strings.ToLower(content), "</tt>")
}

func (r DFXPReader) Read(content string) (*CaptionSet, error) {
	doc, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	captions := NewCaptionSet()
	tts := xmlquery.Find(doc, "/tt")
	if len(tts) >= 1 {
		tt := tts[0]
		if timebase := tt.SelectAttr("ttp:timebase"); timebase != "" {
			r.timebase = timebase
		} else {
			r.timebase = "0"
		}

		if framerate := tt.SelectAttr("ttp:framerate"); framerate != "" {
			r.framerate = framerate
		} else {
			r.framerate = "30"
		}

		if multiplier := tt.SelectAttr("ttp:framemultiplier"); multiplier != "" {
			multipliers := strings.Split(multiplier, " ")
			a, err := strconv.Atoi(multipliers[0])
			if err != nil {
				log.Fatalln("failed to read multiplier")
			}
			b, err := strconv.Atoi(multipliers[1])
			if err != nil {
				log.Fatalln("failed to read multiplier")
			}
			r.multiplier = []int{a, b}
		} else {
			r.multiplier = []int{1, 1}
		}
	}
	for _, div := range xmlquery.Find(doc, "//div") {
		lang := div.SelectAttr("xml:lang")
		if lang == "" {
			lang = defaultLanguageCode
		}
		captions.SetCaptions(lang, r.translateDiv(div))
	}

	for _, style := range xmlquery.Find(doc, "//style") {
		id := style.SelectAttr("id")
		if id == "" {
			id = style.SelectAttr("xml:id")
		}
		parsedStyle := r.translateStyle(style)
		parsedStyle.ID = id
		captions.AddStyle(parsedStyle)

	}
	captions = r.combineMatchingCaptions(captions)
	if captions.IsEmpty() {
		return captions, fmt.Errorf("empty caption file")
	}
	return captions, nil
}

func (r DFXPReader) combineMatchingCaptions(captionSet *CaptionSet) *CaptionSet {
	for _, lang := range captionSet.GetLanguages() {
		captions := captionSet.GetCaptions(lang)
		if len(captions) <= 1 {
			return captionSet
		}
		newCaps := captions[:1]

		for _, caption := range captions[1:] {
			lastIndex := len(newCaps) - 1
			if caption.Start == newCaps[lastIndex].Start && caption.End == newCaps[lastIndex].End {
				newCaps[lastIndex].Nodes = append(newCaps[lastIndex].Nodes, CreateBreak())
				for _, node := range caption.Nodes {
					newCaps[lastIndex].Nodes = append(newCaps[lastIndex].Nodes, node)
				}
				continue
			}
			newCaps = append(newCaps, caption)
			captionSet.SetCaptions(lang, newCaps)
		}
	}
	return captionSet
}

func (r DFXPReader) translateDiv(div *xmlquery.Node) []*Caption {
	captions := []*Caption{}
	for _, pTag := range xmlquery.Find(div, "//p") {
		if c, err := r.translatePtag(pTag); err == nil {
			captions = append(captions, c)
		}
	}
	return captions
}

func (r DFXPReader) translatePtag(pTag *xmlquery.Node) (*Caption, error) {
	start, end, err := r.findTimes(pTag)
	if err != nil {
		return nil, err
	}
	r.nodes = []captionNode{}
	r.translateTag(pTag)
	styles := r.translateStyle(pTag)
	caption := NewCaption(start, end, r.nodes, styles)
	return &caption, nil
}

func (r DFXPReader) translateTag(tag *xmlquery.Node) {
	switch tag.Data {
	case "p":
		text := strings.TrimSpace(tag.InnerText())
		if text != "" {
			r.nodes = append(r.nodes, CreateText(text))
		}
	case "br":
		r.nodes = append(r.nodes, CreateBreak())
	case "span":
		r.translateSpan(tag)
	default:
		for _, child := range xmlquery.Find(tag, "child::*") {
			r.translateTag(child)
		}
	}
}

func (r DFXPReader) translateSpan(tag *xmlquery.Node) {
	style := r.translateStyle(tag)
	captionStyle := CreateCaptionStyle(true, style)
	r.nodes = append(r.nodes, captionStyle)
	for _, child := range xmlquery.Find(tag, "child::*") {
		r.translateTag(child)
	}
	secondStyle := CreateCaptionStyle(false, style)
	r.nodes = append(r.nodes, secondStyle)
	return
	// FIXME this is duped, porting as is for now
	for _, child := range xmlquery.Find(tag, "child::*") {
		r.translateTag(child)
	}
}

//FIXME: since this seems to have its own limit set of supported props,
// return its own type instead of map[string]string, so we can have same defaults.
func (r DFXPReader) translateStyle(tag *xmlquery.Node) Style {
	style := Style{}
	for _, attr := range tag.Attr {
		switch strings.ToLower(attr.Name.Local) {
		case "style":
			style.Class = attr.Value
		case "fontstyle":
			style.Italics = attr.Value == "italic"
		case "textalign":
			style.TextAlign = attr.Value
		case "fontfamily":
			style.FontFamily = attr.Value
		case "fontsize":
			style.FontSize = attr.Value
		case "color":
			style.Color = attr.Value
		case "fontweight":
			style.Bold = attr.Value == "bold"
		case "textdecoration":
			style.Underline = attr.Value == "underline"
		}
	}
	return style
}

func (r DFXPReader) findTimes(root *xmlquery.Node) (int, int, error) {
	begin := root.SelectAttr("begin")
	if begin == "" {
		return 0, 0, fmt.Errorf("tag doesnt have a time begin")
	}
	start, err := r.translateTime(begin)
	if err != nil {
		return 0, 0, err
	}
	endValue := root.SelectAttr("end")
	if endValue == "" {
		if dur := root.SelectAttr("dur"); dur != "" {
			durParsed, err := r.translateTime(dur)
			if err != nil {
				return 0, 0, err
			}
			return start, start + durParsed, nil
		}
	}
	end, err := r.translateTime(endValue)
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

func (r DFXPReader) translateTime(stamp string) (int, error) {
	timesplit := strings.Split(stamp, ":")
	if !strings.Contains(timesplit[2], ".") {
		timesplit[2] = timesplit[2] + ".000"
	}
	timesplit0, err := strconv.Atoi(timesplit[0])
	if err != nil {
		return 0, err
	}
	timesplit1, err := strconv.Atoi(timesplit[1])
	if err != nil {
		return 0, err
	}
	secsplit := strings.Split(timesplit[2], ".")
	secsplit0, err := strconv.Atoi(secsplit[0])
	if err != nil {
		return 0, err
	}
	secsplit1, err := strconv.Atoi(secsplit[1])
	if err != nil {
		return 0, err
	}
	if len(timesplit) > 3 {
		timesplit3, err := strconv.ParseFloat(timesplit[3], 32)
		if err != nil {
			return 0, err
		}
		framerate, err := strconv.ParseFloat(r.framerate, 32)
		if err != nil {
			return 0, err
		}
		if r.timebase == "smpte" {
			secsplit1 = int(timesplit3 / framerate * 1000.0)
		} else {
			secsplit1 = int(float64(int(timesplit3)*r.multiplier[1]) / framerate * float64(r.multiplier[0]) * 1000.0)
		}
	}
	microseconds := int(timesplit0)*3600000000 +
		int(timesplit1)*60000000 +
		int(secsplit0)*1000000 +
		int(secsplit1)*1000

	if r.timebase == "smpte" {
		return int(float64(microseconds) * float64(r.multiplier[1]) / float64(r.multiplier[0])), nil
	}

	return microseconds, nil
}

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

type dfxpP struct {
	XMLName xml.Name `xml:"p"`
	Begin   string   `xml:"begin,attr"`
	End     string   `xml:"end,attr"`
	StyleID string   `xml:"style,attr"`
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

type dfxpStyle struct {
	XMLName       xml.Name `xml:"style"`
	ID            string   `xml:"xml:id,attr"`
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
	base := NewDFXPBaseMarkup()
	base.Head = DFXPHead{Style: st, Layout: dfxpDefaultRegion()}
	for _, lang := range captions.GetLanguages() {
		divLang := dfxpLang{Lang: lang, Ps: []dfxpP{}}
		for _, caption := range captions.GetCaptions(lang) {
			s := dfxpDefaultStyle()
			//			if caption.Style.ID != "" {
			//				s = caption.Style
			//			}
			divLang.Ps = append(divlang.Ps, dfxpP{BegingP})
		}
	}
	return base, nil
}
