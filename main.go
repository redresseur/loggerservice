package main

import (
	"github.com/redresseur/loggerservice/protos/v1"
	"context"
	"google.golang.org/grpc"
	"net"
)

type LoggerServerImpl struct {

}

func (ls *LoggerServerImpl)Commit(context.Context, *v1.LogMessageRequest) (*v1.LogMessageReply, error)  {
	return nil, nil
}

func main()  {
	loggerSrv := grpc.NewServer()
	v1.RegisterLoggerServer(loggerSrv, &LoggerServerImpl{})
	unixAddr, err := net.ResolveUnixAddr("unix", "/tmp/logger.sock");
	if  err != nil{
		panic(err)
	}

	listener, err := net.ListenUnix("unix", unixAddr)
	if err != nil{
		panic(err)
	}

	loggerSrv.Serve(listener)
}