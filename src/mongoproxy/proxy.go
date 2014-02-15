package mongoproxy

import (
	"errors"
	"time"

	"github.com/golang/glog"
)

// Proxy is a mongo proxy.
type Proxy struct {
	ListenAddr string
	// MongoAddr address of the mongo host to proxy
	MongoAddr string
	// MaxMongoConnections to open to a mongo node. Each mongo can support
	// 20,000 nodes at most.
	MaxMongoConnections uint
	// DispatchQueueLen is the length of the queue b/w incoming & outgoing
	// requests. The incoming requests are blocked when the queue is full.
	DispatchQueueLen uint

	// DispatcherTimeout
	DispatcherTimeout time.Duration

	// ListenerTimeout
	ListenerTimeout time.Duration

	listener *Listener
	dsp      *Dispatcher
}

// Start the mongo proxy is blocking
func (p *Proxy) Start() error {

	if p.MaxMongoConnections == 0 {
		glog.Info("setting MaxMongoConnections to 20,000")
		p.MaxMongoConnections = 20000
	}
	if p.MaxMongoConnections > 20000 {
		return errors.New("maxMongoConnections cannot be greater than 20000")
	}

	p.dsp = p.createDispatcher()
	p.dsp.Start()

	p.listener = &Listener{
		Addr:       p.ListenAddr,
		dispatcher: p.dsp,
		Timeout:    p.ListenerTimeout,
		metrics:    newListenerMetric(),
	}
	return p.listener.Start()
}

func (p *Proxy) createDispatcher() *Dispatcher {
	timeout := p.DispatcherTimeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &Dispatcher{
		ChannelLen:  p.DispatchQueueLen,
		TargetAddr:  p.MongoAddr,
		Timeout:     timeout,
		metrics:     newDispatchMetrics(),
		NumRoutines: p.MaxMongoConnections,
	}
}

// Stop stops the proxy
func (p *Proxy) Stop() {
	p.listener.Stop()
}
