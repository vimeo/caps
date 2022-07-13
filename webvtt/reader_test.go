package webvtt

import (
	"testing"

	"github.com/vimeo/caps"
)

const sampleWebVTT = `WEBVTT
00:09.209 --> 00:12.312
( clock ticking )
00:14.848 --> 00:17.000
MAN:
When we think
♪ ...say bow, wow, ♪
00:17.000 --> 00:18.752
we have this vision of Einstein
00:18.752 --> 00:20.887
as an old, wrinkly man
with white hair.
00:20.887 --> 00:26.760
MAN 2:
E equals m c-squared is
not about an old Einstein.
00:26.760 --> 00:32.200
MAN 2:
It's all about an eternal Einstein.
00:32.200 --> 00:36.200
<LAUGHING & WHOOPS!>
`

const sampleWebVTT2 = `WEBVTT
1
00:00:00.000 --> 00:00:43.000
- HELLO WORLD!
2
00:00:59.000 --> 00:01:30.000
- LOOKING GOOOOD.
3
00:01:40.000 --> 00:02:00.000
- HA HA HA!
4
00:02:05.105 --> 00:03:07.007
- HI. WELCOME TO SESAME STREET.
5
00:04:07.007 --> 00:05:38.441
ON TONIGHT'S SHOW...
6
00:05:58.441 --> 00:06:40.543
- I'M NOT GOING TO WATCH THIS.
7
00:07:10.543 --> 00:07:51.711
HEY. WATCH THIS.
`

func TestWebVTTDetect(t *testing.T) {
	reader := NewReader(false)
	if !reader.Detect([]byte(sampleWebVTT)) {
		t.Error("should be detected as webvtt")
	}
}

func TestCaptionLength(t *testing.T) {
	captions, err := NewReader(false).Read([]byte(sampleWebVTT2))
	if err != nil {
		t.Errorf("error reading captions: %v", err)
	}
	if len(captions.GetCaptions(caps.DefaultLang)) != 7 {
		t.Error("there should be 7 captions for default lang")
	}
}
