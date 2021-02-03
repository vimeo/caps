package sami

import (
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/thiagopnts/caps"
)

const defaultLanguageCode = "en-US"

type Reader struct {
}

func (r Reader) Detect(content []byte) bool {
	return r.DetectString(string(content))
}

func (Reader) DetectString(content string) bool {
	return strings.Contains(strings.ToLower(content), "<sami")
}

func (r Reader) ReadString(content string) (*caps.CaptionSet, error) {
	if strings.Contains(strings.ToLower(content), "<html") {
		return nil, fmt.Errorf("sami file seems to be an html file")
	}
	doc, err := htmlquery.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	fmt.Println(doc)
	return nil, nil
}
