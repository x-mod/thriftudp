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
