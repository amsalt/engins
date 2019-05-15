package conf

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

func init() {
	Register(&YamlParser{})
}

type YamlParser struct {
}

func (y *YamlParser) Name() string {
	return "yaml"
}

// Parse parses yaml file in path to target.
func (y *YamlParser) Parse(path string, target interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal([]byte(data), target)
}
