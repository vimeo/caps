package dfxp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antchfx/xmlquery"
	"github.com/thiagopnts/caps"
)

const defaultLanguageCode = "en-US"

// TODO implement io.Reader interface?
type Reader struct {
	framerate  string
	multiplier []int
	timebase   string
	nodes      []caps.CaptionNode
}

func NewReader() *Reader {
	return &Reader{
		framerate:  "30",
		multiplier: []int{1, 1},
		timebase:   "media",
		nodes:      []caps.CaptionNode{},
	}
}

func (Reader) Detect(content string) bool {
	return strings.Contains(strings.ToLower(content), "</tt>")
}

func (r Reader) Read(content string) (*caps.CaptionSet, error) {
	doc, err := xmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	captions := caps.NewCaptionSet()
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
				return nil, fmt.Errorf("failed to read multiplier: %w", err)
			}
			b, err := strconv.Atoi(multipliers[1])
			if err != nil {
				return nil, fmt.Errorf("failed to read multiplier: %w", err)
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

func (r Reader) combineMatchingCaptions(captionSet *caps.CaptionSet) *caps.CaptionSet {
	for _, lang := range captionSet.GetLanguages() {
		captions := captionSet.GetCaptions(lang)
		if len(captions) <= 1 {
			return captionSet
		}
		newCaps := captions[:1]

		for _, caption := range captions[1:] {
			lastIndex := len(newCaps) - 1
			if caption.Start == newCaps[lastIndex].Start && caption.End == newCaps[lastIndex].End {
				newCaps[lastIndex].Nodes = append(newCaps[lastIndex].Nodes, caps.CreateBreak())
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

func (r Reader) translateDiv(div *xmlquery.Node) []*caps.Caption {
	captions := []*caps.Caption{}
	for _, pTag := range xmlquery.Find(div, "//p") {
		if c, err := r.translatePtag(pTag); err == nil {
			captions = append(captions, c)
		}
	}
	return captions
}

func (r *Reader) translatePtag(pTag *xmlquery.Node) (*caps.Caption, error) {
	start, end, err := r.findTimes(pTag)
	if err != nil {
		return nil, err
	}
	r.nodes = []caps.CaptionNode{}

	brs := xmlquery.Find(pTag, "//br")
	if len(brs) == 0 {
		r.translateTag(pTag)
	} else {
		child := pTag.FirstChild

		for child != nil {
			r.translateTag(child)
			child = child.NextSibling
		}
	}

	styles := r.translateStyle(pTag)
	caption := caps.NewCaption(start, end, r.nodes, styles)
	return &caption, nil
}

func (r *Reader) translateTag(tag *xmlquery.Node) {
	switch tag.Data {
	case "br":
		r.nodes = append(r.nodes, caps.CreateBreak())
	case "span":
		r.translateSpan(tag)
	case "p":
		fallthrough
	default:
		if (tag.Data == "p" && tag.FirstChild == nil && tag.Type == 2) || tag.Type == 3 {
			text := strings.TrimSpace(tag.InnerText())
			if text != "" {
				r.nodes = append(r.nodes, caps.CreateText(text))
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

func (r *Reader) translateSpan(tag *xmlquery.Node) {
	style := r.translateStyle(tag)
	captionStyle := caps.CreateCaptionStyle(true, style)
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

func (r Reader) translateStyle(tag *xmlquery.Node) caps.Style {
	style := caps.Style{}
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

func (r Reader) findTimes(root *xmlquery.Node) (int, int, error) {
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

func (r Reader) translateTime(stamp string) (int, error) {
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
