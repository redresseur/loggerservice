package main

import (
	implProtocol "github.com/redresseur/loggerservice/impl/protocol"
	"github.com/redresseur/loggerservice/protos/protocol"
	"github.com/redresseur/loggerservice/protos/v1"
	"net"

	implV1 "github.com/redresseur/loggerservice/impl/v1"
	"google.golang.org/grpc"
)

func main() {
	loggerSrv := grpc.NewServer()

	protocolHandler := implProtocol.ProtocolServerImpl{}
	protocol.RegisterLoggerServer(loggerSrv, &protocolHandler)

	loggerV1Handler := implV1.LoggerServerImplV1{}
	v1.RegisterLoggerServer(loggerSrv, &loggerV1Handler)

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
