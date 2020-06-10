package srt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/thiagopnts/caps"
)

type Writer struct{}

func (Writer) Write(captionSet *caps.CaptionSet) string {
	contents := []string{}
	for _, lang := range captionSet.GetLanguages() {
		contents = append(contents, recreateLang(captionSet.GetCaptions(lang)))
	}
	return strings.Join(contents, "MULTI-LANGUAGE SRT\n")
}

func recreateLang(captions []*caps.Caption) string {
	content := ""
	count := 1
	for _, caption := range captions {
		content += fmt.Sprintf("%s\n", strconv.Itoa(count))
		start := caption.FormatStartWithSeparator(",")
		end := caption.FormatEndWithSeparator(",")
		content += fmt.Sprintf("%s --> %s\n", start[:12], end[:12])
		newContent := ""
		for _, node := range caption.Nodes {
			newContent += recreateLine(node)
		}
		content += fmt.Sprintf("%s\n\n", strings.ReplaceAll(newContent, "\n\n", "\n"))
		count++
	}
	return content[:len(content)-1]
}

func recreateLine(node caps.CaptionContent) string {
	if node.IsText() {
		return node.GetContent()
	}
	if node.IsLineBreak() {
		return "\n"
	}
	return ""
}
