package webvtt

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/vimeo/caps"
)

const (
	bitSize64  int     = 64
	secHr      float64 = 3600
	secMin     float64 = 60
	microSec   float64 = 1000000
	microMilli float64 = 1000
)

var (
	timingPattern    = regexp.MustCompile("^(.+?) --> (.+)")
	timestampPattern = regexp.MustCompile(`^(\d+):(\d{2}):?(\d{2})?\.(\d{3})`)
	voiceSpanPattern = regexp.MustCompile(`<v(\\.\\w+)* ([^>]*)>`)
	otherSpanPattern = regexp.MustCompile("</?([cibuv]|ruby|rt|lang).*?>")
	webvttTiming     = "-->"
)

type reader struct {
	ignoreTimingErrors bool
}

func (r *reader) Read(content []byte) (*caps.CaptionSet, error) {
	captionSet := caps.NewCaptionSet()
	captions, err := r.parse(caps.SplitLines(string(content)))
	if err != nil {
		return nil, fmt.Errorf("error parsing webvtt: %w", err)
	}
	captionSet.SetCaptions(caps.DefaultLang, captions)
	if captionSet.IsEmpty() {
		return nil, fmt.Errorf("empty caption file")
	}
	return captionSet, nil
}

func (r *reader) parse(lines []string) ([]*caps.Caption, error) {
	captions := []*caps.Caption{}
	foundTiming := false
	var caption *caps.Caption
	var err error
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, webvttTiming) {
			foundTiming = true
			timingLine := i
			lastStartTime := 0.0
			if len(captions) != 0 {
				lastStartTime = *captions[len(captions)-1].Start
			}
			caption, err = parseTimingLine(line, lastStartTime, r.ignoreTimingErrors)
			if err != nil {
				return nil, fmt.Errorf("error parsing timing line %d: %w", timingLine, err)
			}
		} else if line == "" {
			if foundTiming {
				foundTiming = false
				if caption != nil && !caption.IsEmpty() {
					captions = append(captions, caption)
				}
				caption = nil
			}
		} else {
			if foundTiming {
				if caption != nil && !caption.IsEmpty() {
					caption.Nodes = append(caption.Nodes, caps.NewLineBreak())
				}
				caption.Nodes = append(caption.Nodes, caps.NewCaptionText(removeStyles(line)))
			}
		}
	}
	if caption != nil && !caption.IsEmpty() {
		captions = append(captions, caption)
	}
	return captions, nil
}

func (r *reader) Detect(content []byte) bool {
	return strings.Contains(string(content), "WEBVTT")
}

// Reader helpers
func microseconds(h, m, s, f string) (float64, error) {
	hh, err := strconv.ParseFloat(h, bitSize64)
	if err != nil {
		return 0, err
	}
	mm, err := strconv.ParseFloat(m, bitSize64)
	if err != nil {
		return 0, err
	}
	ss, err := strconv.ParseFloat(s, bitSize64)
	if err != nil {
		return 0, err
	}
	ff, err := strconv.ParseFloat(f, bitSize64)
	if err != nil {
		return 0, err
	}
	return (hh*secHr+mm*secMin+ss)*microSec + ff*microMilli, nil
}

func parseTimingLine(line string, lastStartTime float64, ignoreTimingErrors bool) (*caps.Caption, error) {
	matches := timingPattern.FindStringSubmatch(line)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid timing format")
	}
	start, err := parseTimestamp(matches[1])
	if err != nil {
		return nil, err
	}
	end, err := parseTimestamp(matches[2])
	if err != nil {
		return nil, err
	}
	caption := &caps.Caption{Start: start, End: end}
	if !ignoreTimingErrors {
		err := validateTimings(caption, lastStartTime)
		if err != nil {
			return nil, err
		}
	}
	return caption, nil
}

func parseTimestamp(input string) (*float64, error) {
	matches := timestampPattern.FindStringSubmatch(input)
	if len(matches) < 5 {
		return nil, nil
	}
	if matches[3] != "" {
		tmstp, err := microseconds(matches[1], matches[2], strings.ReplaceAll(matches[3], ":", ""), matches[4])
		return &tmstp, err
	}
	tmstp, err := microseconds("0", matches[1], matches[2], matches[4])
	return &tmstp, err
}

func removeStyles(line string) string {
	partialResult := voiceSpanPattern.ReplaceAllString(line, "\\2: ")
	return otherSpanPattern.ReplaceAllString(partialResult, "")
}

func validateTimings(caption *caps.Caption, lastStartTime float64) error {
	// FIXME: we might need to use a *float64 for Start/End so we can check for unset values.
	if caption.Start == nil {
		return fmt.Errorf("invalid cue start timestamp")
	}
	if caption.End == nil {
		return fmt.Errorf("invalid cue end timestamp")
	}
	if *caption.Start > *caption.End {
		return fmt.Errorf("end timestamp is not greater than start timestamp")
	}

	if *caption.Start < lastStartTime {
		return fmt.Errorf("start timestamp is not greater to start timestamp of previous cue")
	}
	return nil
}
