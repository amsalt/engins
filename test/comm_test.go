package test

import (
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/gnetlog"
)

func init() {
	gnetlog.Init()
	log.Debugf("15")
}

type tcpChannel struct {
	Msg string
}
