package test

import "github.com/amsalt/nginet/gnetlog"

func init() {
	gnetlog.Init()
}

type tcpChannel struct {
	Msg string
}
