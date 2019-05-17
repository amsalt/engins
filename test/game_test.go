package test

import (
	"testing"
	"time"

	"github.com/amsalt/cluster/resolver/static"
	"github.com/amsalt/engins"
	"github.com/amsalt/engins/scluster"
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/gnetlog"
)

func TestGame(t *testing.T) {
	gnetlog.Init()

	engins.RegisterMsgByID(1, &tcpChannel{})
	engins.RegisterProcessorByID(1, func(ctx *core.ChannelContext, msg interface{}, args ...interface{}) {
		if m, ok := msg.([]byte); ok {
			if len(args) > 0 {
				log.Infof("received message %+v, %+v ", string(m), args[0])
			} else {
				log.Infof("received message %+v", string(m))
			}

		} else {
			if len(args) > 0 {
				log.Infof("received message %+v, %+v", msg, args[0])
			} else {
				log.Infof("received message %+v", msg)
			}

		}
	})

	resolver := static.NewConfigBasedResolver()

	// b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("game"), stickiness.WithResolver(resolver))

	c := scluster.NewCluster(resolver)
	c.Init()

	// game server role
	// for gate server connect.
	c.BuildServer("game", ":7879", core.TCPServBuilder, scluster.WithServerRelay(false))

	// // player client role
	// // connect gate server.
	// c.BuildClient("gate", "game", scluster.WithBalancer(b))

	c.Start()

	time.Sleep(time.Second * 600)
}
