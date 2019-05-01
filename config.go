package thriftudp

import (
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/uber/jaeger-lib/metrics"
)

type serverConfig struct {
	address       string
	metrics       metrics.Factory
	processor     thrift.TProcessor
	protosIn      thrift.TProtocolFactory
	protosOut     thrift.TProtocolFactory
	concurrent    int
	maxQueueSize  int
	maxPacketSize int
}

//ServerOpt option for server
type ServerOpt func(*serverConfig)

//ListenAddress ServerOpt | Required
func ListenAddress(addr string) ServerOpt {
	return func(cf *serverConfig) {
		cf.address = addr
	}
}

//Processor ServerOpt | Required
func Processor(exec thrift.TProcessor, concurrent int) ServerOpt {
	return func(cf *serverConfig) {
		cf.processor = exec
		if concurrent != 0 {
			cf.concurrent = concurrent
		}
	}
}

//MetricsFactory ServerOpt | Optional
func MetricsFactory(factory metrics.Factory) ServerOpt {
	return func(cf *serverConfig) {
		cf.metrics = factory
	}
}

//ProtocolFactory ServerOpt | Optional
func ProtocolFactory(in, out thrift.TProtocolFactory) ServerOpt {
	return func(cf *serverConfig) {
		cf.protosIn = in
		cf.protosOut = out
	}
}

//MaxQueueSize ServerOpt | Optional
func MaxQueueSize(sz int) ServerOpt {
	return func(cf *serverConfig) {
		cf.maxQueueSize = sz
	}
}

//MaxPacketSize ServerOpt | Optional
func MaxPacketSize(sz int) ServerOpt {
	return func(cf *serverConfig) {
		cf.maxPacketSize = sz
	}
}
