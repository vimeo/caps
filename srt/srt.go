package srt

import (
	"regexp"

	"github.com/thiagopnts/caps"
)

var re = regexp.MustCompile("[0-9]{1,}")
var reTiming = regexp.MustCompile("^([0-9]{1,}:[0-9]{1,}:[0-9]{1,},[0-9]{1,}) --> ([0-9]{1,}:[0-9]{1,}:[0-9]{1,},[0-9]{1,})")
var reFont = regexp.MustCompile("(?i)<font color=\"#[0-9a-f]{6}\">")
var reEndFont = regexp.MustCompile("(?i)</font>")

func NewReader() caps.CaptionReader {
	return Reader{}
}
