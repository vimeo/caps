package scc

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vimeo/caps"
)

func TestDetect(t *testing.T) {
	if !DefaultReader().Detect(sampleSCC) {
		t.Error("valid scc sample should be detected")
	}
	if DefaultReader().Detect(badSampleSCC) {
		t.Error("invalid scc sample should be detected")
	}
}

func TestCaptionLength(t *testing.T) {
	captionSet, err := DefaultReader().Read(sampleSCC)
	if err != nil {
		t.Errorf("scc reader failed: %v", err)
	}
	captions := captionSet.GetCaptions(caps.DefaultLang)
	if len(captions) != 7 {
		t.Errorf("expected captions len to be %d, but got %d", 7, len(captions))
	}
}

func TestCaptionContent(t *testing.T) {
	captionSet, err := DefaultReader().Read(sampleSCC)
	if err != nil {
		t.Errorf("scc reader failed: %v", err)
	}
	captions := captionSet.GetCaptions(caps.DefaultLang)
	paragraph := captions[2]
	if math.Abs(float64(paragraph.Start)-17000000) > toleranceMicroseconds {
		t.Error("paragraph start timestamp over microseconds tolerance")
	}
	if math.Abs(float64(paragraph.End)-18752000) > toleranceMicroseconds {
		t.Error("paragraph end timestamp over microseconds tolerance")
	}
}

func TestEmptyFile(t *testing.T) {
	_, err := DefaultReader().Read(sampleSCCempty)
	if err == nil {
		t.Errorf("should have returned an error")
	}
}

func TestWriter(t *testing.T) {
	type sccToSCCTests struct {
		inputSCC []byte
		wantSCC  []byte
	}
	srtConvertionTests := []sccToSCCTests{
		{inputSCC: sampleSCC, wantSCC: sccSamplePostCapsConvertion},
	}
	for _, test := range srtConvertionTests {
		captionsSet, err := DefaultReader().Read(test.inputSCC)
		assert.Nil(t, err)
		result, _ := NewWriter().Write(captionsSet)
		assert.Equal(t, test.wantSCC, result)
	}
}

const toleranceMicroseconds = 500 * 1000

var sampleSCC = []byte(`Scenarist_SCC V1.0

00:00:09:05	94ae 94ae 9420 9420 9470 9470 a820 e3ec efe3 6b20 f4e9 e36b e96e 6720 2980 942c 942c 942f 942f

00:00:12:08	942c 942c

00:00:13:18	94ae 94ae 9420 9420 1370 1370 cdc1 ceba 94d0 94d0 5768 e56e 20f7 e520 f468 e96e 6b80 9470 9470 efe6 20a2 4520 e5f1 7561 ec73 206d 20e3 ad73 f175 61f2 e564 a22c 942c 942c 942f 942f

00:00:16:03	94ae 94ae 9420 9420 9470 9470 f7e5 2068 6176 e520 f468 e973 2076 e973 e9ef 6e20 efe6 2045 e96e 73f4 e5e9 6e80 942c 942c 942f 942f

00:00:17:20	94ae 94ae 9420 9420 94d0 94d0 6173 2061 6e20 efec 642c 20f7 f2e9 6e6b ec79 206d 616e 9470 9470 f7e9 f468 20f7 68e9 f4e5 2068 61e9 f2ae 942c 942c 942f 942f

00:00:19:13	94ae 94ae 9420 9420 1370 1370 cdc1 ce20 32ba 94d0 94d0 4520 e5f1 7561 ec73 206d 20e3 ad73 f175 61f2 e564 20e9 7380 9470 9470 6eef f420 6162 ef75 f420 616e 20ef ec64 2045 e96e 73f4 e5e9 6eae 942c 942c 942f 942f

00:00:25:16	94ae 94ae 9420 9420 1370 1370 cdc1 ce20 32ba 94d0 94d0 49f4 a773 2061 ecec 2061 62ef 75f4 2061 6e20 e5f4 e5f2 6e61 ec80 9470 9470 45e9 6e73 f4e5 e96e ae80 942c 942c 942f 942f

00:00:31:15	94ae 94ae 9420 9420 9470 9470 bc4c c1d5 c7c8 49ce c720 2620 57c8 4f4f d0d3 a13e 942c 942c 942f 942f

00:00:36:04	942c 942c

`)

var sccSamplePostCapsConvertion = []byte(`Scenarist_SCC V1.0

00:00:09:23	94ae 94ae 9420 9420 9470 9470 a820 e3ec efe3 6b20 f4e9 e36b e96e 6720 2980 942c 942c 942f 942f

00:00:12:08	942c 942c

00:00:14:12	94ae 94ae 9420 9420 1370 1370 cdc1 ceba 94d0 94d0 5768 e56e 20f7 e520 f468 e96e 6b80 9470 9470 efe6 20a2 4520 e5f1 7561 ec73 206d 20e3 ad73 f175 61f2 e564 a22c 942c 942c 942f 942f

00:00:16:22	94ae 94ae 9420 9420 9470 9470 f7e5 2068 6176 e520 f468 e973 2076 e973 e9ef 6e20 efe6 2045 e96e 73f4 e5e9 6e80 942c 942c 942f 942f

00:00:18:12	94ae 94ae 9420 9420 94d0 94d0 6173 2061 6e20 efec 642c 20f7 f2e9 6e6b ec79 206d 616e 9470 9470 f7e9 f468 20f7 68e9 f4e5 2068 61e9 f2ae 942c 942c 942f 942f

00:00:20:10	94ae 94ae 9420 9420 1370 1370 cdc1 ce20 32ba 94d0 94d0 4520 e5f1 7561 ec73 206d 20e3 ad73 f175 61f2 e564 20e9 7380 9470 9470 6eef f420 6162 ef75 f420 616e 20ef ec64 2045 e96e 73f4 e5e9 6eae 942c 942c 942f 942f

00:00:26:09	94ae 94ae 9420 9420 1370 1370 cdc1 ce20 32ba 94d0 94d0 49f4 a773 2061 ecec 2061 62ef 75f4 2061 6e20 e5f4 e5f2 6e61 ec80 9470 9470 45e9 6e73 f4e5 e96e ae80 942c 942c 942f 942f

00:00:31:29	94ae 94ae 9420 9420 9470 9470 bc4c c1d5 c7c8 49ce c720 2620 57c8 4f4f d0d3 a13e 942c 942c 942f 942f

00:00:36:04	942c 942c

`)

var sampleSCCempty = []byte(`Scenarist_SCC V1.0
`)

var badSampleSCC = []byte(`
00:00:20:10	94ae 94ae 9420 9420 1370 1370 cdc1 ce20 32ba 94d0 94d0 4520 e5f1 7561 ec73 206d 20e3 ad73 f175 61f2 e564 20e9 7380 9470 9470 6eef f420 6162 ef75 f420 616e 20ef ec64 2045 e96e 73f4 e5e9 6eae 942c 942c 942f 942f

00:00:26:09	94ae 94ae 9420 9420 1370 1370 cdc1 ce20 32ba 94d0 94d0 49f4 a773 2061 ecec 2061 62ef 75f4 2061 6e20 e5f4 e5f2 6e61 ec80 9470 9470 45e9 6e73 f4e5 e96e ae80 942c 942c 942f 942f

00:00:31:29	94ae 94ae 9420 9420 9470 9470 bc4c c1d5 c7c8 49ce c720 2620 57c8 4f4f d0d3 a13e 942c 942c 942f 942f

00:00:36:04	942c 942c

`)
