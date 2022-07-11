package scc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/thiagopnts/caps"
)

type Reader struct {
	scc              []*caps.Caption
	time             string
	popBuffer        string
	paintBuffer      string
	lastCommand      string
	rollRows         []string
	rollRowsExpected int
	paintTime        float64
	popTime          float64
	popOn            bool
	paintOn          bool
	simulateRollUp   bool
	openItalic       bool
	firstElement     bool
	frameCount       int
	offset           int
}

var timestampWords = regexp.MustCompile(`([0-9:;]*)([\s\t]*)((.)*)`)

func (Reader) Detect(content []byte) bool {
	return strings.HasPrefix(strings.TrimLeft(string(content), " "), header)
}

func (r *Reader) Read(content []byte) (*caps.CaptionSet, error) {
	lines := strings.Split(string(content), "\n")
	for _, line := range lines[1:] {
		r.translateLine(line)
	}
	if r.paintBuffer != "" {
		r.rollUp()
	}
	set := caps.NewCaptionSet()
	set.SetCaptions(caps.DefaultLang, r.scc)
	if set.IsEmpty() {
		return set, fmt.Errorf("empty caption file")
	}
	return set, nil
}

func (r *Reader) translateLine(line string) {
	if strings.Trim(line, " ") == "" {
		return
	}
	parts := timestampWords.FindAllStringSubmatch(strings.ToLower(line), -1)
	if len(parts) == 0 {
		return
	}
	r.time = string(parts[0][1])
	r.frameCount = 0
	for _, word := range strings.Split(parts[0][3], " ") {
		if word != "" {
			r.translateWord(word)
		}
	}
}

func (r *Reader) translateWord(word string) {
	r.frameCount += 1
	if _, ok := commands[word]; ok {
		r.translateCommand(word)
	} else if _, ok := specialChars[word]; ok {
		r.translateSpecialChar(word)
	} else if _, ok := extendedChars[word]; ok {
		r.translateExtendedChar(word)
	} else {
		r.translateCharacters(word)
	}
}

func (r *Reader) translateExtendedChar(word string) {
	if r.handleDoubleCommand(word) {
		return
	}
	if _, ok := extendedChars[word]; !ok {
		return
	}
	if r.paintOn {
		if r.paintBuffer != "" {
			r.paintBuffer = r.paintBuffer[:len(r.paintBuffer)-1]
		}
		r.paintBuffer += extendedChars[word]
		return
	}
	if r.popBuffer != "" {
		r.popBuffer = r.popBuffer[:len(r.popBuffer)-1]
	}
	r.popBuffer += extendedChars[word]
}

func (r *Reader) translateCharacters(word string) {
	if len(word) < 2 {
		return
	}
	byte1 := word[:2]
	byte2 := word[2:]
	if _, ok := characters[byte1]; !ok {
		return
	}
	if _, ok := characters[byte2]; !ok {
		return
	}
	if r.paintOn {
		r.paintBuffer += characters[byte1] + characters[byte2]
		return
	}
	r.popBuffer += characters[byte1] + characters[byte2]
}

func (r *Reader) translateSpecialChar(word string) {
	if r.handleDoubleCommand(word) {
		return
	}
	if _, ok := specialChars[word]; !ok {
		return
	}
	if r.paintOn {
		r.paintBuffer += specialChars[word]
		return
	}
	r.popBuffer += specialChars[word]
}

func (r *Reader) translateCommand(word string) error {
	if r.handleDoubleCommand(word) {
		return nil
	}
	if word == "9420" {
		r.popOn = true
		r.paintOn = false
	} else if word == "9429" || word == "9425" || word == "9426" || word == "94a7" {
		r.paintOn = true
		r.popOn = false
		switch word {
		case "9429":
			r.rollRowsExpected = 1
		case "9425":
			r.rollRowsExpected = 2
		case "9426":
			r.rollRowsExpected = 3
		case "94a7":
			r.rollRowsExpected = 4
		}
		if r.paintBuffer != "" {
			r.convertToCaption(r.paintBuffer, r.paintTime)
			r.paintBuffer = ""
		}
		r.rollRows = []string{}
		paintTime, err := r.translateCurrentTime()
		if err != nil {
			return err
		}
		r.paintTime = paintTime
	} else if word == "94ae" {
		r.popBuffer = ""
	} else if word == "942f" && r.popBuffer != "" {
		popTime, err := r.translateCurrentTime()
		if err != nil {
			return err
		}
		r.popTime = popTime
		r.convertToCaption(r.popBuffer, r.popTime)
		r.popBuffer = ""
	} else if word == "94ad" {
		if r.paintBuffer != "" {
			r.rollUp()
		}
	} else if word == "942c" {
		r.rollRows = []string{}
		if r.paintBuffer != "" {
			r.rollUp()
		}
		if len(r.scc) > 0 && r.scc[len(r.scc)-1].End == 0 {
			lastTime, err := r.translateCurrentTime()
			if err != nil {
				return err
			}
			r.scc[len(r.scc)-1].End = lastTime
		}
	} else {
		if r.paintOn {
			r.paintBuffer += commands[word]
			return nil
		}
		r.popBuffer += commands[word]
	}
	return nil
}

func (r *Reader) rollUp() error {
	if !r.simulateRollUp {
		r.rollRows = []string{}
	}
	if len(r.rollRows) >= r.rollRowsExpected {
		r.rollRows = r.rollRows[1:]
	}
	r.rollRows = append(r.rollRows, r.paintBuffer)
	r.paintBuffer = strings.Join(r.rollRows, " ")
	r.convertToCaption(r.paintBuffer, r.paintTime)
	r.paintBuffer = ""
	paintTime, err := r.translateCurrentTime()
	if err != nil {
		return err
	}
	r.paintTime = paintTime
	if len(r.scc) == 0 {
		return nil
	}
	r.scc[len(r.scc)-1].End = r.paintTime
	return nil
}

func (r *Reader) handleDoubleCommand(word string) bool {
	if word == r.lastCommand {
		r.lastCommand = ""
		return true
	}
	r.lastCommand = word
	return false
}

func (r *Reader) convertToCaption(buffer string, start float64) {
	if len(r.scc) > 0 && r.scc[len(r.scc)-1].End == 0 {
		r.scc[len(r.scc)-1].End = r.scc[len(r.scc)-1].Start
	}
	r.openItalic = false
	r.firstElement = true
	caption := &caps.Caption{Start: start, End: 0}
	for _, element := range strings.Split(buffer, "<$>") {
		if strings.Trim(element, " ") == "" {
			continue
		} else if element == "{break}" {
			r.translateBreak(caption)
		} else if element == "{italic}" {
			style := caps.DefaultStyleProps()
			style.Italics = true
			caption.Nodes = append(caption.Nodes, caps.NewCaptionStyle(true, style))
			r.openItalic = true
			r.firstElement = false
		} else if element == "{end-italic}" && r.openItalic {
			style := caps.DefaultStyleProps()
			style.Italics = true
			caption.Nodes = append(caption.Nodes, caps.NewCaptionStyle(false, style))
			r.openItalic = false
		} else {
			//FIXME this is ' '.join(element.split())
			caption.Nodes = append(caption.Nodes, caps.NewCaptionText(element))
			r.firstElement = false
		}
	}
	// close any open italics left
	if r.openItalic == true {
		style := caps.DefaultStyleProps()
		style.Italics = true
		caption.Nodes = append(caption.Nodes, caps.NewCaptionStyle(false, style))
	}
	r.removeExtraItalics(caption)
	if len(caption.Nodes) > 0 {
		r.scc = append(r.scc, caption)
	}
}

func (r *Reader) translateBreak(caption *caps.Caption) {
	if r.firstElement {
		return
	} else if len(caption.Nodes) > 0 && caption.Nodes[len(caption.Nodes)-1].LineBreak() {
		return
	} else if r.openItalic {
		style := caps.DefaultStyleProps()
		style.Italics = true
		caption.Nodes = append(caption.Nodes, caps.NewCaptionStyle(false, style))
		r.openItalic = false
	}
	caption.Nodes = append(caption.Nodes, caps.NewLineBreak())
}

func (r *Reader) removeExtraItalics(caption *caps.Caption) {
	i := 0
	length := len(caption.Nodes)
	if length < 0 {
		length = 0
	}
	for i < length {
		if caption.Nodes[i].Style() && caption.Nodes[i+1].LineBreak() && caption.Nodes[i+2].Style() {
			style := caption.Nodes[i].(caps.CaptionStyle)
			style2 := caption.Nodes[i+2].(caps.CaptionStyle)
			if style.Props.Italics && style2.Props.Italics {
				caption.Nodes[i] = caption.Nodes[len(caption.Nodes)-1]
				caption.Nodes[i+1] = caption.Nodes[len(caption.Nodes)-2]
				caption.Nodes = caption.Nodes[:len(caption.Nodes)-2]
			}
			length -= 2
		}
		i += 1
	}
}

func (r *Reader) translateCurrentTime() (float64, error) {
	n, err := strconv.Atoi(r.time[len(r.time)-2:])
	if err != nil {
		return 0, err
	}
	currStamp := fmt.Sprintf("%s%d", r.time[:len(r.time)-2], n+r.frameCount)
	secondsPerTimestampSecond := 1001.0 / 1000.0
	if strings.Contains(currStamp, ";") {
		secondsPerTimestampSecond = 1.0
	}
	timesplit := strings.Split(strings.ReplaceAll(strings.Trim(currStamp, " "), ";", ":"), ":")
	t0, err := strconv.ParseFloat(timesplit[0], 64)
	if err != nil {
		return 0, err
	}
	t1, err := strconv.ParseFloat(timesplit[1], 64)
	if err != nil {
		return 0, err
	}
	t2, err := strconv.ParseFloat(timesplit[2], 64)
	if err != nil {
		return 0, err
	}
	t3, err := strconv.ParseFloat(timesplit[3], 64)
	if err != nil {
		return 0, err
	}
	timestampSeconds := float64(t0*3600 +
		t1*60 +
		t2 +
		t3/30.0)
	microseconds := timestampSeconds * 1000.0 * 1000.0
	microseconds -= float64(r.offset)
	microseconds *= secondsPerTimestampSecond
	if microseconds < 0 {
		return 0, nil
	}
	return microseconds, nil
}
