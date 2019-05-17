package test

import (
	"testing"
	"time"

	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/balancer/stickiness"
	"github.com/amsalt/cluster/resolver/static"
	"github.com/amsalt/engins"
	"github.com/amsalt/engins/scluster"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/gnetlog"
)

func TestGate(t *testing.T) {
	gnetlog.Init()

	engins.RegisterMsgByID(1, &tcpChannel{})

	resolver := static.NewConfigBasedResolver()

	c := scluster.NewCluster(resolver)
	c.Init()
	c.RegisterRouter(1, "game")

	// gate server role
	// for player client connecting.
	c.BuildServer("gate", ":7878", core.TCPServBuilder, scluster.WithServerRelay(true))
	// to connect game server.
	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("game"), stickiness.WithResolver(resolver))
	c.BuildClient("game", "gate", scluster.WithBalancer(b))
	c.Start()
	resolver.Register("game", ":7879")

	// time.Sleep(time.Second * 5)
	// err := c.Write("game", &tcpChannel{Msg: "cluster send message1"})
	// log.Errorf("send message result: %+v", err)

	time.Sleep(time.Second * 600)
}
