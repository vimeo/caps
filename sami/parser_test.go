package sami

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"golang.org/x/net/html"
)

func render(n *html.Node) string {
	var buf bytes.Buffer
	w := io.Writer(&buf)
	html.Render(w, n)
	return buf.String()
}

func TestParser(t *testing.T) {
	tests := []struct {
		name        string
		input       io.Reader
		expected    string
		expectedErr bool
	}{
		{
			name:  "sami with multiple styles",
			input: strings.NewReader(sample1),
			expected: `<sami>
        <head>
          <title>test123</title>
          <style type="text/css">
            <!--
            P { margin-right:  1pt; }
            -->
          </style>
          <style type="text/css">
            <!--
            P { margin-left:  1pt; color: #ffeedd; }
            .ENCC {Name: English; lang: en-US; SAMI_Type: CC;}
            -->
          </style>
        </head>
        <body>
          <sync start="9209">
            <p class="ENCC">
              ( clock ticking )
            </p>
          </sync>
          <sync start="12312"><p class="ENCC"> </p></sync>
          <sync start="14848">
            <p class="ENCC">
              MAN:<br/>
              When we think<br/>
              of &#34;E equals m c-squared&#34;,
            </p>
          </sync>
        </body>
      </sami>`,
			expectedErr: false,
		},
		{
			name:  "sami with encoded chars",
			input: strings.NewReader(sample2),
			expected: `<sami>
        <head>
          <title>test123</title>
          <style type="text/css">
          <!--
          P { margin-left:  1pt; color: #ffeedd; }
          .ENCC {Name: English; lang: en-US; SAMI_Type: CC;}
          -->
          </style>
        </head>
        <body>
          <sync start="12312"><p class="ENCC"> </p></sync>
          <sync start="17000">
          <p class="ENCC">
          <span style="text-align:right;">we have this vision of Einstein</span>
          </p>
          </sync>
          <sync start="32200">
          <p class="ENCC">
          &lt;LAUGHING &amp; WHOOPS!&gt;
          </p>
          </sync>
        </body>
      </sami>`,
			expectedErr: false,
		},
		{
			name:  "sami with tts span style",
			input: strings.NewReader(ttsSample),
			expected: `
        <sami>
          <head>
          </head>
          <body>
            <sync start="6534661"><p class="ENCC">
            <span tts:fontstyle="italic">NOVA</span> is <span foo="bar">a</span> production<br/>
            of WGBH Boston.
            </p></sync>
          </body>
        </sami>`,
			expectedErr: false,
		},
		{
			name:  "sami with tts span style",
			input: strings.NewReader(ttsSample),
			expected: `
              <sami>
                <head>
                </head>
                <body>
                  <sync start="6534661"><p class="ENCC">
                  <span tts:fontstyle="italic">NOVA</span> is <span foo="bar">a</span> production<br/>
                  of WGBH Boston.
                  </p></sync>
                </body>
              </sami>`,
			expectedErr: false,
		},
		{
			name:  "follows whatever the tokenizer parses when handling invalid tag syntax",
			input: strings.NewReader(syntaxErrorSample),
			expected: `
        <sami>
        <head>
        <title>ir2014_111</title>
        <style type="text/css">
        <!--
        P { margin-left:  1pt;
        margin-right: 1pt;
        margin-bottom: 2pt;
        margin-top: 2pt;
        text-align: center;
        font-size: 10pt;
        font-family: Arial;
        font-weight: normal;
        font-style: normal;
        color: #ffffff; }

        #Small {Name:SmallTxt; font-family:Arial;font-weight:normal;font-size:10pt;color:#ffffff;}
        #Big {Name:BigTxt; font-family:Arial;font-weight:bold;font-size:12pt;color:#ffffff;}

        .ENCC {Name:English; lang: en-US; SAMI_Type: CC;}

        -->

        </style>
        </head>
        <body>
        <sync start="0"><p class="ENCC">
        <sync start="5905"><p class="ENCC">&gt;&gt;&gt; PRESENTATION OF &#34;IDAHO<br`,
			expectedErr: false,
		},
		{
			name:  "can handle tag nesting properly",
			input: strings.NewReader(nestedTags),
			expected: `
        <sami>
          <head>
          </head>
          <body>
            <p>
              something
              <span>
                <span>
                  <span>
                    <span>
                      else
                    </span>
                  </span>
                </span>
              </span>
            </p>
            <p>another p</p>
          </body>
        </sami>`,
			expectedErr: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			root, err := Parse(test.input)
			if err != nil && !test.expectedErr {
				tt.Errorf("got unexpected error: %s", err)
			}
			s := render(root)
			if e, g := diff.TrimLinesInString(test.expected), diff.TrimLinesInString(s); e != g {
				tt.Errorf("unexpected result:\n%v", diff.LineDiff(e, g))
			}
		})
	}
}

const (
	ttsSample = `
  <SAMI>
  <HEAD>
  </HEAD>
  <BODY>
  <SYNC start="6534661"><P class="ENCC">
      <span tts:fontStyle="italic">NOVA</span> is <span foo="bar">a</span> production<br/>
         of WGBH Boston.
</P></SYNC>
</BODY>
  `
	pycaptionSample = `<SAMI>
  <HEAD>
    <TITLE>NOVA3213</TITLE>
    <STYLE TYPE="text/css">
    <!--
    P { margin-left:  1pt;
        margin-right: 1pt;
        margin-bottom: 2pt;
        margin-top: 2pt;
        text-align: center;
        font-size: 10pt;
        font-family: Arial;
        font-weight: normal;
        font-style: normal;
        color: #ffeedd; }

    .ENCC {Name: English; lang: en-US; SAMI_Type: CC;}

    -->
    </STYLE>
  </HEAD>
  <BODY>
    <SYNC start="9209">
      <P class="ENCC">
        ( clock ticking )
      </P>
    </SYNC>
    <SYNC start="12312"><P class="ENCC">&nbsp;</P></SYNC>
    <SYNC start="14848">
      <P class="ENCC">
        MAN:<br/>
        When we think<br/>
        of "E equals m c-squared",
      </P>
    </SYNC>
    <SYNC start="17000">
      <P class="ENCC">
        <SPAN Style="text-align:right;">we have this vision of Einstein</SPAN>
      </P>
    </SYNC>
    <SYNC start="18752">
      <P class="ENCC">
        as an old, wrinkly man<br/>
        with white hair.
      </P>
    </SYNC>
    <SYNC start="20887">
      <P class="ENCC">
        MAN 2:<br/>
        E equals m c-squared is<br/>
        not about an old Einstein.
      </P>
    </SYNC>
    <SYNC start="26760">
      <P class="ENCC">
        MAN 2:<br/>
        It's all about an eternal Einstein.
      </P>
    </SYNC>
    <SYNC start="32200">
      <P class="ENCC">
        &lt;LAUGHING &amp; WHOOPS!&gt;
      </P>
    </SYNC>
    <SYNC start="34400">
      <P class="ENCC">
        <br/>some more text
      </P>
    </SYNC>
  </BODY>
</SAMI>`
	sample1 = `
<SAMI>
  <HEAD>
    <TITLE>test123</TITLE>
    <STYLE type="text/css">
      <!--
      P { margin-right:  1pt; }
      -->
    </STYLE>
    <STYLE TYPE="text/css">
    <!--
    P { margin-left:  1pt; color: #ffeedd; }
    .ENCC {Name: English; lang: en-US; SAMI_Type: CC;}
    -->
    </STYLE>
  </HEAD>
  <BODY>
    <SYNC start="9209">
      <P class="ENCC">
        ( clock ticking )
      </P>
    </SYNC>
    <SYNC start="12312"><P class="ENCC">&nbsp;</P></SYNC>
    <SYNC start="14848">
      <P class="ENCC">
        MAN:<br/>
        When we think<br/>
        of "E equals m c-squared",
      </P>
    </SYNC>
  </BODY>
</SAMI>`

	sample2 = `<SAMI>
  <HEAD>
    <TITLE>test123</TITLE>
    <STYLE TYPE="text/css">
    <!--
    P { margin-left:  1pt; color: #ffeedd; }
    .ENCC {Name: English; lang: en-US; SAMI_Type: CC;}
    -->
    </STYLE>
  </HEAD>
  <BODY>
    <SYNC start="12312"><P class="ENCC">&nbsp;</P></SYNC>
    <SYNC start="17000">
      <P class="ENCC">
        <SPAN Style="text-align:right;">we have this vision of Einstein</SPAN>
      </P>
    </SYNC>
    <SYNC start="32200">
      <P class="ENCC">
        &lt;LAUGHING &amp; WHOOPS!&gt;
      </P>
    </SYNC>
  </BODY>
</SAMI>`

	syntaxErrorSample = `
<SAMI>
<Head>
<title>ir2014_111</title>
  <STYLE TYPE="text/css">
    <!--
    P { margin-left:  1pt;
      margin-right: 1pt;
      margin-bottom: 2pt;
      margin-top: 2pt;
      text-align: center;
      font-size: 10pt;
      font-family: Arial;
      font-weight: normal;
      font-style: normal;
      color: #ffffff; }

    #Small {Name:SmallTxt; font-family:Arial;font-weight:normal;font-size:10pt;color:#ffffff;}
    #Big {Name:BigTxt; font-family:Arial;font-weight:bold;font-size:12pt;color:#ffffff;}

    .ENCC {Name:English; lang: en-US; SAMI_Type: CC;}

    -->

  </Style>
</Head>
<BODY>
<Sync Start=0><P Class=ENCC>
<Sync Start=5905><P Class=ENCC>>>> PRESENTATION OF "IDAHO<br>REPORTS" ON IDAHO PUBLIC
<Sync Start=7073><P Class=ENCC>TELEVISION IS MADE POSSIBLE<br>THROUGH THE GENEROUS SUPPORT OF

</Body>
</SAMI>
`
	nestedTags = `
  <sami>
    <head>
    </head>
  <body>
    <p>
      something
      <span>
        <span>
          <span>
            <span>
              else
            </span>
          </span>
        </span>
      </span>
    </p>
    <p>another p</p>
  </body>
  `
)
