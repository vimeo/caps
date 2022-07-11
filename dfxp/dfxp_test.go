package dfxp

import (
	"encoding/xml"
	"fmt"
	"testing"
)

func TestStructToXML(t *testing.T) {
	base := Span{xml.Name{}, "bla blablab bla alba", Style{TTSTextAlign: "center"}}
	output, err := xml.MarshalIndent(base, "  ", "    ")
	fmt.Println(err)
	fmt.Println(string(output))
}
