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
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/gnetlog"
)

type tcpChannel struct {
	Msg string
}

func TestSCluster(t *testing.T) {
	gnetlog.Init()

	engins.RegisterMsgByID(1, &tcpChannel{})
	engins.RegisterProcessorByID(1, func(ctx *core.ChannelContext, msg interface{}, args ...interface{}) {
		if m, ok := msg.([]byte); ok {
			log.Infof("tcpChannel handler: %+v", string(m))
		} else {
			log.Infof("tcpChannel handler: %+v", msg)
		}
	})

	resolver := static.NewConfigBasedResolver()
	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("game"), stickiness.WithResolver(resolver))

	c := scluster.NewCluster(resolver)
	c.RegisterRouter(1, "game")

	// gate server role
	c.BuildServer("gate", ":7878", core.TCPServBuilder, scluster.WithServerRelay(true))
	c.BuildClient("game", scluster.WithBalancer(b))

	// game server role
	c.BuildServer("game", ":7879", core.TCPServBuilder, scluster.WithServerRelay(false))

	// player client role
	c.BuildClient("gate", scluster.WithBalancer(b))

	c.Start()

	time.Sleep(time.Second * 3)

	// client role.
	c.Write("gate", &tcpChannel{Msg: "cluster send message1"})
	time.Sleep(time.Second * 30)

}
