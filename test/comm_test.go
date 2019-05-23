package test

import (
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/gnetlog"
)

func init() {
	gnetlog.Init()
	log.Debugf("2")
}

type tcpChannel struct {
	Msg string
}
