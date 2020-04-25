package caps

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleDFXP string = `
<?xml version="1.0" encoding="utf-8"?>
<tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml"
    xmlns:tts="http://www.w3.org/ns/ttml#styling">
 <head>
  <styling>
   <style xml:id="p" tts:color="#ffeedd" tts:fontfamily="Arial"
          tts:fontsize="10pt" tts:textAlign="center"/>
  </styling>
  <layout>
  <region tts:displayAlign="after" tts:textAlign="center" xml:id="bottom"></region>
  </layout>
 </head>
 <body>
  <div xml:lang="en-US">
   <p begin="00:00:09.209" end="00:00:12.312" style="p">
    ( clock ticking )
   </p>
   <p begin="00:00:14.848" end="00:00:17.000" style="p">
    MAN:<br/>
    When we think<br/>
    ♪ ...say bow, wow, ♪
   </p>
   <p begin="00:00:17.000" end="00:00:18.752" style="p">
    <span tts:textalign="right">we have this vision of Einstein</span>
   </p>
   <p begin="00:00:18.752" end="00:00:20.887" style="p">
   <br/>
    as an old, wrinkly man<br/>
    with white hair.
   </p>
   <p begin="00:00:20.887" end="00:00:26.760" style="p">
    MAN 2:<br/>
    E equals m c-squared is<br/>
    not about an old Einstein.
   </p>
   <p begin="00:00:26.760" end="00:00:32.200" style="p">
    MAN 2:<br/>
    It's all about an eternal Einstein.
   </p>
   <p begin="00:00:32.200" end="00:00:36.200" style="p">
    &lt;LAUGHING &amp; WHOOPS!&gt;
   </p>
   <p begin="00:00:34.400" end="00:00:38.400" region="bottom" style="p">some more text</p>
  </div>
 </body>
</tt> `

const sampleDFXPSyntaxError = `
  <?xml version="1.0" encoding="UTF-8"?>
  <tt xml:lang="en" xmlns="http://www.w3.org/ns/ttml">
  <body>
    <div>
      <p begin="0:00:02.07" end="0:00:05.07">>>THE GENERAL ASSEMBLY'S 2014</p>
      <p begin="0:00:05.07" end="0:00:06.21">SESSION GOT OFF TO A LATE START,</p>
    </div>
   </body>
  </tt>
`

const sampleDFXPEmpty = `
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
`

func TestDection(t *testing.T) {
	assert.True(t, NewDFXPReader().Detect(sampleDFXP))
}

func TestCaptionLength(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXP)
	assert.Nil(t, err)
	assert.Equal(t, 8, len(captionSet.GetCaptions("en-US")))
}

func TestEmptyFile(t *testing.T) {
	set, err := NewDFXPReader().Read(sampleDFXPEmpty)
	assert.NotNil(t, err)
	assert.True(t, set.IsEmpty())
}

func TestProperTimestamps(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXP)
	assert.Nil(t, err)

	paragraph := captionSet.GetCaptions("en-US")[2]
	assert.Equal(t, 17000000, paragraph.Start)
	assert.Equal(t, 18752000, paragraph.End)
}

func TestInvalidMarkupIsProperlyHandled(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXPSyntaxError)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(captionSet.GetCaptions("en-US")))
}

func TestStructToXML(t *testing.T) {
	base := DFXPBaseMarkup{
		TtXMLLang:  "en",
		TtXMLns:    "http://www.w3.org/ns/ttml",
		TtXMLnsTTS: "http://www.w3.org/ns/ttml#styling",
	}

	output, err := xml.MarshalIndent(base, "  ", "    ")
	fmt.Println(err)
	fmt.Println(string(output))
}

func TestDFXPWriter(t *testing.T) {
	captionSet, err := NewDFXPReader().Read(sampleDFXP)
	assert.Nil(t, err)
	data, _ := NewDFXPWriter().Write(captionSet)
	output, err := xml.MarshalIndent(data, "  ", "    ")
	fmt.Println(err)
	fmt.Println(string(output))
	fmt.Println("--------------------------")
	fmt.Println(sampleDFXP)
}
