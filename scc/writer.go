package scc

import (
	"bytes"
	"fmt"
	"math"
	"strings"

	"github.com/vimeo/caps"
)

type Writer struct{}

type codeMetadata struct {
	Code  string
	Start *float64
	End   *float64
}

func (w *Writer) Write(captionSet *caps.CaptionSet) ([]byte, error) {
	output := bytes.NewBufferString(header)
	output.WriteString("\n\n")
	if captionSet.IsEmpty() || len(captionSet.Languages()) <= 0 {
		return output.Bytes(), nil
	}
	// support only one language
	lang := captionSet.Languages()[0]
	captions := captionSet.GetCaptions(lang)
	codes := []codeMetadata{}
	for _, caption := range captions {
		code := w.textToCode(caption)
		codes = append(codes, codeMetadata{code, &caption.Start, &caption.End})
	}

	for index, metadata := range codes {
		codeWords := len(metadata.Code) / 13
		codeTimeMicroseconds := float64(codeWords) * microsecondsPerCodeword
		codeStart := *metadata.Start - codeTimeMicroseconds
		if index == 0 {
			continue
		}
		prevMetadata := codes[index-1]
		if prevMetadata.End != nil && *prevMetadata.End+3*microsecondsPerCodeword >= codeStart {
			codes[index-1] = codeMetadata{prevMetadata.Code, prevMetadata.Start, nil}
		}
		codeStartCast := codeStart
		codes[index] = codeMetadata{metadata.Code, &codeStartCast, metadata.End}
	}

	for _, metadata := range codes {
		if metadata.Start == nil {
			continue
		}
		output.WriteString(fmt.Sprintf("%s\t", w.formatTimestamp(*metadata.Start)))
		output.WriteString("94ae 94ae 9420 9420 ")
		output.WriteString(metadata.Code)
		output.WriteString("942c 942c 942f 942f\n\n")
		if metadata.End != nil {
			output.WriteString(fmt.Sprintf("%s\t942c 942c\n\n", w.formatTimestamp(*metadata.End)))
		}
	}
	return output.Bytes(), nil
}

func (w *Writer) textToCode(caption *caps.Caption) string {
	code := bytes.NewBufferString("")
	lines := []string{}
	for _, line := range strings.Split(w.layoutLine(caption), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	for row, line := range lines {
		index := 16 - len(lines) + row
		for i := 0; i < 2; i++ {
			value := fmt.Sprintf("%s%s ", pacHighByteByRow[index], pacLowByteByRow[index])
			code.WriteString(value)
		}

		for _, char := range line {
			w.printCharacter(code, string(char))
			w.maybeSpace(code)
		}
		w.maybeAlign(code)
	}
	return code.String()
}

func (w *Writer) printCharacter(buf *bytes.Buffer, char string) {
	var charCode string
	if code, ok := charactersToCode[char]; ok {
		charCode = code
	} else if code, ok = specialExtendedToCode[char]; ok {
		charCode = code
	} else {
		// use £ as "unknown character" symbol
		charCode = "91b6"
	}
	if len(charCode) == 2 {
		buf.WriteString(charCode)
	} else if len(charCode) == 4 {
		w.maybeAlign(buf)
		buf.WriteString(charCode)
	}
}
func (w *Writer) formatTimestamp(microseconds float64) string {
	secondsFloat := float64(microseconds) / 1000.0 / 1000.0
	// convert to non-drop-frame timecode
	secondsFloat *= 1000.0 / 1001.0
	hours := math.Floor(secondsFloat / 3600)
	secondsFloat -= hours * 3600
	minutes := math.Floor(secondsFloat / 60)
	secondsFloat -= minutes * 60

	seconds := math.Floor(secondsFloat)
	secondsFloat -= seconds
	frames := math.Floor(secondsFloat * 30)
	return fmt.Sprintf("%02d:%02d:%02d:%02d", int(hours), int(minutes), int(seconds), int(frames))
}

func (w *Writer) maybeSpace(buf *bytes.Buffer) {
	if len(buf.String())%5 == 4 {
		buf.WriteString(" ")
	}
}

func (w *Writer) maybeAlign(buf *bytes.Buffer) {
	// finish a half-word with a no-op so we can move to a full word
	if buf.Len()%5 == 2 {
		buf.WriteString("80 ")
	}
}

func (w *Writer) layoutLine(caption *caps.Caption) string {
	capText := bytes.NewBufferString("")
	for _, node := range caption.Nodes {
		if node.Text() || node.LineBreak() {
			capText.WriteString(node.Content())
		}
	}
	innerLines := strings.Split(capText.String(), "\n")
	innerLinesLaidOut := bytes.NewBufferString("")
	// [str[i * 32:(i + 1)*32] for i in range(math.ceil(len(str)/32))]
	for _, line := range innerLines {
		for i := 0; i < int(math.Ceil(float64(len(line))/32)); i++ {
			content := line
			if len(line) > 32 {
				content = line[i*32 : (i+1)*32]
			}
			innerLinesLaidOut.WriteString(content)
			innerLinesLaidOut.WriteString("\n")
		}
	}
	// pop out the last \n
	innerLinesLaidOut.UnreadByte()
	innerLinesLaidOut.UnreadByte()
	return innerLinesLaidOut.String()
}
