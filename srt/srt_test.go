package srt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vimeo/caps"
)

func TestSRTDetection(t *testing.T) {
	assert.True(t, Reader{}.Detect(SampleSRT))
	assert.False(t, Reader{}.Detect(InvalidSRT1))
	assert.False(t, Reader{}.Detect(InvalidSRT2))
}

func TestSRTCaptionLength(t *testing.T) {
	reader := NewReader()
	captions, err := reader.Read(SampleSRT)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(captions.GetCaptions(caps.DefaultLang)))
}

func TestSRTTimestamp(t *testing.T) {
	reader := NewReader()
	captions, err := reader.Read(SampleSRT)
	assert.Nil(t, err)
	p := captions.GetCaptions(caps.DefaultLang)[2]
	assert.Equal(t, 17000000, int(*p.Start))
	assert.Equal(t, 18752000, int(*p.End))
}

func TestSRTStripFontColor(t *testing.T) {
	reader := NewReader()
	captions, err := reader.Read(SampleSRTFontColor)
	text := captions.GetCaptions(caps.DefaultLang)[3].Nodes[0].Content()
	assert.Nil(t, err)
	assert.Equal(t, 4, len(captions.GetCaptions(caps.DefaultLang)))
	assert.Equal(t, "as an old, wrinkly man", text)
}

func TestSRTNumeric(t *testing.T) {
	reader := NewReader()
	captions, err := reader.Read(SampleSRTNumeric)
	assert.Nil(t, err)
	assert.Equal(t, 7, len(captions.GetCaptions(caps.DefaultLang)))
}

func TestSRTEmptyFile(t *testing.T) {
	reader := NewReader()
	_, err := reader.Read(SampleSRTEmpty)
	assert.NotNil(t, err)
}

func TestSRTExtraEmpty(t *testing.T) {
	reader := NewReader()
	captions, err := reader.Read(SampleSRTBlankLines)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(captions.GetCaptions(caps.DefaultLang)))
}

func TestSRTtoSRT(t *testing.T) {
	type srtToSRTTests struct {
		inputSRT []byte
		wantSRT  []byte
	}
	srtConvertionTests := []srtToSRTTests{
		{inputSRT: SampleSRT, wantSRT: SampleSRT},
		{inputSRT: SampleSRTutf8, wantSRT: SampleSRTutf8},
		{inputSRT: SampleSRTUnicode, wantSRT: SampleSRTUnicode},
	}
	for _, test := range srtConvertionTests {
		captions, err := NewReader().Read(test.inputSRT)
		assert.Nil(t, err)
		result, _ := NewWriter().Write(captions)
		assert.Equal(t, test.wantSRT, result)
	}
}

var InvalidSRT1 = []byte(`1
blank line
`)

var InvalidSRT2 = []byte(`1
00:00:09,209 -> 00:00:12,312
incorrect separator
`)

var SampleSRTUnicode = []byte(`1
00:00:09,209 --> 00:00:12,312
( clock ticking )

2
00:00:14,848 --> 00:00:17,000
MAN:
When we think
\u266a ...say bow, wow, \u266a

3
00:00:17,000 --> 00:00:18,752
we have this vision of Einstein

4
00:00:18,752 --> 00:00:20,887
as an old, wrinkly man
with white hair.

5
00:00:20,887 --> 00:00:26,760
MAN 2:
E equals m c-squared is
not about an old Einstein.

6
00:00:26,760 --> 00:00:32,200
MAN 2:
It's all about an eternal Einstein.

7
00:00:32,200 --> 00:00:36,200
<LAUGHING & WHOOPS!>
`)

var SampleSRTutf8 = []byte(`1
00:00:09,209 --> 00:00:12,312
( clock ticking )

2
00:00:14,848 --> 00:00:17,000
MAN:
When we think
♪ ...say bow, wow, ♪

3
00:00:17,000 --> 00:00:18,752
we have this vision of Einstein

4
00:00:18,752 --> 00:00:20,887
as an old, wrinkly man
with white hair.

5
00:00:20,887 --> 00:00:26,760
MAN 2:
E equals m c-squared is
not about an old Einstein.

6
00:00:26,760 --> 00:00:32,200
MAN 2:
It's all about an eternal Einstein.

7
00:00:32,200 --> 00:00:36,200
<LAUGHING & WHOOPS!>
`)

var SampleSRT = []byte(`1
00:00:09,209 --> 00:00:12,312
( clock ticking )

2
00:00:14,848 --> 00:00:17,000
MAN:
When we think
of "E equals m c-squared",

3
00:00:17,000 --> 00:00:18,752
we have this vision of Einstein

4
00:00:18,752 --> 00:00:20,887
as an old, wrinkly man
with white hair.

5
00:00:20,887 --> 00:00:26,760
MAN 2:
E equals m c-squared is
not about an old Einstein.

6
00:00:26,760 --> 00:00:32,200
MAN 2:
It's all about an eternal Einstein.

7
00:00:32,200 --> 00:00:34,400
<LAUGHING & WHOOPS!>

8
00:00:34,400 --> 00:00:38,400
some more text
`)

var SampleSRTFontColor = []byte(`1
00:00:09,209 --> 00:00:12,312
( clock ticking )

2
00:00:14,848 --> 00:00:17,000
MAN:
When we think
of "E equals m c-squared",

3
00:00:17,000 --> 00:00:18,752
<LAUGHING & WHOOPS!>

4
00:00:18,752 --> 00:00:20,887
<font color="white">as an old, wrinkly man</font>
`)

var SampleSRTNumeric = []byte(`35
00:00:32,290 --> 00:00:32,890
TO  FIND  HIM.            IF

36
00:00:32,990 --> 00:00:33,590
YOU  HAVE  ANY  INFORMATION

37
00:00:33,690 --> 00:00:34,290
THAT  CAN  HELP,  CALL  THE

38
00:00:34,390 --> 00:00:35,020
STOPPERS  LINE.          THAT

39
00:00:35,120 --> 00:00:35,760
NUMBER  IS  662-429-84-77.

40
00:00:35,860 --> 00:00:36,360
STD  OUT

41
00:00:36,460 --> 00:02:11,500
3
`)

var SampleSRTEmpty = []byte(`
`)

var SampleSRTBlankLines []byte = []byte(`35
00:00:32,290 --> 00:00:32,890


36
00:00:32,990 --> 00:00:33,590
YOU  HAVE  ANY  INFORMATION

`)
