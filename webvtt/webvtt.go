package webvtt

import "github.com/vimeo/caps"

func NewReader(ignoreTimingErrors bool) caps.CaptionReader {
	return &Reader{
		ignoreTimingErrors,
	}
}

func NewWriter() caps.CaptionWriter {
	return &Writer{}
}
