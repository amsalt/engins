package test

import (
	"testing"
	"time"

	"github.com/amsalt/engins"
	"github.com/amsalt/engins/cluster"
	"github.com/amsalt/log"
	"github.com/amsalt/ngicluster/resolver/static"
	"github.com/amsalt/nginet/core"
)

func TestGame(t *testing.T) {
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

	// game server role
	// for gate server connect.
	c := cluster.NewCluster(resolver)
	c.Init()
	c.BuildServer("game", ":7879", core.TCPServBuilder, cluster.WithServerRelay(false))
	c.Start()

	time.Sleep(time.Second * 300)
}
