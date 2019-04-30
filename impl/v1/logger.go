package v1

import (
	"os"
	"sync"

	v1 "github.com/redresseur/loggerservice/protos/v1"
	"github.com/redresseur/loggerservice/utils/ioutils"
)

type LoggerV1 struct {
	*os.File
	path     string
	clientId string
	sync.Mutex
	id     string
	fileId string
}

func (lg *LoggerV1) write(message *v1.Message, logger *LoggerV1) error {
	// 把数据写入日志
	logger.Lock()
	defer logger.Unlock()
	_, err := logger.Write([]byte(message.GetMessage()))
	if err != nil {
		return err
	}

	// 主要判断closed 和 fd 无效
	if err == os.ErrInvalid || os.IsNotExist(err) || err == os.ErrClosed {
		var openErr error
		//other := filepath.Join(ls.tempPath, logger.clientId, logger.fileId)
		//if logger.File, openErr = ioutils.OpenFile(path, other); openErr != nil{
		//	return err
		//}

		if logger.File, openErr = ioutils.OpenFile(lg.path, ""); openErr != nil {
			return err
		}

		_, err = logger.Write([]byte(message.GetMessage()))
	}

	return err
}
