package scluster

import (
	"github.com/amsalt/cluster"
	"github.com/amsalt/cluster/balancer"
	"github.com/amsalt/engins"
	"github.com/amsalt/nginet/core"
)

const (
	ChannelName   = "ChannelName"
	RelatedServer = "RelatedServer"
	RelatedClient = "RelatedClient"

	DefaultRelayStickiness = "UserID"
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

	server.OnConnect(func(ctx *core.ChannelContext, channel core.Channel) {
		ctx.Attr().SetValue(RelatedServer, server)

		if opts.OnConnect != nil {
			server.OnConnect(opts.OnConnect)
		}
	})

	server.OnDisconnect(func(*core.ChannelContext) {
		if opts.OnConnect != nil {
			server.OnDisconnect(opts.OnDisconnect)
		}
	})

	if opts.IsRelay {
		relayHandler := cluster.NewRelayHandler(servName, c.clus, DefaultRelayStickiness)
		server.AddAfterHandler("IDParser", nil, "RelayHandler", relayHandler)
	}

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

	c.servers[server] = addr
}

// BuildClient builds a client with Options and service type.
func (c *Cluster) BuildClient(servName string, clientName string, opt ...BuildOption) *cluster.Client {
	opts := defaultConfigOpts
	for _, o := range opt {
		o(&opts)
	}
	client := cluster.NewClientWithBufSize(opts.ReadBufSize, opts.WriteBufSize)
	client.InitConnector(opts.Executor, engins.Register, engins.Dispatcher)

	client.OnConnect(func(ctx *core.ChannelContext, channel core.Channel) {
		ctx.Attr().SetValue(RelatedClient, client)
		c.identifingSelf(clientName, ctx)

		if opts.OnConnect != nil {
			client.OnConnect(opts.OnConnect)
		}
	})

	client.OnDisconnect(func(*core.ChannelContext) {
		if opts.OnConnect != nil {
			client.OnDisconnect(opts.OnDisconnect)
		}
	})

	c.clus.AddClient(servName, client, opts.Balancer)
	return client
}

func (c *Cluster) identifingSelf(servName string, ctx *core.ChannelContext) {
	ctx.Write(&IdentifySelf{Name: servName})
}

// BuildClientWithConnector builds a client with Connector and service type.
// NOTE: When use this method, register and dispatcher must use  engins.Register and engins.Dispatcher
func (c *Cluster) BuildClientWithConnector(servName string, connector core.ConnectorChannel, opt ...BuildOption) {
	opts := defaultConfigOpts
	for _, o := range opt {
		o(&opts)
	}
	client := cluster.NewClientWithBufSize(opts.ReadBufSize, opts.WriteBufSize)
	client.SetConnector(connector)
	c.clus.AddClient(servName, client, opts.Balancer)
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

	// server specific
	MaxConn int  // limit the max connection number to the server.
	IsRelay bool // whether the server is a relay server.
}

var defaultConfigOpts = ConfigOpts{
	WriteBufSize: 1024 * 10,
	ReadBufSize:  1024 * 10,
	MaxConn:      1000000,
}
