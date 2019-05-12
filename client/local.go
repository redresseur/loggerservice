package client

import (
	"github.com/redresseur/loggerservice/common"
	"github.com/redresseur/loggerservice/utils/ioutils"
	"io"
	"path/filepath"
)

var (
	gLocalRootDir string
)

// 修改之后可以支持多个模块
func operatorLocal(module string) (io.Writer, error)  {
	path := filepath.Join(gLocalRootDir, module)
	if _, err := ioutils.CreateDirIfMissing(path); err != nil{
		return nil, err
	}

	path = filepath.Join(path, common.NewLogFileID(module))
	localWriter := ioutils.NewFileMutexIO(true)
	if fd , err := ioutils.OpenFile(path, ""); err != nil{
		return nil, err
	}else {
		localWriter.SetPath(path)
		localWriter.Set(fd)
	}

	return localWriter, nil
}