thriftudp 
===

[中文说明](https://gitdig.com/go-udp-thrift-server/)

Thrift is one of the best `IDL` language for the `UDP` protocol, for its grammar support `oneway` keyword. But the official package only support `TCP` transport & server framework.

This package's purpose is to extend the `thrift` with `UDP` support. It provides the following features:

- udp transport
- udp server

For the project [github.com/jaegertracing/jaeger](https://github.com/jaegertracing/jaeger) has a very beautiful implemention in an *OLD* thrift version. This project just use its `transport` implemention with tiny modification for the *NEW* version support. 

## Quick Start

please make sure [apache/thrift](https://github.com/apache/thrift) have already been installed.

### IDL definition

````thrift
namespace go echo

struct Request {
    1: string message;
}

service Echo {
    oneway void Ping(1: Request request);
}

````

use `thrift` generate the framework code:

````bash
$: cd example
$: thrift -out thrift -r --gen go idl/echo.thrift 
````

### Server Implemention

````go

package main

import (
	"context"
	"log"

	"github.com/x-mod/routine"
	"github.com/x-mod/thriftudp"
	"github.com/x-mod/thriftudp/example/thrift/echo"
)

type echoImpl struct{}

func (srv *echoImpl) Ping(ctx context.Context, request *echo.Request) (err error) {
	log.Println("echo Ping Request Message: ", request.Message)
	return nil
}

func main() {
	routine.Main(context.TODO(), routine.ExecutorFunc(func(ctx context.Context) error {
		srv := thriftudp.NewServer(
			thriftudp.ListenAddress("127.0.0.1:8888"),
			thriftudp.Processor(
				echo.NewEchoProcessor(&echoImpl{}),
				2),
		)
		if err := srv.Open(ctx); err != nil {
			return err
		}
		log.Println("serving at 127.0.0.1:8888 ...")
		return srv.Serv(ctx)
	}))
}

````

### Client Implemention

````go
package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/x-mod/thriftudp/example/thrift/echo"
	"github.com/x-mod/thriftudp/transport"
)

func main() {
	tr, err := transport.NewTUDPClientTransport("127.0.0.1:8888", "")
	if err != nil {
		log.Println(err)
		return
	}
	client := echo.NewEchoClientFactory(tr, thrift.NewTCompactProtocolFactory())

	req := echo.NewRequest()
	req.Message = strings.Join(os.Args[1:], " ")

	if err := client.Ping(context.TODO(), req); err != nil {
		log.Println(err)
		return
	}
}

````

More details, Please check the `example`.