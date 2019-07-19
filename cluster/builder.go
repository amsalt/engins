package cluster

import (
	"github.com/amsalt/engins"
	"github.com/amsalt/log"
	"github.com/amsalt/ngicluster"
	"github.com/amsalt/ngicluster/balancer"
	"github.com/amsalt/ngicluster/balancer/stickiness"
	"github.com/amsalt/nginet/core"
)

const (
	ChannelNameKey            = "ChannelName"
	AssociatedServerKey       = "AssociatedServer"
	AssociatedClientKey       = "AssociatedClient"
	DefaultRelayStickinessKey = "UserID"
)

// BuildOption helper method to build a new client or a server.
type BuildOption func(interface{})

// BuildServer builds a new server with ServerName, address, serverType and Options.
func (c *Cluster) BuildServer(servName string, addr string, servType string, opt ...BuildOption) {
	opts := defaultConfigOpts
	for _, o := range opt {
		o(&opts)
	}
	server := c.clus.NewServerWithConfig(servName, opts.ReadBufSize, opts.WriteBufSize, opts.MaxConn)

	server.InitAcceptor(opts.Executor, engins.Register, engins.Dispatcher, servType)

	c.registerServListener(server, servName, &opts)
	c.servers[server] = addr
}

// BuildServerWithAcceptor builds a new server with serverName, address, acceptor and Options.
// NOTE: When use this method, register and dispatcher must use  engins.Register and engins.Dispatcher
func (c *Cluster) BuildServerWithAcceptor(servName string, addr string, acceptor core.AcceptorChannel, opt ...BuildOption) {
	opts := defaultConfigOpts
	for _, o := range opt {
		o(&opts)
	}
	server := c.clus.NewServerWithConfig(servName, opts.ReadBufSize, opts.WriteBufSize, opts.MaxConn)
	server.SetAcceptor(acceptor)

	c.registerServListener(server, servName, &opts)
	c.servers[server] = addr
}

func (c *Cluster) registerServListener(server *ngicluster.Server, servName string, opts *ConfigOpts) {
	server.OnConnect(func(ctx *core.ChannelContext, channel core.Channel) {
		ctx.Attr().SetValue(AssociatedServerKey, server)
		if opts.OnConnect != nil {
			opts.OnConnect(ctx, channel)
		}
	})

	server.OnDisconnect(func(ctx *core.ChannelContext) {
		if opts.OnDisconnect != nil {
			opts.OnDisconnect(ctx)
		}
	})

	if opts.IsRelay {
		relayHandler := ngicluster.NewRelayHandler(servName, c.clus, DefaultRelayStickinessKey)
		server.AddAfterHandler("IDParser", nil, "RelayHandler", relayHandler)
	}
}

// BuildClient builds a client with Options and service type.
// servName: the service name to connect.
// clientName: the name to be used to represents this client.
func (c *Cluster) BuildClient(servName string, clientName string, opt ...BuildOption) *ngicluster.Client {
	opts := defaultConfigOpts
	for _, o := range opt {
		o(&opts)
	}
	client := ngicluster.NewClientWithBufSize(opts.ReadBufSize, opts.WriteBufSize)
	client.InitConnector(opts.Executor, engins.Register, engins.Dispatcher, true)

	c.registerCliListener(client, clientName, &opts)
	c.clus.AddClient(servName, client, opts.Balancer)
	return client
}

// BuildClientWithConnector builds a client with Connector and service type.
// NOTE: When use this method, register and dispatcher must use  engins.Register and engins.Dispatcher
func (c *Cluster) BuildClientWithConnector(servName string, clientName string, connector core.ConnectorChannel, opt ...BuildOption) {
	opts := defaultConfigOpts
	for _, o := range opt {
		o(&opts)
	}
	client := ngicluster.NewClientWithBufSize(opts.ReadBufSize, opts.WriteBufSize)
	client.SetConnector(connector)

	c.registerCliListener(client, clientName, &opts)
	c.clus.AddClient(servName, client, opts.Balancer)
}

func (c *Cluster) registerCliListener(client *ngicluster.Client, clientName string, opts *ConfigOpts) {
	client.OnConnect(func(ctx *core.ChannelContext, channel core.Channel) {
		ctx.Attr().SetValue(AssociatedClientKey, client)
		c.identifingSelf(clientName, ctx)

		if opts.OnConnect != nil {
			opts.OnConnect(ctx, channel)
		}
	})

	client.OnDisconnect(func(ctx *core.ChannelContext) {
		if opts.OnDisconnect != nil {
			opts.OnDisconnect(ctx)
		}
	})
}

func (c *Cluster) identifingSelf(servName string, ctx *core.ChannelContext) {
	ctx.Write(&IdentifySelf{Name: servName})
}

// WithOnConnect register handler When Connect.
func WithOnConnect(f func(*core.ChannelContext, core.Channel)) BuildOption {
	return func(o interface{}) {
		o.(*ConfigOpts).OnConnect = f
	}
}

// WithOnDisConnect register handler When DisConnect.
func WithOnDisConnect(f func(ctx *core.ChannelContext)) BuildOption {
	return func(o interface{}) {
		o.(*ConfigOpts).OnDisconnect = f
	}
}

// WithWriteBufSize sets max size of pending wirte.
func WithWriteBufSize(s int) BuildOption {
	return func(o interface{}) {
		o.(*ConfigOpts).WriteBufSize = s
	}
}

// WithReadBufSize sets max size of pending read.
func WithReadBufSize(s int) BuildOption {
	return func(o interface{}) {
		o.(*ConfigOpts).ReadBufSize = s
	}
}

// WithExecutor sets executor.
func WithExecutor(e core.Executor) BuildOption {
	return func(o interface{}) {
		o.(*ConfigOpts).Executor = e
	}
}

// WithBalancer sets executor.
func WithBalancer(b balancer.Balancer) BuildOption {
	return func(o interface{}) {
		o.(*ConfigOpts).Balancer = b
	}
}

// WithServerMaxConnSize sets max size of connected clients.
func WithServerMaxConnSize(m int) BuildOption {
	return func(o interface{}) {
		o.(*ConfigOpts).MaxConn = m
	}
}

// WithServerRelay sets whether the server is a relay server.
func WithServerRelay(b bool) BuildOption {
	return func(o interface{}) {
		o.(*ConfigOpts).IsRelay = b
	}
}

// ConfigOpts represents the options to build a new cluster.Server or cluster.Client
type ConfigOpts struct {
	OnConnect    func(*core.ChannelContext, core.Channel)
	OnDisconnect func(ctx *core.ChannelContext)

	Executor     core.Executor     // sets which goroutine logic handler run.
	WriteBufSize int               // sets the size of write buffer.
	ReadBufSize  int               // sets the size of read buffer.
	Balancer     balancer.Balancer // sets the balancer to dispatch message in servers.

	// server specifics
	MaxConn int  // limit the max connection number to the server.
	IsRelay bool // whether the server is a relay server.
}

var defaultConfigOpts = ConfigOpts{
	WriteBufSize: 1024 * 10,
	ReadBufSize:  1024 * 10,
	MaxConn:      1000000,
}

func (c *Cluster) identityClientHandler(ctx *core.ChannelContext, msg interface{}, args ...interface{}) {
	relatedServer := ctx.Attr().Value(AssociatedServerKey)
	if relatedServer == nil {
		return
	}

	server, isServer := relatedServer.(*ngicluster.Server)
	identify, isIdentifySelf := msg.(*IdentifySelf)

	if isServer && isIdentifySelf {
		ctx.Attr().SetValue(ChannelNameKey, identify.Name)
		if server.GetBalancer(identify.Name) == nil {
			c.storages[identify.Name] = stickiness.NewDefaultStorage()
			b := balancer.GetBuilder(stickiness.Name).Build(
				stickiness.WithStorage(c.storages[identify.Name]),
				stickiness.WithServName(identify.Name),
				stickiness.WithResolver(c.resolver),
			)

			server.SetBalancer(identify.Name, b)
			c.resolver.RegisterSubChannel(identify.Name, ctx.Channel().(core.SubChannel))
			log.Debugf("server set balancer for client: %+v", identify.Name)
		}
	}
}
