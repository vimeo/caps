package dfxp

import (
	"bytes"
	"encoding/xml"
	"strings"

	"github.com/thiagopnts/caps"
)

type Writer struct {
	pStyle   bool
	openSpan bool
}

// TODO: rewrite all _recreate from python's DFXPWriter class

func (w Writer) Write(captions *caps.CaptionSet) ([]byte, error) {
	st := defaultStyle()
	for _, style := range captions.GetStyles() {
		st = newStyle(style)
	}
	sid := st.ID
	base := newBaseMarkup()
	base.Head = Head{
		Style:  st,
		Layout: defaultRegion(),
	}
	for _, lang := range captions.Languages() {
		divLang := Lang{
			Lang: lang,
			Ps:   []Paragraph{},
		}
		for _, c := range captions.GetCaptions(lang) {
			if c.Style.ID != "" {
				sid = c.Style.ID
			}
			divLang.Ps = append(divLang.Ps, newParagraph(c, sid))
		}
		base.Body.Langs = append(base.Body.Langs, divLang)
	}
	content, err := xml.Marshal(base)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func newStyle(style caps.StyleProps) Style {
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

func defaultStyle() Style {
	return Style{
		ID:            "default",
		TTSColor:      "white",
		TTSFontFamily: "monospace",
		TTSFontSize:   "1c",
	}
}

func defaultRegion() Region {
	return Region{
		ID:              "bottom",
		TTSTextAlign:    "center",
		TTSDisplayAlign: "after",
	}
}

func newSpan(s string, style Style) *Span {
	return &Span{xml.Name{}, s, style}
}

func newParagraph(caption *caps.Caption, s string) Paragraph {
	start := caption.FormatStart()
	end := caption.FormatEnd()
	line := ""
	var sp *Span

	for _, node := range caption.Nodes {
		if node.Text() && sp == nil {
			buf := bytes.Buffer{}
			xml.Escape(&buf, []byte(node.Content()))
			str := buf.String()
			str = strings.ReplaceAll(str, `&#39;`, `'`)
			str = strings.ReplaceAll(str, `&#34;`, `"`)
			str = strings.ReplaceAll(str, `&#xA;`, ``)
			line += str
		} else if node.LineBreak() && sp == nil {
			line += "<br/>"
		} else if node.Style() && sp == nil {
			sp = newSpan(line, newStyle(node.(caps.CaptionStyle).Props))
		} else if sp != nil {
			// FIXME do all the strings.ReplaceAll here too
			line += node.Content()
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

func newBaseMarkup() BaseMarkup {
	return BaseMarkup{
		TtXMLLang:  "en",
		TtXMLns:    "http://www.w3.org/ns/ttml",
		TtXMLnsTTS: "http://www.w3.org/ns/ttml#styling",
	}
}
