package conversion

import (
	"testing"

	"github.com/vimeo/caps/dfxp"
	"github.com/vimeo/caps/scc"
	"github.com/vimeo/caps/srt"
	"github.com/vimeo/caps/webvtt"
)

func TestSAMIConversion(t *testing.T) {
	scc.NewReader(false, 0)
	srt.NewReader()
	webvtt.NewReader(false)
	dfxp.NewReader()
}
