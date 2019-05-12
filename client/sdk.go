package client

import (
	"context"
	clientV1 "github.com/redresseur/loggerservice/client/v1"
	_errors "github.com/redresseur/loggerservice/errors"
	"io"
	"os"
	"time"
)

type LoggerType int32
const (
	// 发送到远端服务器
	LoggerGrpc LoggerType = 1 << iota
	// 本地直接写文件
	LoggerLocal
	// 直接打印到控制台
	LoggerStd
	// 优先缓存到内存，
	// 再根据初始化时设置的
	// 类型选择同步到本地或远端
	// 注： 只是为了说明，没有实际意义
	LoggerChannelIO
	// 同时保存在本地 和远端服务器，
	// 优先写入，再同步到远端服务器
	// 注： 只是为了说明，没有实际意义
	LoggerProductionIO
)

const (
	ClientGrpcProtocol = 1.0
)

var (
	gSdkCtx, gSdkCancel = context.WithCancel(context.Background())
	gOperators   = map[LoggerType]func(string) (io.Writer, error){}
	gLoggerType  = LoggerLocal
)

func init() {
	if gLocalRootDir = os.Getenv("TEMP"); gLocalRootDir == ""{
		gLocalRootDir = "/tmp"
	}

	gOperators[LoggerGrpc] = operatorGrpc
	gOperators[LoggerLocal] = operatorLocal
	gOperators[LoggerStd] = func(s string) (io.Writer, error) {
		return os.Stdout, nil
	}
}

type SdkOption func()

func WithLoggerServerAddr(grpcAddr string) SdkOption {
	return func() {
		gLoggerServerAddr = append(gLoggerServerAddr, grpcAddr)
	}
}

func WithLocalRootDir(rootDir string)SdkOption  {
	return func() {
		gLocalRootDir = rootDir
	}
}

func WithLoggerType(loggerType2 LoggerType)SdkOption  {
	return func() {
		gLoggerType = loggerType2
	}
}

func WithGrpcHeartTime(time time.Duration) SdkOption {
	return func() {
		gGrpcHeartTime = time
	}
}

func InitSDK(options... SdkOption) error {
	for _, op := range options{
		op()
	}

	if len(gLoggerServerAddr) != 0{
		return initGrpc()
	}

	return nil
}


func OpenChannelIoLogger(module string) (io.Writer, error){
	if w, err := openLogger(module); err != nil{
		return nil, err
	}else {
		return NewChannelIO(w, module), nil
	}
}

func CloseLogger(writer io.Writer)  {
	if ch, isOK := writer.(*channelIO); isOK {
		ch.Close()
	}
}

func openLogger(module string) (io.Writer, error) {
	if op, isOk := gOperators[gLoggerType]; isOk {
		return op(module)
	}

	return nil, _errors.ErrorNoMatchProtocol
}

type Accident func(error)error
// OpenAccident 目前只有grpc 类型的Logger 支持此功能
func OpenAccident(writer io.Writer) (Accident, error) {

	if logger, isOK := writer.(*clientV1.LoggerIOV1); isOK {
		return func(e error) error {
			return logger.Accident(e)
		}, nil
	}

	if logger, isOK := writer.(*channelIO); isOK {
		if grpcLogger, isOK := logger.realWriter.(*clientV1.LoggerIOV1); isOK {
			return func(e error) error {
				return grpcLogger.Accident(e)
			}, nil
		}
	}

	return nil, _errors.ErrorFunctionNotSupported
}

func ReleaseSDK() {
	// 一定要先结束生命周期
	gSdkCancel()

	releaseGrpc()
	// 然后结束IO
	//gChannelIO.Close()

}
