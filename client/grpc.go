package client

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	clientV1 "github.com/redresseur/loggerservice/client/v1"
	_errors "github.com/redresseur/loggerservice/errors"
	"github.com/redresseur/loggerservice/protos/protocol"
	"github.com/redresseur/loggerservice/protos/v1"
	"google.golang.org/grpc"
	"io"
	"sync/atomic"
	"time"
)

var (
	gGrpcProtocolMatched   float32
	gLoggerServerAddr  []string

	gGrpcConnection *grpc.ClientConn = nil // 共用同一条连接
	gGrpcConnIsUseFul = &atomic.Value{}
	// gGrpcConnEorrorNotify chan error
	gGrpcHeartTime = 500*time.Millisecond
)

func openGrpcConn() (err error){
	var connect *grpc.ClientConn

	if gGrpcConnection != nil && gGrpcConnIsUseFul.Load().(bool){
		return nil
	}else {
		// 清除不可用链接
		if gGrpcConnection != nil{
			gGrpcConnection.Close()
		}

		gGrpcConnection = nil
	}

	gGrpcConnIsUseFul.Store(false)
	for _, addr := range gLoggerServerAddr{
		if connect, err = grpc.DialContext(gSdkCtx, addr, grpc.WithInsecure()); err != nil {
			continue
		} else {
			// 先进行协议磋商
			cc := protocol.NewProtocolClient(connect)
			if gGrpcProtocolMatched, _ = consensus(cc); gGrpcProtocolMatched < 0 {
				err = _errors.ErrorNoMatchProtocol
				continue
			}else {
				err = nil // 清空错误信息
				gGrpcConnection = connect
				gGrpcConnIsUseFul.Store(true)
				break
			}
		}
	}

	return
}


// 开启一个协程，保证gGrpcConnection链接
func initGrpc()(err error){
	if err = openGrpcConn(); err != nil{
		return
	}

	// gGrpcConnEorrorNotify = make(chan error, 16)
	// 开启一个监听线程
	// 通过心跳机制判断对端服务的状态
	go func() {
		counter := int64(0)
		ctx, _ := context.WithCancel(gSdkCtx)

		// 每隔心跳时间Ping一次
		// 如果打不开重复打开
		// 维持一条长链接
		for true {
			timer := time.After(gGrpcHeartTime)
			select {
			case <-ctx.Done():
				return
			case <-timer:
				if err = openGrpcConn(); err != nil{
					//fmt.Printf("打开连接失败了")
					continue
				}

				pp := protocol.NewPingPongClient(gGrpcConnection)
				pong, err := pp.PingIng(ctx, &protocol.Ping{Counter:counter})
				if err != nil || pong.Counter != counter{
					gGrpcConnIsUseFul.Store(false)
					// 此时说明不可用了, 计数清零
					counter = 0
				}

				counter++
			}
		}

	}()

	return
}

// 修改后可以支持同时打开多个模块
func operatorGrpc(module string) (res io.Writer, err error) {
	// 多个服务器地址的情况下
	if err = openGrpcConn(); err != nil {
		return nil, err
	}

	ctx, _ := context.WithCancel(gSdkCtx)
	cc := v1.NewLoggerV1Client(gGrpcConnection)

	clientInfo := v1.ClientInfo{
		Version:  ClientGrpcProtocol,
		ClientId: module,
	}

	if rsp, err := cc.Registry(gSdkCtx, &clientInfo); err != nil {
		return nil, err
	} else {
		if rsp.Status != 200 {
			err = errors.New(string(rsp.Payload))
			return nil, err
		}

		registryRsp := v1.RegistryRespond{}
		proto.Unmarshal(rsp.Payload, &registryRsp)

		ctx = context.WithValue(ctx, "LoggerUUID", registryRsp.LoggerId)
	}

	res = clientV1.NewLogger(cc, ctx)
	return res, nil
}


// 版本协商
func consensus(cc protocol.ProtocolClient) (float32, error) {
	rqs := protocol.ProtocolRequest{}
	rqs.SupportProtocol = append(rqs.SupportProtocol, ClientGrpcProtocol)
	if rsp, err := cc.FetchProtocolInfo(gSdkCtx, &rqs); err != nil {
		//panic(err)
		return -1, err
	} else {
		for _, v := range rsp.SupportProtocol {
			if v == ClientGrpcProtocol {
				return ClientGrpcProtocol, nil
			}
		}
	}

	return -1, _errors.ErrorNoMatchProtocol
}

func releaseGrpc()  {
	if gGrpcConnection != nil{
		gGrpcConnection.Close()
	}

	gGrpcConnection = nil
}