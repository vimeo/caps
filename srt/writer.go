package srt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/thiagopnts/caps"
)

type Writer struct{}

func (w Writer) Write(captionSet *caps.CaptionSet) ([]byte, error) {
	content, err := w.WriteString(captionSet)
	return []byte(content), err
}

func (Writer) WriteString(captionSet *caps.CaptionSet) (string, error) {
	contents := []string{}
	for _, lang := range captionSet.Languages() {
		contents = append(contents, recreateLang(captionSet.GetCaptions(lang)))
	}
	return strings.Join(contents, "MULTI-LANGUAGE SRT\n"), nil
}

func recreateLang(captions []*caps.Caption) string {
	content := ""
	count := 1
	for _, caption := range captions {
		content += fmt.Sprintf("%s\n", strconv.Itoa(count))

		start := caption.FormatStartWithSeparator(",")
		end := caption.FormatEndWithSeparator(",")
		content += fmt.Sprintf("%s %s %s\n", start[:12], timecodeSeparator, end[:12])

		lines := ""
		for _, node := range caption.Nodes {
			lines += recreateLine(node)
		}
		content += fmt.Sprintf("%s\n\n", strings.ReplaceAll(lines, "\n\n", "\n"))
		count++
	}
	return content[:len(content)-1]
}

func recreateLine(node caps.CaptionContent) string {
	if node.Text() {
		return node.Content()
	}
	if node.LineBreak() {
		return "\n"
	}
	return ""
}
