package webvtt

import "github.com/vimeo/caps"

func NewReader(ignoreTimingErrors bool) caps.CaptionReader {
	return &reader{
		ignoreTimingErrors,
	}
}
