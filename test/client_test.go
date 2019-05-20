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
)

func TestClient(t *testing.T) {
	engins.RegisterMsgByID(1, &tcpChannel{})
	resolver := static.NewConfigBasedResolver()
	c := cluster.NewCluster(resolver)
	c.Init()

	// player client role
	// connect gate server.
	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("gate"), stickiness.WithResolver(resolver))
	c.BuildClient("gate", "player", cluster.WithBalancer(b))

	c.Start()
	resolver.Register("gate", ":7878")
	time.Sleep(time.Second * 3)

	// player client role.
	// write message to gate server.
	// Assert relay to `game` server.
	err := c.Write("gate", &tcpChannel{Msg: "cluster send message1"})
	log.Infof("send message result: %+v", err)

	time.Sleep(time.Second * 211)
}
