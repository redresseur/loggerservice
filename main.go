package main

import (
	"github.com/redresseur/loggerservice/utils/ioutils"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"

	implProtocol "github.com/redresseur/loggerservice/impl/protocol"
	implV1 "github.com/redresseur/loggerservice/impl/v1"
	"github.com/redresseur/loggerservice/protos/protocol"
	v1 "github.com/redresseur/loggerservice/protos/v1"
	"google.golang.org/grpc"
)

// startHttpService 开启一个http服务，用于浏览日志
func startHTTPService(httpServerAddr, rootPath string) {
	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
		//http.Handle("/logs", http.StripPrefix("/logs", http.FileServer(http.Dir(rootPath))))
		http.Handle("/", http.FileServer(http.Dir(rootPath)))
		http.ListenAndServe(httpServerAddr, nil)
	}()
}

func getConfig() (*implV1.LoggerSerivceConfV1, error) {
	// default config
	// defaultConf := implV1.LoggerSerivceConfV1{
	// 	GrpcServerAddr: "/tmp/logger.sock",
	// 	HttpServerAddr: ":10030",
	// 	RootDir:        "/tmp/logger",
	// 	NetWork:        "unix",
	// }

	defaultConf := implV1.LoggerSerivceConfV1{
		GrpcServerAddr: ":10041",
		HttpServerAddr: ":10030",
		RootDir:        filepath.Join(ioutils.TempDir(), "logger"),
		NetWork:        "tcp",
	}

	return &defaultConf, nil
}

func main() {
	loggerSrv := grpc.NewServer()

	protocolHandler := implProtocol.ProtocolServerImpl{}
	implProtocol.RegistryProtocol(&protocolHandler, implV1.ProtocolVersion)
	protocol.RegisterProtocolServer(loggerSrv, &protocolHandler)

	pingPongHandler := implProtocol.PingPongImpl{}
	protocol.RegisterPingPongServer(loggerSrv, &pingPongHandler)

	conf, err := getConfig()
	if err != nil {
		panic(err)
	}

	loggerV1Handler, err := implV1.NewLoggerServerImplV1(conf)
	if err != nil {
		panic(err)
	}

	v1.RegisterLoggerV1Server(loggerSrv, loggerV1Handler)

	var listener net.Listener
	switch conf.NetWork {
	case "unix":
		os.Remove(conf.GrpcServerAddr)
		unixAddr, err := net.ResolveUnixAddr("unix", conf.GrpcServerAddr) //"/tmp/logger.sock"
		if err != nil {
			panic(err)
		}

		listener, err = net.ListenUnix("unix", unixAddr)
		if err != nil {
			panic(err)
		}
	case "tcp":
		fallthrough
	case "tcp6":
		listener, err = net.Listen(conf.NetWork, conf.GrpcServerAddr)
		if err != nil {
			panic(err)
		}
	}

	startHTTPService(conf.HttpServerAddr, conf.RootDir)
	loggerSrv.Serve(listener)
}
