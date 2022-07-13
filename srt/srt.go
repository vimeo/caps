package srt

import (
	"regexp"

	"github.com/vimeo/caps"
)

const (
	base10            int    = 10
	bitSize64         int    = 64
	microHr           int64  = 3600000000
	microMin          int64  = 60000000
	microSec          int64  = 1000000
	microMilli        int64  = 1000
	timecodeSeparator string = "-->"
)

var (
	re        = regexp.MustCompile("^[0-9]{1,}$")
	reTiming  = regexp.MustCompile("^([0-9]{1,}:[0-9]{1,}:[0-9]{1,},[0-9]{1,}) --> ([0-9]{1,}:[0-9]{1,}:[0-9]{1,},[0-9]{1,})")
	reFont    = regexp.MustCompile("(?i)<font color=\"[0-9a-zA-Z]*\">")
	reEndFont = regexp.MustCompile("(?i)</font>")
)

func NewReader() caps.CaptionReader {
	return Reader{}
}

func NewWriter() caps.CaptionWriter {
	return Writer{}
}
