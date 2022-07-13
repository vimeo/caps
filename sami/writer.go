package sami

import (
	"fmt"
	"html"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/vimeo/caps"
)

const (
	baseMarkup = `
<html>
	<head>
		<style type="text/css"></style>
	</head>
	<body></body>
</html>`
)

type Writer struct {
	openSpan bool
	lastTime float64
}

func (w *Writer) Write(cs *caps.CaptionSet) ([]byte, error) {
	sami, err := goquery.NewDocumentFromReader(strings.NewReader(baseMarkup))
	if err != nil {
		return nil, err
	}
	st := w.recreateStyleSheet(cs)
	sami.Find("style").First().SetHtml(st)
	langs := cs.Languages()
	for _, lang := range langs {
		w.lastTime = -1
		for _, c := range cs.GetCaptions(lang) {
			w.recreateParagraph(c, sami, lang, langs[0], cs)
		}
	}
	// NOTE: skipping
	//a = sami.prettify(formatter=None).split(u'\n')
	//caption_content = u'\n'.join(a[1:])

	// 😬😬😬
	sami.Find("html").Nodes[0].Data = "sami"
	content, err := sami.Html()
	if err != nil {
		return nil, fmt.Errorf("error building caption content: %s", err)
	}
	return []byte(content), nil
}

func (w *Writer) recreateParagraphLang(c *caps.Caption, lang string, cs *caps.CaptionSet) string {
	if s := cs.GetStyle(c.Style.Class); s.Lang != "" {
		return c.Style.Class
	}
	return lang
}

func (w *Writer) recreateParagraph(c *caps.Caption, doc *goquery.Document, lang string, primaryLang string, cs *caps.CaptionSet) {
	time := *c.Start / 1000
	if w.lastTime != -1 && time != w.lastTime {
		w.recreateBlankTag(c, doc, lang, primaryLang, cs)
	}
	w.lastTime = *c.End / 1000
	sync := w.recreateSync(c, doc, lang, primaryLang, cs, int64(time))
	class := w.recreateParagraphLang(c, lang, cs)
	text := w.recreateText(c)
	var style strings.Builder
	for attr, value := range w.recreateStyle(c.Style) {
		style.WriteString(fmt.Sprintf("%s:%s;", attr, value))
	}
	p := fmt.Sprintf(`<p class="%s">%s</p>`, class, text)
	ps := style.String()
	if ps != "" {
		p = fmt.Sprintf(`<p class="%s" p_style="%s">%s</p>`, class, ps, text)
	}
	sync.AppendHtml(p)
}

func (w *Writer) recreateText(cc *caps.Caption) string {
	var t strings.Builder
	for _, n := range cc.Nodes {
		if n.Text() {
			t.WriteString(html.EscapeString(n.Content()))
			continue
		}
		if n.LineBreak() {
			t.WriteString("<br/>\n    ")
			continue
		}
		if n.Style() {
			w.recreateLineStyle(&t, cc)
			continue
		}
	}
	return strings.TrimRight(t.String(), " ")
}

func (w *Writer) recreateLineStyle(s *strings.Builder, c *caps.Caption) {
	if c.Start != nil {
		if w.openSpan {
			s.WriteString("</span> ")
		}
		w.recreateSpan(s, c.Style)
	} else {
		if w.openSpan {
			s.WriteString("</span> ")
			w.openSpan = false
		}
	}
}
func (w *Writer) recreateSpan(s *strings.Builder, sp caps.StyleProps) {
	var (
		style strings.Builder
		class string
	)
	if sp.Class != "" {
		class = fmt.Sprintf(` class="%s"`, sp.Class)
	}
	for attr, value := range w.recreateStyle(sp) {
		style.WriteString(fmt.Sprintf(`%s:%s;`, attr, value))
	}
	st := style.String()
	if class != "" {
		if st != "" {
			st = fmt.Sprintf(` style="%s"`, st)
		}
		fmt.Fprintf(s, `<span%s%s>`, class, st)
		w.openSpan = true
	}
}

func (w *Writer) recreateSync(c *caps.Caption, doc *goquery.Document, lang string, primaryLang string, cs *caps.CaptionSet, time int64) *goquery.Selection {
	query := fmt.Sprintf(`sync[start="%d"]`, time)
	if lang == primaryLang {
		sync := fmt.Sprintf(`<sync start="%d"></sync>`, time)
		doc.Find("body").First().AppendHtml(sync)
		return doc.Find("body").First().Find(query).Last()
	} else {
		sync := doc.Find(fmt.Sprintf(`sync[start="%d"]`, time)).First()
		if sync.Length() > 0 {
			return sync
		}
		return w.findClosestSync(doc, time)
	}
	return nil
}

func (w *Writer) findClosestSync(doc *goquery.Document, time int64) *goquery.Selection {
	sync := fmt.Sprintf(`<sync start="%d">`, time)
	earlier := []*goquery.Selection{}
	doc.Find("sync").Each(func(_ int, s *goquery.Selection) {
		start, _ := strconv.ParseInt(s.AttrOr("start", "0"), 10, 64)
		if start < time {
			earlier = append(earlier, s)
		}
	})
	if len(earlier) > 0 {
		lastSync := earlier[len(earlier)-1]
		lastSync.AfterHtml(sync)
		return lastSync.Next()
	}

	later := []*goquery.Selection{}
	doc.Find("sync").Each(func(_ int, s *goquery.Selection) {
		start, _ := strconv.ParseInt(s.AttrOr("start", "0"), 10, 64)
		if start > time {
			later = append(later, s)
		}
	})
	if len(later) > 0 {
		lastSync := later[0]
		lastSync.BeforeHtml(sync)
		return lastSync.Prev()
	}
	return nil
}

func (w *Writer) recreateBlankTag(c *caps.Caption, doc *goquery.Document, lang string, primaryLang string, cs *caps.CaptionSet) {
	sync := w.recreateSync(c, doc, lang, primaryLang, cs, int64(w.lastTime))
	if sync == nil {
		return
	}
	class := w.recreateParagraphLang(c, lang, cs)
	sync.AppendHtml(fmt.Sprintf(`<p class="%s">&nbsp;</p>`, class))
}

func (w *Writer) recreateStyleSheet(cs *caps.CaptionSet) string {
	var st strings.Builder
	st.WriteString("<!--")
	for el, style := range cs.Styles {
		st.WriteString(w.recreateStyleTag(el, style))
	}

	currStyles := st.String()
	for _, lang := range cs.Languages() {
		if !strings.Contains(currStyles, "lang: "+lang) {
			st.WriteString(fmt.Sprintf("\n    .%s {\n     lang: %s;\n    }\n", lang, lang))
		}
	}
	st.WriteString("   -->")
	return st.String()
}

func (w *Writer) recreateStyleTag(el string, s caps.StyleProps) string {
	target := ""
	if el != "p" && el != "sync" && el != "span" {
		target = "."
	}
	var samiStyle strings.Builder
	samiStyle.WriteString(fmt.Sprintf("\n    %s%s {\n    ", target, el))
	for k, v := range w.recreateStyle(s) {
		samiStyle.WriteString(fmt.Sprintf(" %s: %s;\n    ", k, v))
	}
	samiStyle.WriteString("}\n")
	return samiStyle.String()
}

func (w *Writer) recreateStyle(s caps.StyleProps) map[string]string {
	styles := map[string]string{}
	if s.TextAlign != "" {
		styles["text-align"] = s.TextAlign
	}
	if s.Italics {
		styles["font-style"] = "italic"
	}
	if s.FontFamily != "" {
		styles["font-family"] = s.FontFamily
	}
	if s.FontSize != "" {
		styles["font-size"] = s.FontSize
	}
	if s.Color != "" {
		styles["color"] = s.Color
	}
	if s.Lang != "" {
		styles["lang"] = s.Lang
	}
	return styles
}
