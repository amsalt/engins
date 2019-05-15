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

	time.Sleep(time.Second * 60)
}
