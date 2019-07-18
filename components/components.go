package components

import (
	"sync"

	"github.com/amsalt/nginet/safe"
)

// ensure all components exit.
var wg sync.WaitGroup
var components []Component

// Component defines the lifecycle process.
type Component interface {
	// Lifecycle control
	Init()
	Start()
	Stop()
}

// Register registers new component for run.
func Register(c ...Component) {
	components = append(components, c...)
}

// Run initializes the components and starts.
func Run() {
	for _, c := range components {
		c.Init()
	}

	for _, c := range components {
		go c.Start()
	}
}

// Stop stops all components.
func Stop() {
	for _, c := range components {
		safe.Call(func() {
			c.Stop()
		})
	}
}
