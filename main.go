package main

import (
	"net"

	"github.com/redresseur/loggerservice/impl"
	v1 "github.com/redresseur/loggerservice/protos/v1"
	"google.golang.org/grpc"
)

func main() {
	loggerSrv := grpc.NewServer()
	v1.RegisterLoggerServer(loggerSrv, &impl.LoggerServerImpl{})
	unixAddr, err := net.ResolveUnixAddr("unix", "/tmp/logger.sock")
	if err != nil {
		panic(err)
	}

	listener, err := net.ListenUnix("unix", unixAddr)
	if err != nil {
		panic(err)
	}

	loggerSrv.Serve(listener)
}
