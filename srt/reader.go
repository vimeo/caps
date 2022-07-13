package srt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/vimeo/caps"
)

type Reader struct{}

func (Reader) Detect(content []byte) bool {
	lines := splitLines(string(content))
	if len(lines) < 2 {
		return false
	}
	return isDigit(lines[0]) && strings.Contains(lines[1], timecodeSeparator)
}
func (r Reader) Read(content []byte) (*caps.CaptionSet, error) {
	captionSet := caps.NewCaptionSet()
	captions := []*caps.Caption{}
	lines := splitLines(string(content))
	startLine := 0

	for startLine < len(lines) {
		if !isDigit(lines[startLine]) {
			break
		}
		var capStart int64
		var capEnd int64
		var err error
		endLine := findTextLine(startLine, lines)
		if matches := reTiming.FindAllString(lines[startLine+1], -1); len(matches) >= 3 {
			capStart, err = srtToMicro(matches[1])
			if err != nil {
				return nil, err
			}
			capEnd, err = srtToMicro(matches[2])
			if err != nil {
				return nil, err
			}
		} else {
			timing := strings.Split(lines[startLine+1], timecodeSeparator)
			if len(timing) < 2 {
				return nil, fmt.Errorf("malformed srt file")
			}
			capStart, err = srtToMicro(strings.Trim(timing[0], " \r\n"))
			if err != nil {
				return nil, err
			}
			capEnd, err = srtToMicro(strings.Trim(timing[1], " \r\n"))
			if err != nil {
				return nil, err
			}
		}
		capNodes := []caps.CaptionContent{}
		for _, line := range lines[startLine+2 : endLine-1] {
			cleanLine := reFont.ReplaceAllString(line, "")
			cleanLine = reEndFont.ReplaceAllString(cleanLine, "")
			if len(capNodes) == 0 || line != "" {
				capNodes = append(capNodes, caps.NewCaptionText(cleanLine))
				capNodes = append(capNodes, caps.NewLineBreak())
			}
		}
		if len(capNodes) > 0 {
			capNodes = capNodes[:len(capNodes)-1]
			c := caps.NewCaption(float64(capStart), float64(capEnd), capNodes, caps.DefaultStyleProps())
			captions = append(captions, &c)
		}
		startLine = endLine
	}
	captionSet.SetCaptions(caps.DefaultLang, captions)
	if captionSet.IsEmpty() {
		return nil, fmt.Errorf("empty srt file")
	}
	return captionSet, nil
}

func findTextLine(startLine int, lines []string) int {
	endLine := startLine
	found := false
	for endLine < len(lines) {
		if strings.TrimSpace(lines[endLine]) == "" {
			found = true
		} else if found {
			endLine--
			break
		}
		endLine++
	}
	return endLine + 1
}

func srtToMicro(stamp string) (int64, error) {
	timesplit := strings.Split(stamp, ":")
	if len(timesplit) < 3 {
		return 0, fmt.Errorf("invalid srt timestamp")
	}
	if !strings.Contains(timesplit[2], ",") {
		timesplit[2] = fmt.Sprintf("%s,000", timesplit[2])
	}
	timesplit0, err := strconv.ParseInt(timesplit[0], base10, bitSize64)
	if err != nil {
		return 0, err
	}
	timesplit1, err := strconv.ParseInt(timesplit[1], base10, bitSize64)
	if err != nil {
		return 0, err
	}
	secsplit := strings.Split(timesplit[2], ",")
	secsplit0, err := strconv.ParseInt(secsplit[0], base10, bitSize64)
	if err != nil {
		return 0, err
	}
	secsplit1, err := strconv.ParseInt(secsplit[1], base10, bitSize64)
	if err != nil {
		return 0, err
	}
	microseconds := timesplit0*microHr + timesplit1*microMin + secsplit0*microSec + secsplit1*microMilli
	return microseconds, nil
}

func isDigit(line string) bool {
	return re.Match([]byte(line))
}

func splitLines(content string) []string {
	return strings.Split(
		strings.ReplaceAll(
			strings.ReplaceAll(content, "\r\n", "\n"),
			"\r",
			"\n",
		),
		"\n",
	)
}
