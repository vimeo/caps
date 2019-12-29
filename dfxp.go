package caps

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
)

const defaultLanguageCode = "en-US"

type DFXReader struct {
	framerate  string
	multiplier []int
	timebase   string
	nodes      []captionNode
}

func NewDFXReader() DFXReader {
	return DFXReader{
		framerate:  "30",
		multiplier: []int{1, 1},
		timebase:   "media",
		nodes:      []captionNode{},
	}
}

func (DFXReader) Detect(content string) bool {
	return strings.Contains(strings.ToLower(content), "</tt>")
}

func (r DFXReader) Read(content string) (*CaptionSet, error) {
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
		captions.AddStyle(id, r.translateStyle(style))
	}
	captions = r.combineMatchingCaptions(captions)
	if captions.IsEmpty() {
		return captions, fmt.Errorf("empty caption file")
	}
	return captions, nil
}

func (r DFXReader) combineMatchingCaptions(captionSet *CaptionSet) *CaptionSet {
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

func (r DFXReader) translateDiv(div *xmlquery.Node) []*Caption {
	captions := []*Caption{}
	for _, pTag := range xmlquery.Find(div, "//p") {
		if c, err := r.translatePtag(pTag); err == nil {
			captions = append(captions, c)
		}
	}
	return captions
}

func (r DFXReader) translatePtag(pTag *xmlquery.Node) (*Caption, error) {
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

func (r DFXReader) translateTag(tag *xmlquery.Node) {
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

func (r DFXReader) translateSpan(tag *xmlquery.Node) {
	styleAttrs := r.translateStyle(tag)
	if len(styleAttrs) > 0 {
		style := CreateStyle(true, styleAttrs)
		r.nodes = append(r.nodes, style)
		for _, child := range xmlquery.Find(tag, "child::*") {
			r.translateTag(child)
		}
		secondStyle := CreateStyle(false, styleAttrs)
		r.nodes = append(r.nodes, secondStyle)
		return
	}
	// FIXME this is duped, porting as is for now
	for _, child := range xmlquery.Find(tag, "child::*") {
		r.translateTag(child)
	}
}

//FIXME: since this seems to have its own limit set of supported props,
// return its own type instead of map[string]string, so we can have same defaults.
func (r DFXReader) translateStyle(tag *xmlquery.Node) map[string]string {
	attrs := map[string]string{}
	for _, attr := range tag.Attr {
		switch strings.ToLower(attr.Name.Local) {
		case "style":
			attrs["class"] = attr.Value
		case "fontstyle":
			if attr.Value == "italic" {
				attrs["italics"] = "true"
			}
		case "textalign":
			attrs["text-align"] = attr.Value
		case "fontfamily":
			attrs["font-family"] = attr.Value
		case "fontsize":
			attrs["font-size"] = attr.Value
		case "color":
			attrs["color"] = attr.Value
		case "fontweight":
			if attr.Value == "bold" {
				attrs["bold"] = "true"
			}
		case "textdecoration":
			if attr.Value == "underline" {
				attrs["underline"] = "true"
			}
		}
	}
	return attrs
}

func (r DFXReader) findTimes(root *xmlquery.Node) (int, int, error) {
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

func (r DFXReader) translateTime(stamp string) (int, error) {
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
