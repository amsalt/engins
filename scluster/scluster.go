package scluster

import (
	"github.com/amsalt/cluster"
	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/cluster/balancer/stickiness"
	"github.com/amsalt/cluster/consts"
	"github.com/amsalt/cluster/resolver"
	"github.com/amsalt/engins"
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/encoding"
	"github.com/amsalt/nginet/encoding/json"
)

// IdentifySelf is a protocol for cluster client to register self information when connected with server.
type IdentifySelf struct {
	Name string
	Addr string
}

type Cluster struct {
	resolver resolver.Resolver
	clus     *cluster.Cluster
	servers  map[*cluster.Server]string
	storages map[string]balancer.Storage
}

func NewCluster(rsv resolver.Resolver) *Cluster {
	c := &Cluster{resolver: rsv}
	c.clus = cluster.NewCluster(rsv)
	c.servers = make(map[*cluster.Server]string)
	c.storages = make(map[string]balancer.Storage)

	return c
}

func (c *Cluster) RegisterRouter(msgID interface{}, servName string) {
	c.clus.Register(msgID, servName)
}

func (c *Cluster) Clients(servName string) []core.SubChannel {
	return c.clus.Clients(servName)
}

func (c *Cluster) Write(servName string, msg interface{}, ctx ...interface{}) error {
	err := c.clus.Write(servName, msg, ctx...)

	if err != nil {
		for s := range c.servers {
			log.Errorf("server %+v writing message to %+v", s, servName)
			if len(ctx) > 0 {
				err = s.Write(servName, msg, ctx[0])
			} else {
				s.Write(servName, msg, nil)
			}

			if err == nil {
				return nil
			}

		}
	}
	return err
}

func (c *Cluster) Init() {
	engins.RegisterMsgByID(consts.SystemProtocolIdentifySelf, &IdentifySelf{}).SetCodec(encoding.MustGetCodec(json.CodecJSON))
	engins.RegisterProcessorByID(consts.SystemProtocolIdentifySelf, func(ctx *core.ChannelContext, msg interface{}, args ...interface{}) {
		relatedServer := ctx.Attr().Value(RelatedServer)
		log.Debugf("%+v receive msg from client %+v", relatedServer, msg)
		if relatedServer != nil {
			if server, ok := relatedServer.(*cluster.Server); ok {
				if identify, ok := msg.(*IdentifySelf); ok {
					ctx.Attr().SetValue(ChannelName, identify.Name)
					if server.GetBalancer(identify.Name) == nil {
						c.storages[identify.Name] = stickiness.NewDefaultStorage()
						b := balancer.GetBuilder("stickiness").Build(stickiness.WithStorage(c.storages[identify.Name]), stickiness.WithServName(identify.Name), stickiness.WithResolver(c.resolver))
						server.SetBalancer(identify.Name, b)
						c.resolver.RegisterSubChannel(identify.Name, ctx.Channel().(core.SubChannel))
						log.Debugf("server set balancer for client: %+v", identify.Name)
					}
				}
			}
		}
	})
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
