package test

import (
	"testing"

	"github.com/amsalt/engins"
	"github.com/amsalt/engins/cluster"
	"github.com/amsalt/log"
	"github.com/amsalt/ngicluster/balancer"
	"github.com/amsalt/ngicluster/balancer/stickiness"
	"github.com/amsalt/ngicluster/resolver/static"
	"github.com/amsalt/nginet/core"
)

func TestClient(t *testing.T) {
	engins.RegisterMsgByID(1, &tcpChannel{})
	resolver := static.NewConfigBasedResolver()
	c := cluster.NewCluster(resolver)

	resolver.Register("gate", "localhost:7878")

	// player client role
	// connect gate server.
	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("gate"), stickiness.WithResolver(resolver))
	c.BuildClient("gate", "player", cluster.WithBalancer(b), cluster.WithOnConnect(func(*core.ChannelContext, core.Channel) {
		// player client role.
		// write message to gate server.
		// Assert relay to `game` server.
		err := c.Write("gate", &tcpChannel{Msg: "client send message"})
		log.Infof("send message result: %+v", err)
	}), cluster.WithOnDisConnect(func(ctx *core.ChannelContext) {
		log.Errorf("client connection closed")
	}))

	engins.Run(c)
}
