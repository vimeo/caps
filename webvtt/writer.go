package webvtt

import (
	"bytes"
	"fmt"

	"github.com/vimeo/caps"
)

type Writer struct{}

func (w *Writer) Write(captionSet *caps.CaptionSet) ([]byte, error) {
	output := bytes.NewBufferString("WEBVTT\n\n")
	if captionSet.IsEmpty() || len(captionSet.Languages()) <= 0 {
		return output.Bytes(), nil
	}

	lang := captionSet.Languages()[0]
	captions := captionSet.GetCaptions(lang)

	for i, caption := range captions {
		output.WriteString(writeCaption(*caption))
		if i != len(captions)-1 {
			output.WriteString("\n")
		}

	}

	return output.Bytes(), nil
}

func writeCaption(caption caps.Caption) string {
	start := caption.FormatStart()
	end := caption.FormatEnd()

	timing := fmt.Sprintf("%s --> %s\n", start, end)
	output := bytes.NewBufferString(timing)

	if len(caption.Nodes) == 0 {
		output.WriteString("&nbsp;")
	} else {
		output.WriteString(formatNodes(caption.Nodes))
	}

	output.WriteString("\n")

	return output.String()
}

func formatNodes(nodes []caps.CaptionContent) string {
	content := ""

	for i, node := range nodes {
		if node.Text() {
			if node.Content() != "" {
				content += node.Content()
			} else {
				content += "&nbsp;"
			}
		} else if node.LineBreak() {
			isFirstNode := i == 0
			isTrailingBreak := i > 0 && nodes[i-1].LineBreak()

			if isFirstNode || isTrailingBreak {
				content += "&nbsp;"
			}

			content += "\n"
		}
	}

	return content
}
