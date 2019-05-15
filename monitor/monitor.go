package monitor

import (
	"fmt"
	"net"

	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/core/tcp"
	"github.com/amsalt/nginet/handler"
)

// monitor follow the telent protocol.

const (
	ReadBufferSize   = 1024 * 10
	WriteBufferSize  = 1024 * 1024
	MaxConnectionNum = 5
)

type Monitor struct {
	server core.AcceptorChannel
	port   string
}

func NewMonitor(port string) *Monitor {
	m := &Monitor{port: port}
	return m
}

func (m *Monitor) Init() {
	m.server = core.GetAcceptorBuilder(core.TCPServBuilder).Build(
		tcp.WithReadBufSize(ReadBufferSize),
		tcp.WithWriteBufSize(WriteBufferSize),
		tcp.WithMaxConnNum(MaxConnectionNum),
	)

	m.server.InitSubChannel(func(channel core.SubChannel) {
		channel.Pipeline().AddLast(nil, "StringEncoder", handler.NewStringEncoder())
		channel.Pipeline().AddLast(nil, "TextHandler", NewTextHandler())
	})
}

func (m *Monitor) Start() {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf(":%s", m.port))
	if err != nil {
		panic("bad net addr")
	}

	m.server.Listen(addr)
	go m.server.Accept()
}

func (m *Monitor) Stop() {
	m.server.Close()
}
