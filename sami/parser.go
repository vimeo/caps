import (
	"github.com/gammazero/deque"
)

type Parser struct {
	sami           string
	line           string
	styles         map[string]interface{}
	langs          map[string]interface{}
	queue          deque.Deque
	lastElement    string
	name2codepoint map[string]int
}

func NewParse() *Parser {
	return &Parser{
		sami:           "",
		line:           "",
		styles:         map[string]interface{}{},
		langs:          map[string]interface{}{},
		lastElement:    "",
		name2codepoint: map[string]interface{}{"apos": 0x0027},
	}
}



