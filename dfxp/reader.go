package dfxp

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/vimeo/caps"
)

type reader struct {
	framerate  string
	multiplier []int
	timebase   string
	nodes      []caps.CaptionContent
}

func (r reader) Detect(content []byte) bool {
	validXML := xml.Unmarshal(content, new(interface{})) == nil
	return strings.Contains(strings.ToLower(string(content)), "</tt>") && validXML
}

func (r reader) Read(content []byte) (*caps.CaptionSet, error) {
	doc, err := xmlquery.Parse(strings.NewReader(string(content)))
	if err != nil {
		return nil, err
	}

	captions := caps.NewCaptionSet()
	tts := xmlquery.Find(doc, "/tt")
	if len(tts) >= 1 {
		tt := tts[0]
		r.timebase = "0"
		if timebase := tt.SelectAttr("ttp:timebase"); timebase != "" {
			r.timebase = timebase
		}
		r.framerate = "30"
		if framerate := tt.SelectAttr("ttp:framerate"); framerate != "" {
			r.framerate = framerate
		}

		r.multiplier = []int{1, 1}
		if multiplier := tt.SelectAttr("ttp:framemultiplier"); multiplier != "" {
			multipliers := strings.Split(multiplier, " ")
			a, err := strconv.Atoi(multipliers[0])
			if err != nil {
				return nil, fmt.Errorf("failed to read multiplier: %w", err)
			}
			b, err := strconv.Atoi(multipliers[1])
			if err != nil {
				return nil, fmt.Errorf("failed to read multiplier: %w", err)
			}
			r.multiplier = []int{a, b}
		}
	}
	for _, div := range xmlquery.Find(doc, "//div") {
		lang := div.SelectAttr("xml:lang")
		if lang == "" {
			lang = caps.DefaultLang
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

func (r reader) combineMatchingCaptions(captionSet *caps.CaptionSet) *caps.CaptionSet {
	for _, lang := range captionSet.Languages() {
		captions := captionSet.GetCaptions(lang)
		if len(captions) <= 1 {
			return captionSet
		}
		newCaps := captions[:1]

		for _, caption := range captions[1:] {
			lastIndex := len(newCaps) - 1
			if caption.Start == newCaps[lastIndex].Start && caption.End == newCaps[lastIndex].End {
				newCaps[lastIndex].Nodes = append(newCaps[lastIndex].Nodes, caps.NewLineBreak())
				newCaps[lastIndex].Nodes = append(newCaps[lastIndex].Nodes, caption.Nodes...)
				continue
			}
			newCaps = append(newCaps, caption)
			captionSet.SetCaptions(lang, newCaps)
		}
	}
	return captionSet
}

func (r reader) translateDiv(div *xmlquery.Node) []*caps.Caption {
	captions := []*caps.Caption{}
	for _, pTag := range xmlquery.Find(div, "//p") {
		if start, end, err := r.findTimes(div); err == nil {
			captions = append(captions, r.translateParentTimedParagraph(pTag, start, end))
		} else if c, err := r.translatePtag(pTag); err == nil {
			captions = append(captions, c)
		}
	}
	return captions
}

func (r *reader) translateParentTimedParagraph(paragraph *xmlquery.Node, start, end int) *caps.Caption {
	r.nodes = []caps.CaptionContent{}

	brs := xmlquery.Find(paragraph, "//br")
	if len(brs) == 0 {
		r.translateTag(paragraph)
	} else {
		child := paragraph.FirstChild

		for child != nil {
			r.translateTag(child)
			child = child.NextSibling
		}
	}

	styles := r.translateStyle(paragraph)
	caption := caps.NewCaption(float64(start), float64(end), r.nodes, styles)
	return &caption
}

func (r *reader) translatePtag(paragraph *xmlquery.Node) (*caps.Caption, error) {
	start, end, err := r.findTimes(paragraph)
	if err != nil {
		return nil, err
	}
	captions := r.translateParentTimedParagraph(paragraph, start, end)
	return captions, nil
}

func (r *reader) translateTag(tag *xmlquery.Node) {
	switch tag.Data {
	case "br":
		r.nodes = append(r.nodes, caps.NewLineBreak())
	case "span":
		r.translateSpan(tag)
	case "p":
		fallthrough
	default:
		if (tag.Data == "p" && tag.FirstChild == nil && tag.Type == 2) || tag.Type == 3 {
			text := strings.TrimSpace(tag.InnerText())
			if text != "" {
				r.nodes = append(r.nodes, caps.NewCaptionText(text))
			}
		} else {
			child := tag.FirstChild
			for child != nil {
				r.translateTag(child)
				child = child.NextSibling
			}
		}
	}
}

func (r *reader) translateSpan(tag *xmlquery.Node) {
	style := r.translateStyle(tag)
	captionStyle := caps.NewCaptionStyle(true, style)
	r.nodes = append(r.nodes, captionStyle)
	// for some reason xmlquery.Find(tag, "child::*") doesnt work here
	child := tag.FirstChild
	for child != nil {
		r.translateTag(child)
		child = child.NextSibling
	}
	//	secondStyle := CreateCaptionStyle(false, style)
	//	r.nodes = append(r.nodes, secondStyle)
	//	return // <- this return was uncommented
	// FIXME this is duped, porting as is for now
	//	for _, child := range xmlquery.Find(tag, "child::*") {
	//r.translateTag(child)
	//	}
}

func (r reader) translateStyle(tag *xmlquery.Node) caps.StyleProps {
	style := caps.StyleProps{}
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

func (r reader) findTimes(root *xmlquery.Node) (int, int, error) {
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

func (r reader) translateTime(stamp string) (int, error) {
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
