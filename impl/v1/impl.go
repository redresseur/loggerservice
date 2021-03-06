package v1

import (
	"context"
	"fmt"
	"github.com/redresseur/loggerservice/common"
	"github.com/redresseur/loggerservice/errors"
	"os"
	"path/filepath"
	"sync"

	v1 "github.com/redresseur/loggerservice/protos/v1"
	"github.com/redresseur/loggerservice/utils/ioutils"
)

type LoggerServerImplV1 struct {
	files    sync.Map // the lg files io sets
	config   *LoggerSerivceConfV1
	tempPath string
}

func NewLoggerServerImplV1(conf *LoggerSerivceConfV1) (v1.LoggerV1Server, error) {
	// create the rootdir
	if _, err := ioutils.CreateDirIfMissing(conf.RootDir); err != nil {
		return nil, err
	}

	return &LoggerServerImplV1{
		config: conf,
	}, nil
}

// loggerPath 为client生成存储的日志文件的地址
func (ls *LoggerServerImplV1) loggerPath(path string) (string, error) {
	path = filepath.Join(ls.config.RootDir, path)
	if _, err := ioutils.CreateDirIfMissing(path); err != nil {
		return "", err
	}
	return path, nil
}

func (ls *LoggerServerImplV1) Registry(ctx context.Context, req *v1.ClientInfo) (*v1.Respond, error) {
	// TODO :增加版本判断
	if "" == req.GetClientId() {
		return ErrorRespond(400, errors.ErrorIDIsEmpty), nil
	}

	path, err := ls.loggerPath(req.GetClientId())
	if err != nil {
		return ErrorRespond(500, err), nil
	}

	fileId := common.NewLogFileID(req.GetClientId())
	path = filepath.Join(path, fileId)

	fd, err := os.Create(path)
	if err != nil {
		return ErrorRespond(500, err), nil
	}

	loggerId := common.NewClientUUID()
	logger := LoggerV1{
		path:     path,
		clientId: req.GetClientId(),
		id:       loggerId,
		fileId:   fileId,
		FileMutexIO: ioutils.NewFileMutexIO(false),
	}

	logger.Set(fd)
	ls.files.Store(loggerId, &logger)

	return RegistryRespond(loggerId), nil
}

func (ls *LoggerServerImplV1) Commit(ctx context.Context, message *v1.Message) (*v1.Respond, error) {
	// TODO :增加版本判断
	var logger *LoggerV1
	if v, isOK := ls.files.Load(message.GetLoggerId()); !isOK {
		return ErrorRespond(400, errors.ErrorLoggerIDIsNotValid), nil
	} else {
		logger = v.(*LoggerV1)
	}

	//message.LoggerId = ctx.Value("LoggerUUID").(string)
	if err := logger.write(message, logger); err != nil {
		return ErrorRespond(400, err), nil
	}

	// TODO: 重大Accident 通知警告
	if message.GetTag() == v1.LogMessageTag_ACCIDENT {
		fmt.Printf("ACCIDNET %s", string(message.GetMessage()))
	}

	return &v1.Respond{
		Status:  200,
		Version: ProtocolVersion,
	}, nil
}
