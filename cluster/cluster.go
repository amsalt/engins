package cluster

import (
	"github.com/amsalt/engins"
	"github.com/amsalt/ngicluster"
	"github.com/amsalt/ngicluster/balancer"
	"github.com/amsalt/ngicluster/consts"
	"github.com/amsalt/ngicluster/resolver"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/encoding"
	"github.com/amsalt/nginet/encoding/json"
)

// package cluster provides the API for building an auto service discovery cluster.
// it's easy to build a client auto connect to the special kinds of servers,
// and it's also easy to build a server for relay message.
// By assigning balancer, a message router strategy can work, and the default
// balancer `stickiness` will record the router path and router message to the
// right server or client.

// IdentifySelf is a protocol for cluster client to register self information when connected with server.
type IdentifySelf struct {
	Name string
	Addr string
}

type Cluster struct {
	resolver resolver.Resolver
	clus     *ngicluster.Cluster
	servers  map[*ngicluster.Server]string
	storages map[string]balancer.Storage
}

func NewCluster(rsv resolver.Resolver) *Cluster {
	c := &Cluster{resolver: rsv}
	c.clus = ngicluster.NewCluster(rsv)
	c.servers = make(map[*ngicluster.Server]string)
	c.storages = make(map[string]balancer.Storage)
	c.Init()

	return c
}

// RegisterRelayRouter registers the relay router mapping.
// when registers message with msgID to service name `servName`,message
// with the msgID will be sent to the server with Name `servName` automatically
func (c *Cluster) RegisterRelayRouter(msgID interface{}, servName string) {
	c.clus.Register(msgID, servName)
}

func (c *Cluster) Clients(servName string) []core.SubChannel {
	return c.clus.Clients(servName)
}

// Write sends message to server with name `servName`.
// it will try to use client connection or server connection to send message.
func (c *Cluster) Write(servName string, msg interface{}, ctx ...interface{}) error {
	var err error
	if len(c.Clients(servName)) > 0 {
		err = c.clus.Write(servName, msg, ctx...)
	} else {
		for s := range c.servers { // support multiple servers.
			if len(ctx) > 0 {
				err = s.Write(servName, msg, ctx[0])
			} else {
				err = s.Write(servName, msg, nil)
			}

			if err == nil {
				return nil
			}
		}
	}

	return err
}

func (c *Cluster) Init() {
	engins.RegisterMsgByID(
		consts.SystemIdentifySelf,
		&IdentifySelf{}).SetCodec(encoding.MustGetCodec(json.CodecJSON))
	engins.RegisterProcessorByID(consts.SystemIdentifySelf, c.identityClientHandler)
}

// Start starts the Cluster
func (c *Cluster) Start() {
	for s, addr := range c.servers {
		s.Listen(addr)
		go s.Accept()
	}
}

// Stop stops the Cluster
func (c *Cluster) Stop() {
	for s := range c.servers {
		s.Close()
	}
}
