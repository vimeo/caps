package caps

import (
	"fmt"
	"math"
	"strings"
)

type Caption struct {
	Start int
	End   int
	Nodes []CaptionContent
	Style StyleProps
}

func (c Caption) IsEmpty() bool {
	return len(c.Nodes) == 0
}

func (c Caption) GetText() string {
	var content strings.Builder
	for _, node := range c.Nodes {
		if !node.IsStyle() {
			content.WriteString(node.GetContent())
		}
	}
	return content.String()
}

func (c Caption) String() string {
	return fmt.Sprintf("%s --> %s\n%s", c.FormatStart(), c.FormatEnd(), c.GetText())
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

func formatTimestamp(value int, sep string) string {
	value /= 1000
	seconds := math.Mod(float64(value)/1000, 60)
	minutes := (value / (1000 * 60)) % 60
	hours := (value / (1000 * 60 * 60) % 24)
	timestamp := fmt.Sprintf("%02d:%02d:%06.3f", hours, minutes, seconds)
	if sep != "." {
		return strings.ReplaceAll(timestamp, ".", sep)
	}
	return timestamp
}

func NewCaption(start, end int, nodes []CaptionContent, style StyleProps) Caption {
	return Caption{
		start,
		end,
		nodes,
		style,
	}
}