package test

import (
	"testing"
	"time"

	"github.com/amsalt/engins/monitor"
	"github.com/amsalt/nginet/gnetlog"
)

func TestMonitor(t *testing.T) {
	gnetlog.Init()
	m := monitor.NewMonitor("7878")
	m.Init()
	m.Start()

	// to test this case, use `telnet 127.0.0.1 7878` and
	// type some message and hit the `enter` key.
	time.Sleep(time.Second * 60)
}
