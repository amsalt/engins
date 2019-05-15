package conf

type Parser interface {
	Name() string
	Parse(path string, target interface{}) error
}

var parsers map[string]Parser

func init() {
	parsers = make(map[string]Parser)
}

func Register(parser Parser) {
	parsers[parser.Name()] = parser
}
