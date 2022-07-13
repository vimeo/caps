package srt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindTextLine(t *testing.T) {
	var sampleLines = []string{"00:00:10,200 --> 00:00:12,300", "test line", "another test line", "", "00:00:11,300 --> 00:00:15,200", "test line 2", ""}

	tests := []struct {
		name           string
		inputStartLine int
		expected       int
	}{
		{"find text line - return endLine of first cue", 0, 4},
		{"find text line - return endLine of second cue", 4, 8},
		{"find text line - endLine > # of lines", 8, 9},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := findTextLine(test.inputStartLine, sampleLines)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSRTtoMicro(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		err      string
	}{
		{"parse timestamp - happy path", "23:59:59,999", 86399999000, ""},
		{"parse timestamp - milliseconds missing", "23:00:00", 82800000000, ""},
		{"parse invalid timestamp - hours missing", "59:00,000", 0, "invalid srt timestamp"},
		{"parse invalid milliseconds", "00:00:00,9z9", 0, "strconv.ParseInt: parsing \"9z9\": invalid syntax"},
		{"parse invalid seconds", "00:00:5z,000", 0, "strconv.ParseInt: parsing \"5z\": invalid syntax"},
		{"parse invalid minutes", "00:5z:00,000", 0, "strconv.ParseInt: parsing \"5z\": invalid syntax"},
		{"parse invalid hours", "2z:00:00,000", 0, "strconv.ParseInt: parsing \"2z\": invalid syntax"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualMicro, err := srtToMicro(test.input)
			if test.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, actualMicro)
			} else {
				assert.EqualError(t, err, test.err)
			}
		})
	}
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"is digit - single digit", "0", true},
		{"is digit - multiple digits", "123456789", true},
		{"is digit - digit and letters", "1xx", false},
		{"is digit - letters", "test", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := isDigit(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"split line - no breaks", `00:00:10,200 --> 00:00:12,300`, []string{`00:00:10,200 --> 00:00:12,300`}},
		{"split line - one line break", "00:00:10,200 --> 00:00:12,300\r\ntest line", []string{"00:00:10,200 --> 00:00:12,300", "test line"}},
		{"split line - multiple breaks", "00:00:10,200 --> 00:00:12,300\n\r\ntest line", []string{"00:00:10,200 --> 00:00:12,300", "", "test line"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := splitLines(test.input)
			assert.Equal(t, test.expected, actual)
		})
	}
}
