package components

import (
	"sync"

	"github.com/amsalt/nginet/safe"
)

// ensure all componets exit.
var wg sync.WaitGroup
var components []Component

type Component interface {
	// Lifecycle control
	Init()
	Start()
	Stop()
}

func Register(c ...Component) {
	components = append(components, c...)
}

func Run() {
	for _, c := range components {
		c.Init()
	}

	for _, c := range components {
		go c.Start()
	}
}

func Stop() {
	for _, c := range components {
		safe.Call(func() {
			c.Stop()
		})
	}
}
