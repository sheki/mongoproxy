package mongoproxy

import (
	"net"
	"sync"
	"time"

	"github.com/golang/glog"
)

// Listener listen's on a particular port
type Listener struct {

	// Addr is the address to listen on e.g. "0.0.0.0:6666"
	Addr string

	Timeout time.Duration

	listener   net.Listener
	dispatcher *Dispatcher
	wg         sync.WaitGroup
	isRunning  atomicBool
	metrics    *listenerMetrics
}

type atomicBool struct {
	value bool
	mutex sync.RWMutex
}

func (a *atomicBool) Get() bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.value
}

func (a *atomicBool) Set(b bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.value = b
}

// Stop stops the listener
func (l *Listener) Stop() {
	l.isRunning.Set(false)
	if err := l.listener.Close(); err != nil {
		glog.Error("error closing listener", err)
	}
	block := func() (interface{}, error) {
		l.wg.Wait()
		return nil, nil
	}
	TimeoutIn(block, l.Timeout)
}

// Start your listener, not thread safe, blocking.
func (l *Listener) Start() error {

	var err error
	l.listener, err = net.Listen("tcp", l.Addr)
	if err != nil {
		return err
	}
	l.isRunning.Set(true)
	for {
		if l.isRunning.Get() {
			l.wg.Add(1)
			conn, err := l.listener.Accept()
			if err != nil {
				l.wg.Done()
				glog.Error("error accepting connection ", err)
				continue
			}
			l.metrics.acceptCounter.Mark(1)
			go l.readLoop(conn)
		}
	}
}

func (l *Listener) readLoop(conn net.Conn) {
	defer l.wg.Done()
	defer l.metrics.connectionDrop.Mark(1)
	m := NewMongoConn(conn)
	// keep reading till the connection is alive.
	for {
		h, err := m.ReadHeader()
		if err != nil {
			glog.Error("failed to read header from incoming conn", err)
			m.Close()
			return
		}
		msg := newDispatchMessage(m, h)
		if err := l.dispatcher.Dispatch(msg, l.Timeout); err != nil {
			glog.Error("dispatch timed out", err)
			m.Close()
			return
		}
		if err := msg.Wait(l.Timeout); err != nil {
			glog.Error("timed out waiting for dispatch response")
			m.Close()
			return
		}
	}
}
