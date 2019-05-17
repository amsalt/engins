package test

import (
	"testing"
	"time"

	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/balancer/stickiness"
	"github.com/amsalt/cluster/resolver/static"
	"github.com/amsalt/engins"
	"github.com/amsalt/engins/scluster"
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/gnetlog"
)

func TestClient(t *testing.T) {
	gnetlog.Init()

	engins.RegisterMsgByID(1, &tcpChannel{})
	resolver := static.NewConfigBasedResolver()
	c := scluster.NewCluster(resolver)
	c.Init()

	// player client role
	// connect gate server.
	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("gate"), stickiness.WithResolver(resolver))

	c.BuildClient("gate", "player", scluster.WithBalancer(b))

	c.Start()
	resolver.Register("gate", ":7878")
	time.Sleep(time.Second * 3)

	// player client role.
	// write message to gate server.
	err := c.Write("gate", &tcpChannel{Msg: "cluster send message1"})
	log.Errorf("send message result: %+v", err)
	time.Sleep(time.Second * 111)
}
