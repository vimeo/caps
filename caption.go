package caps

import (
	"fmt"
	"math"
	"strings"
)

type Caption struct {
	Start float64
	End   float64
	Nodes []CaptionContent
	Style StyleProps
}

func (c Caption) Empty() bool {
	return len(c.Nodes) == 0
}

func (c Caption) Text() string {
	var content strings.Builder
	for _, node := range c.Nodes {
		if !node.Style() {
			content.WriteString(node.Content())
		}
	}
	return content.String()
}

func (c Caption) String() string {
	return fmt.Sprintf("%s --> %s\n%s", c.FormatStart(), c.FormatEnd(), c.Text())
}

func (c Caption) FormatStartWithSeparator(sep string) string {
	return formatTimestamp(c.Start, sep)
}

func (c Caption) FormatStart() string {
	return formatTimestamp(c.Start, ".")
}

func (c Caption) FormatEndWithSeparator(sep string) string {
	return formatTimestamp(c.End, sep)
}

func (c Caption) FormatEnd() string {
	return formatTimestamp(c.End, ".")
}

func formatTimestamp(timestamp float64, sep string) string {
	value := int(timestamp / 1000)
	seconds := math.Mod(float64(value)/1000, 60)
	minutes := (value / (1000 * 60)) % 60
	hours := (value / (1000 * 60 * 60) % 24)
	resultTimestamp := fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, seconds)
	if sep != "." {
		return strings.ReplaceAll(resultTimestamp, ".", sep)
	}
	return resultTimestamp
}

func NewCaption(start, end float64, nodes []CaptionContent, style StyleProps) Caption {
	return Caption{
		start,
		end,
		nodes,
		style,
	}
}
