package conf

import (
	"encoding/json"
	"io/ioutil"

	"github.com/amsalt/log"
)

func init() {
	Register(&JSONParser{})
}

type JSONParser struct {
}

func (j *JSONParser) Name() string {
	return "json"
}

// Parse parses JSON format file in path to target.
func (j *JSONParser) Parse(path string, target interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		log.Errorf("%v", err)
	}
	return err
}
