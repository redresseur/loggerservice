package main

import (
	implProtocol "github.com/redresseur/loggerservice/impl/protocol"
	implV1 "github.com/redresseur/loggerservice/impl/v1"
	"github.com/redresseur/loggerservice/protos/protocol"
	"github.com/redresseur/loggerservice/protos/v1"
	"google.golang.org/grpc"
	"net"
	"net/http"
)

// 开启一个http服务，用于浏览日志
func startHttpService(path string)  {
	go func() {
		http.Handle("/logs", http.StripPrefix("/logs", http.FileServer(http.Dir(path))))
		http.ListenAndServe(":10030", nil)
	}()
}

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
