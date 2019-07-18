package test

import (
	"testing"
	"time"

	"github.com/amsalt/engins"
	"github.com/amsalt/engins/cluster"
	"github.com/amsalt/log"
	"github.com/amsalt/ngicluster/balancer"
	"github.com/amsalt/ngicluster/balancer/stickiness"
	"github.com/amsalt/ngicluster/resolver/static"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/gnetlog"
)

func TestSCluster(t *testing.T) {
	gnetlog.Init()

	engins.RegisterMsgByID(1, &tcpChannel{})
	engins.RegisterProcessorByID(1, func(ctx *core.ChannelContext, msg interface{}, args ...interface{}) {
		if m, ok := msg.([]byte); ok {
			log.Infof("received message %+v from server %+v ", string(m), ctx.Attr().Value(cluster.ChannelNameKey))
		} else {
			log.Infof("received message %+v from server %+v ", msg, ctx.Attr().Value(cluster.ChannelNameKey))
		}
	})

	resolver := static.NewConfigBasedResolver()
	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("game"), stickiness.WithResolver(resolver))

	c := cluster.NewCluster(resolver)
	c.Init()
	c.RegisterRelayRouter(1, "game")

	// gate server role
	// for player client connecting.
	c.BuildServer("gate", ":7878", core.TCPServBuilder, cluster.WithServerRelay(true))
	// to connect game server.
	c.BuildClient("game", "gate", cluster.WithBalancer(b))

	// game server role
	// for gate server connect.
	c.BuildServer("game", ":7879", core.TCPServBuilder, cluster.WithServerRelay(false))

	// player client role
	// connect gate server.
	// c.BuildClient("gate", "player", cluster.WithBalancer(b))

	c.Start()

	// time.Sleep(time.Second * 10)

	// player client role.
	// write message to gate server.
	// err := c.Write("gate", &tcpChannel{Msg: "cluster send message1"})
	// log.Errorf("send message result: %+v", err)
	time.Sleep(time.Second * 100)
}
