package conf

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/amsalt/engins/errs"
)

// Configurator represents a configuration parser.
type Configurator struct {
	baseDir string
}

// NewConfigurator creates a new Configurator instance.
// If baseDir not set, current working dir will be used.
func NewConfigurator(baseDir ...string) *Configurator {
	c := &Configurator{}
	if len(baseDir) > 0 {
		c.baseDir = baseDir[0]
	} else {
		c.baseDir = c.getWorkDir()
	}

	return c
}

// Parse parses config files in paths to target.
// support different config file types.
func (c *Configurator) Parse(target interface{}, paths ...string) {
	for _, path := range paths {
		fields := strings.Split(path, ".")
		suffix := fields[len(fields)-1]
		fullPath := c.baseDir + path

		parser := parsers[suffix]
		if parser == nil {
			panic(errs.NewUnsupportedType(suffix))
		} else {
			parser.Parse(fullPath, target)
		}
	}
}

func (c *Configurator) getWorkDir() string {
	exec, err := os.Executable()
	if err != nil {
		panic("Get exec path fail!")
	}

	return filepath.Dir(exec)
}
