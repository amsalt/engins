package test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/amsalt/engins"
	"github.com/amsalt/engins/cluster"
	"github.com/amsalt/ngicluster/balancer"
	"github.com/amsalt/ngicluster/balancer/stickiness"
	"github.com/amsalt/ngicluster/resolver/static"
	"github.com/amsalt/nginet/core"
)

func TestGate(t *testing.T) {

	engins.RegisterMsgByID(1, &tcpChannel{})

	resolver := static.NewConfigBasedResolver()
	resolver.Register("game", ":7879") // register game server for test.

	c := cluster.NewCluster(resolver)

	// register router mapping
	// message whose ID is 1 will be relay to special `game` server.
	c.RegisterRelayRouter(1, "game")

	// gate server role
	// for player client connecting.
	c.BuildServer("gate", ":7878", core.TCPServBuilder, cluster.WithServerRelay(true))

	// connect game server.
	b := balancer.GetBuilder("stickiness").Build(stickiness.WithServName("game"), stickiness.WithResolver(resolver))
	c.BuildClient("game", "gate", cluster.WithBalancer(b))

	fmt.Println(rand.Intn(19))
	engins.Run(c)
}
