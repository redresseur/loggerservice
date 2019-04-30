package client

import (
	"context"
	"errors"
	"github.com/golang/protobuf/proto"
	clientV1 "github.com/redresseur/loggerservice/client/v1"
	"github.com/redresseur/loggerservice/protos/protocol"
	"github.com/redresseur/loggerservice/protos/v1"
	"google.golang.org/grpc"
	"io"
)

var (
	sdkCtx, sdkCancle = context.WithCancel(context.Background())
	protocolMatched float32
	loggerServerAddr string
	operators map[float32]func(string)(io.Writer, error)
)

const ClientProtocol  = 1.0

func init()  {
	operatorV1 := func(module string)(io.Writer, error) {
		var res io.Writer
		if connect, err := grpc.DialContext(sdkCtx, loggerServerAddr); err != nil{
			return nil, err
		}else {
			ctx , _ := context.WithCancel(sdkCtx)
			cc := v1.NewLoggerV1Client(connect)

			clientInfo := v1.ClientInfo{
				Version: ClientProtocol,
				ClientId:module,
			}

			if rsp, err := cc.Registry(sdkCtx, &clientInfo); err != nil{
				return nil, err
			}else {
				if rsp.Status != 200{
					return nil, errors.New(string(rsp.Payload))
				}

				registryRsp := v1.RegistryRespond{}
				proto.Unmarshal(rsp.Payload, &registryRsp)

				ctx = context.WithValue(ctx, "LoggerUUID" ,registryRsp.LoggerId)

			}


			res = clientV1.NewLogger(cc, ctx)
		}
		return res, nil
	}

	operators[ClientProtocol] = operatorV1
}

// 版本协商
func consensus(cc protocol.ProtocolClient) float32 {
	rqs := protocol.ProtocolRequest{}
	rqs.SupportProtocol = append(rqs.SupportProtocol, ClientProtocol)
	if rsp, err := cc.FetchProtocolInfo(sdkCtx, &rqs); err != nil{
		panic(err)
		return -1
	}else {
		for _, v := range rsp.SupportProtocol{
			if v == ClientProtocol{
				return ClientProtocol
			}
		}
	}

	return -1
}

var (
	ErrorNoMatchProtocol = errors.New("don't found matched protocol")
)


func InitSDK(grpcAddr string) error {
	//Setup1 协商Protocol
	if connect, err := grpc.DialContext(sdkCtx, grpcAddr); err != nil{
		return err
	}else {
		defer connect.Close()
		cc := protocol.NewProtocolClient(connect)
		if protocolMatched = consensus(cc); protocolMatched < 0{
			return ErrorNoMatchProtocol
		}
	}

	loggerServerAddr = grpcAddr
	return nil
}

type Accident func(error)

func OpenLogger(module string) (io.Writer, error){
	if op, isOk := operators[protocolMatched]; isOk{
		return op(module)
	}

	return nil, ErrorNoMatchProtocol
}

func OpenAccident(writer io .Writer) func(error)(error) {
	logger, isOK := writer.(*clientV1.LoggerIOV1)
	if !isOK{
		return nil
	}

	return func(e error) error {
		return logger.Accident(e)
	}

}

func ReleaseSDK()  {
	sdkCancle()
}