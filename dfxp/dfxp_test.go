package dfxp

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thiagopnts/caps"
)

var sampleDFXP = []byte(`<?xml version="1.0" encoding="utf-8"?>
<tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml" xmlns:tts="http://www.w3.org/ns/ttml#styling">
  <head>
    <styling>
      <style xml:id="p" tts:fontfamily="Arial" tts:fontsize="10pt" tts:textAlign="center" tts:color="#ffeedd"></style>
    </styling>
    <layout>
      <region tts:displayAlign="after" tts:textAlign="center" xml:id="bottom"></region>
    </layout>
  </head>
  <body>
    <div xml:lang="en-US">
      <p begin="00:00:14.848" end="00:00:17.000" style="p">
        MAN:
        <br/>
        When we think
        <br/>
        ♪ ...say bow, wow, ♪
      </p>
     <p begin="00:00:17.000" end="00:00:18.752" style="p">
      <span tts:textalign="right">we have this vision of Einstein</span>
     </p>
     <p begin="00:00:18.752" end="00:00:20.887" style="p">
       <br/>
       as an old, wrinkly man
       <br/>
       with white hair.
     </p>
     <p begin="00:00:20.887" end="00:00:26.760" style="p">
      MAN 2:
      <br/>
      E equals m c-squared is
      <br/>
      not about an old Einstein.
     </p>
     <p begin="00:00:26.760" end="00:00:32.200" style="p">
      MAN 2:
      <br/>
      It's all about an eternal Einstein.  pois é
     </p>
     <p begin="00:00:32.200" end="00:00:36.200" style="p">
      &lt;LAUGHING &amp; WHOOPS!&gt;
     </p>
     <p begin="00:00:34.400" end="00:00:38.400" region="bottom" style="p">
      some more text
     </p>
    </div>
  </body>
</tt>`)

var sampleDFXPSyntaxError = []byte(`
  <?xml version="1.0" encoding="UTF-8"?>
  <tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml">
  <body>
    <div>
      <p begin="0:00:02.07" end="0:00:05.07">>>THE GENERAL ASSEMBLY'S 2014</p>
      <p begin="0:00:05.07" end="0:00:06.21">SESSION GOT OFF TO A LATE START,</p>
    </div>
   </body>
  </tt>
`)

var sampleDFXPEmpty = []byte(`
  <?xml version="1.0" encoding="utf-8"?>
  <tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml"
      xmlns:tts="http://www.w3.org/ns/ttml#styling">
   <head>
    <styling>
     <style xml:id="p" tts:color="#ffeedd" tts:fontfamily="Arial"
          tts:fontsize="10pt" tts:textAlign="center"/>
    </styling>
    <layout>
    </layout>
   </head>
   <body>
    <div xml:lang="en-US">
    </div>
   </body>
  </tt>
`)

func TestDection(t *testing.T) {
	assert.True(t, NewReader().Detect(sampleDFXP))
	assert.False(t, NewReader().Detect([]byte("invalid XML")))
}

func TestCaptionLength(t *testing.T) {
	captionSet, err := NewReader().Read(sampleDFXP)
	assert.Nil(t, err)
	assert.Equal(t, 7, len(captionSet.GetCaptions("en-US")))
}

func TestEmptyFile(t *testing.T) {
	set, err := NewReader().Read(sampleDFXPEmpty)
	assert.NotNil(t, err)
	assert.True(t, set.IsEmpty())
}

func TestProperTimestamps(t *testing.T) {
	captionSet, err := NewReader().Read(sampleDFXP)
	assert.Nil(t, err)

	paragraph := captionSet.GetCaptions("en-US")[2]
	assert.Equal(t, 18752000, int(paragraph.Start))
	assert.Equal(t, 20887000, int(paragraph.End))
}

func TestInvalidMarkupIsProperlyHandled(t *testing.T) {
	captionSet, err := NewReader().Read(sampleDFXPSyntaxError)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(captionSet.GetCaptions("en-US")))
}

func TestCaptionNodes(t *testing.T) {
	captionSet, err := NewReader().Read(sampleDFXP)
	assert.Nil(t, err)
	styles := captionSet.GetStyles()
	assert.Equal(t, 1, len(styles))
	assert.Equal(t, styles[0].ID, "p")
	assert.Equal(t, styles[0].FontFamily, "Arial")
	assert.Equal(t, styles[0].FontFamily, "Arial")
	assert.Equal(t, styles[0].FontSize, "10pt")
	assert.Equal(t, styles[0].TextAlign, "center")
	assert.Equal(t, styles[0].Color, "#ffeedd")
	type captionNodeTest struct {
		wantFormatStart string
		wantFormatEnd   string
		wantText        string
	}
	captionTests := []captionNodeTest{
		{wantFormatStart: "00:00:14.848", wantFormatEnd: "00:00:17.000", wantText: "MAN:\nWhen we think\n♪ ...say bow, wow, ♪"},
		{wantFormatStart: "00:00:17.000", wantFormatEnd: "00:00:18.752", wantText: "we have this vision of Einstein"},
		{wantFormatStart: "00:00:18.752", wantFormatEnd: "00:00:20.887", wantText: "\nas an old, wrinkly man\nwith white hair."},
		{wantFormatStart: "00:00:20.887", wantFormatEnd: "00:00:26.760", wantText: "MAN 2:\nE equals m c-squared is\nnot about an old Einstein."},
		{wantFormatStart: "00:00:26.760", wantFormatEnd: "00:00:32.200", wantText: "MAN 2:\nIt's all about an eternal Einstein.  pois é"},
		{wantFormatStart: "00:00:32.200", wantFormatEnd: "00:00:36.200", wantText: "<LAUGHING & WHOOPS!>"},
		{wantFormatStart: "00:00:34.400", wantFormatEnd: "00:00:38.400", wantText: "some more text"},
	}

	nodeTests := [][]caps.CaptionContent{
		[]caps.CaptionContent{
			caps.NewCaptionText("MAN:"),
			caps.NewLineBreak(),
			caps.NewCaptionText("When we think"),
			caps.NewLineBreak(),
			caps.NewCaptionText("♪ ...say bow, wow, ♪"),
		},
		[]caps.CaptionContent{
			caps.NewCaptionStyle(false, caps.StyleProps{
				TextAlign: "right",
			}),
			caps.NewCaptionText("we have this vision of Einstein"),
		},
		[]caps.CaptionContent{
			caps.NewLineBreak(),
			caps.NewCaptionText("as an old, wrinkly man"),
			caps.NewLineBreak(),
			caps.NewCaptionText("with white hair."),
		},
		[]caps.CaptionContent{
			caps.NewCaptionText("MAN 2:"),
			caps.NewLineBreak(),
			caps.NewCaptionText("E equals m c-squared is"),
			caps.NewLineBreak(),
			caps.NewCaptionText("not about an old Einstein."),
		},
		[]caps.CaptionContent{
			caps.NewCaptionText("MAN 2:"),
			caps.NewLineBreak(),
			caps.NewCaptionText("It's all about an eternal Einstein.  pois é"),
		},
		[]caps.CaptionContent{
			caps.NewCaptionText("<LAUGHING & WHOOPS!>"),
		},
		[]caps.CaptionContent{
			caps.NewCaptionText("some more text"),
		},
	}

	captions := captionSet.GetCaptions("en-US")
	for i, caption := range captions {
		assert.Equal(t, caption.FormatStart(), captionTests[i].wantFormatStart)
		assert.Equal(t, caption.FormatEnd(), captionTests[i].wantFormatEnd)
		assert.Equal(t, caption.Text(), captionTests[i].wantText)
		for j, node := range caption.Nodes {
			assert.Equal(t, node.Text(), nodeTests[i][j].Text())
			assert.Equal(t, node.Style(), nodeTests[i][j].Style())
			assert.Equal(t, node.LineBreak(), nodeTests[i][j].LineBreak())
			assert.Equal(t, node.Content(), nodeTests[i][j].Content())
		}
	}
}

func TestStructToXML(t *testing.T) {
	base := Span{xml.Name{}, "bla blablab bla alba", Style{TTSTextAlign: "center"}}
	output, err := xml.MarshalIndent(base, "  ", "    ")
	fmt.Println(err)
	fmt.Println(string(output))
}

//func TestDFXPWriter(t *testing.T) {
//	captionSet, err := NewReader().Read(sampleDFXP)
//	assert.Nil(t, err)
//	data, _ := NewWriter().Write(captionSet)
//	fmt.Println(sampleDFXP[0])
//	fmt.Println("--------------------------")
//	fmt.Println(gohtml.Format(string(output[0])))
//}
