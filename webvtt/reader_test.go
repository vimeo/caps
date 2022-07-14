package webvtt

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestMicroseconds(t *testing.T) {
	var tests = []struct {
		name     string
		input    []string
		expected float64
		err      string
	}{
		{"parse hh:mm:ss.ttt - milliseconds", []string{"00", "00", "00", "999"}, 999000, ""},
		{"parse hh:mm:ss.ttt - seconds", []string{"00", "00", "59", "000"}, 59000000, ""},
		{"parse hh:mm:ss.ttt - mins", []string{"00", "59", "00", "000"}, 3540000000, ""},
		{"parse hh:mm:ss.ttt - hour", []string{"23", "00", "00", "000"}, 82800000000, ""},
		{"parse hh:mm:ss.ttt", []string{"23", "59", "59", "999"}, 86399999000, ""},
		{"parse invalid milliseconds", []string{"23", "59", "59", "9z9"}, 0, "strconv.ParseFloat: parsing \"9z9\": invalid syntax"},
		{"parse invalid seconds", []string{"23", "59", "5z", "999"}, 0, "strconv.ParseFloat: parsing \"5z\": invalid syntax"},
		{"parse invalid minutes", []string{"23", "5z", "59", "999"}, 0, "strconv.ParseFloat: parsing \"5z\": invalid syntax"},
		{"parse invalid hours", []string{"2z", "59", "59", "999"}, 0, "strconv.ParseFloat: parsing \"2z\": invalid syntax"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := microseconds(tt.input[0], tt.input[1], tt.input[2], tt.input[3])
			if tt.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, actual)
			} else {
				assert.EqualError(t, err, tt.err)
			}

		})
	}
}

func TestParseTimingLine(t *testing.T) {
	var tests = []struct {
		name               string
		line               string
		lastTime           float64
		ignoreTimingErrors bool
		expected           caps.Caption
		err                string
	}{
		{"parse valid cue - without validation", "00:00:50.000 --> 00:00:51.999", 49000000, true, caps.Caption{Start: floatPtr(50000000), End: floatPtr(51999000)}, ""},
		{"parse valid cue - with validation", "00:00:50.000 --> 00:00:51.999", 49000000, false, caps.Caption{Start: floatPtr(50000000), End: floatPtr(51999000)}, ""},
		{"parse invalid cue - without validation", "00:00:50.000 --> 00:00:51.999", 59000000, true, caps.Caption{Start: floatPtr(50000000), End: floatPtr(51999000)}, ""},
		{"parse invalid cue - with validation", "00:00:50.000 --> 00:00:51.999", 59000000, false, caps.Caption{}, "start timestamp is not greater to start timestamp of previous cue"},
		{"parse invalid cue - with validation", "00:00:52.000 --> 00:00:51.999", 49000000, false, caps.Caption{}, "end timestamp is not greater than start timestamp"},
		{"parse invalid timing line - missing end cue", "00:00:50.000 -->", 49000000, false, caps.Caption{}, "invalid timing format"},
		{"parse invalid timing line - missing start cue", "--> 00:00:51.999", 49000000, false, caps.Caption{}, "invalid timing format"},
		{"parse invalid timing line - invalid end cue", "00:00:50.000 --> 00:00:5x.999", 49000000, false, caps.Caption{}, "invalid cue end timestamp"},
		{"parse invalid timing line - invalid start cue", "00:00:5x.000 --> 00:00:51.999", 49000000, false, caps.Caption{}, "invalid cue start timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseTimingLine(tt.line, tt.lastTime, tt.ignoreTimingErrors)
			if tt.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, *actual)
			} else {
				assert.EqualError(t, err, tt.err)
			}

		})
	}
}

func TestParseTimestamp(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected float64
		err      string
	}{
		{"parse hh:mm:ss.ttt - milliseconds", "00:00:00.999", 999000, ""},
		{"parse hh:mm:ss.ttt - seconds", "00:00:59.000", 59000000, ""},
		{"parse hh:mm:ss.ttt - mins", "00:59:00.000", 3540000000, ""},
		{"parse hh:mm:ss.ttt - hour", "23:00:00.000", 82800000000, ""},
		{"parse hh:mm:ss.ttt", "23:59:59.999", 86399999000, ""},
		{"parse mm:ss.ttt - milliseconds", "00:00.999", 999000, ""},
		{"parse mm:ss.ttt - seconds", "00:59.000", 59000000, ""},
		{"parse mm:ss.ttt - mins", "59:00.000", 3540000000, ""},
		{"parse mm:ss.ttt", "59:59.999", 3599999000, ""},
		{"parse invalid timestamp", "00:23:59:59.999", 0, "nil"},
		{"parse invalid timestamp", "59.999", 0, "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := parseTimestamp(tt.input)
			if tt.err == "" {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, *actual)
			} else {
				assert.Equal(t, nil, err)
			}

		})
	}
}

func TestRemoveStyles(t *testing.T) {
	var tests = []struct {
		name     string
		input    string
		expected string
		err      string
	}{
		{"remove styles - voice span", "<v Bob>text</v>", "\\2: text", ""},
		{"remove styles - italic span", "<i>Yellow!</i>", "Yellow!", ""},
		{"remove styles - bold span", "<b>Yellow!</b>", "Yellow!", ""},
		{"remove styles - underline span", "<u>Yellow!</u>", "Yellow!", ""},
		{"remove styles - language span", "<lang en>Yellow!</lang>", "Yellow!", ""},
		{"remove styles - ruby span", "<ruby.loud>Yellow! <rt.loud>Yellow!</rt></ruby>", "Yellow! Yellow!", ""},
		{"remove styles - color span", "<c.yellow.bg_blue.magenta.bg_black>Yellow!</c>", "Yellow!", ""},
		{"remove styles - nested style spans", "Sur les <i.foreignphrase><lang en>playground</lang></i>, ici à Montpellier", "Sur les playground, ici à Montpellier", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := removeStyles(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// test helpers
func floatPtr(x float64) *float64 {
	return &x
}
