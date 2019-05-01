package thriftudp

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/metricstest"
	"github.com/x-mod/thriftudp/transport"
)

//Server struct
type Server struct {
	config        *serverConfig
	transport     thrift.TTransport
	dataChan      chan *ReadBuf
	dataPool      *sync.Pool
	protosInPool  *sync.Pool
	protosOutPool *sync.Pool
	opened        bool
	queueSize     int64
	mu            sync.Mutex
	wgroup        sync.WaitGroup
	metrics       struct {
		// Size of the current server queue
		QueueSize metrics.Gauge `metric:"thrift.udp.server.queue_size"`

		// Size (in bytes) of packets received by server
		PacketSize metrics.Gauge `metric:"thrift.udp.server.packet_size"`

		// Number of packets dropped by server
		PacketsDropped metrics.Counter `metric:"thrift.udp.server.packets.dropped"`

		// Number of packets processed by server
		PacketsProcessed metrics.Counter `metric:"thrift.udp.server.packets.processed"`

		// Number of malformed packets the server received
		ReadError metrics.Counter `metric:"thrift.udp.server.read.errors"`
		// Number of processor handle errors
		ProcessorHandleError metrics.Counter `metric:"thrift.udp.server.processor.handle.errors"`
	}
}

//NewServer new udp server
func NewServer(opts ...ServerOpt) *Server {
	//default config
	config := &serverConfig{
		metrics:       metricstest.NewFactory(time.Second),
		protosIn:      thrift.NewTCompactProtocolFactory(),
		protosOut:     thrift.NewTCompactProtocolFactory(),
		concurrent:    runtime.NumCPU(),
		maxPacketSize: 65536,
		maxQueueSize:  2048,
	}
	//option config
	for _, opt := range opts {
		opt(config)
	}
	return &Server{
		config: config,
	}
}

//Open server
func (srv *Server) Open(ctx context.Context) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if !srv.opened {
		if srv.config.processor == nil {
			return errors.New("processor required")
		}
		if err := metrics.Init(&srv.metrics, srv.config.metrics, nil); err != nil {
			return err
		}
		udps, err := transport.NewTUDPServerTransport(srv.config.address)
		if err != nil {
			return err
		}
		srv.transport = udps
		srv.dataChan = make(chan *ReadBuf, srv.config.maxQueueSize)
		srv.dataPool = &sync.Pool{
			New: func() interface{} {
				return &ReadBuf{bytes: make([]byte, srv.config.maxPacketSize)}
			},
		}
		srv.protosInPool = &sync.Pool{
			New: func() interface{} {
				trans := &transport.TBufferedReadTransport{}
				return srv.config.protosIn.GetProtocol(trans)
			},
		}
		srv.protosOutPool = &sync.Pool{
			New: func() interface{} {
				trans := &transport.TBufferedReadTransport{}
				return srv.config.protosOut.GetProtocol(trans)
			},
		}
		srv.opened = true
	}
	return nil
}

//Serv server
func (srv *Server) Serv(ctx context.Context) error {
	if err := srv.Open(ctx); err != nil {
		return err
	}
	srv.wgroup.Add(srv.config.concurrent)
	for i := 0; i < srv.config.concurrent; i++ {
		go srv.processing(ctx)
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			readBuf := srv.dataPool.Get().(*ReadBuf)
			n, err := srv.transport.Read(readBuf.bytes)
			if err == nil {
				readBuf.n = n
				srv.metrics.PacketSize.Update(int64(n))
				select {
				case srv.dataChan <- readBuf:
					srv.metrics.PacketsProcessed.Inc(1)
					srv.updateQueueSize(1)
				default:
					srv.metrics.PacketsDropped.Inc(1)
				}
			} else {
				srv.metrics.ReadError.Inc(1)
			}
		}
	}
}

func (srv *Server) updateQueueSize(delta int64) {
	atomic.AddInt64(&srv.queueSize, delta)
	srv.metrics.QueueSize.Update(atomic.LoadInt64(&srv.queueSize))
}

func (srv *Server) dataRecv(buf *ReadBuf) {
	srv.updateQueueSize(-1)
	srv.dataPool.Put(buf)
}

func (srv *Server) processing(ctx context.Context) {
	for readBuf := range srv.dataChan {
		in := srv.protosInPool.Get().(thrift.TProtocol)
		out := srv.protosOutPool.Get().(thrift.TProtocol)

		in.Transport().Write(readBuf.GetBytes())
		if _, err := srv.config.processor.Process(ctx, in, out); err != nil {
			srv.metrics.ProcessorHandleError.Inc(1)
		}
		//recycle
		srv.protosInPool.Put(in)
		srv.protosOutPool.Put(out)
		srv.dataRecv(readBuf)
	}
	srv.wgroup.Done()
}

//Close server
func (srv *Server) Close() {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.opened {
		srv.opened = false
		srv.transport.Close()
		close(srv.dataChan)
		srv.wgroup.Wait()
	}
}
