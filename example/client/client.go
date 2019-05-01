package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/x-mod/thriftudp/transport"
	"github.com/x-mod/thriftudp/example/thrift/echo"
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
