package gocaption

import (
	"log"
	"strconv"
	"strings"

	"github.com/anaskhan96/soup"
)

const defaultLanguageCode = "en-US"

type DFXReader struct {
	framerate  string
	multiplier []int
	timebase   string
	nodes      []captionNode
}

func (DFXReader) Detect(content string) bool {
	return strings.Contains(strings.ToLower(content), "</tt>")
}

func (r DFXReader) Read(content string) CaptionSet {
	doc := soup.HTMLParse(content)
	captions := CaptionSet{}
	tts := doc.FindAll("tt")
	if len(tts) >= 1 {
		attrs := tts[0].Attrs()
		if timebase, ok := attrs["ttp:timebase"]; ok {
			r.timebase = timebase
		} else {
			r.timebase = "0"
		}

		if framerate, ok := attrs["ttp:framerate"]; ok {
			r.framerate = framerate
		} else {
			r.framerate = "30"
		}

		if multiplier, ok := attrs["ttp:framemultiplier"]; ok {
			multipliers := strings.Split(multiplier, " ")
			a, err := strconv.Atoi(multipliers[0])
			if err != nil {
				log.Fatalln("failed to read multiplier")
			}
			b, err := strconv.Atoi(multipliers[1])
			if err != nil {
				log.Fatalln("failed to read multiplier")
			}
			r.multiplier = []int{a, b}
		} else {
			r.multiplier = []int{1, 1}
		}
	}
	for _, div := range doc.FindAll("div") {
		lang, ok := div.Attrs()["xml:lang"]
		if !ok {
			lang = defaultLanguageCode
		}
		captions.SetCaptions(lang, translateDiv(div))
	}
	return captions
}

func (r DFXReader) findTimes(stamp string) (int, error) {
}

func (r DFXReader) translateTime(stamp string) (int, error) {
	timesplit := strings.Split(stamp, ":")
	if !strings.Contains(timesplit[2], ".") {
		timesplit[2] = timesplit[2] + ".000"
	}
	timesplit0, err := strconv.Atoi(timesplit[0])
	if err != nil {
		return 0, err
	}
	timesplit1, err := strconv.Atoi(timesplit[1])
	if err != nil {
		return 0, err
	}
	secsplit := strings.Split(timesplit[2], ".")
	secsplit0, err := strconv.Atoi(secsplit[0])
	if err != nil {
		return 0, err
	}
	secsplit1, err := strconv.Atoi(secsplit[1])
	if err != nil {
		return 0, err
	}
	if len(timesplit) > 3 {
		timesplit3, err := strconv.ParseFloat(timesplit[3], 32)
		if err != nil {
			return 0, err
		}
		framerate, err := strconv.ParseFloat(r.framerate, 32)
		if err != nil {
			return 0, err
		}
		if r.timebase == "smpte" {
			secsplit1 = int(timesplit3 / framerate * 1000.0)
		} else {
			secsplit1 = int(float64(int(timesplit3)*r.multiplier[1]) / framerate * float64(r.multiplier[0]) * 1000.0)
		}
	}
	microseconds := int(timesplit0)*3600000000 +
		int(timesplit1)*60000000 +
		int(secsplit0)*1000000 +
		int(secsplit1)*1000

	if r.timebase == "smpte" {
		return int(float64(microseconds) * float64(r.multiplier[1]) / float64(r.multiplier[0])), nil
	}

	return microseconds, nil
}
