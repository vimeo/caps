package dfxp

import (
	"encoding/xml"

	"github.com/thiagopnts/caps"
)

type Writer struct {
	pStyle   bool
	openSpan bool
}

// TODO: rewrite all _recreate from python's DFXPWriter class

func (w Writer) WriteString(captions *caps.CaptionSet) (string, error) {
	bytes, err := w.Write(captions)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (w Writer) Write(captions *caps.CaptionSet) ([]byte, error) {
	st := DefaultStyle()
	for _, style := range captions.GetStyles() {
		st = NewStyle(style)
	}
	sid := st.ID
	base := NewBaseMarkup()
	base.Head = Head{Style: st, Layout: DefaultRegion()}
	for _, lang := range captions.GetLanguages() {
		divLang := Lang{Lang: lang, Ps: []Paragraph{}}
		for _, c := range captions.GetCaptions(lang) {
			if c.Style.ID != "" {
				sid = c.Style.ID
			}
			divLang.Ps = append(divLang.Ps, NewParagraph(c, sid))
		}
		base.Body.Langs = append(base.Body.Langs, divLang)
	}
	content, err := xml.Marshal(base)
	if err != nil {
		return nil, err
	}
	return content, nil
}
