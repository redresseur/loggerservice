package client

import (
	"context"
	"io"
	"os"
	"runtime"
	"time"
)

type Record struct {
	Path string `json:"path"`
	Size uint64 `json:"size"`
	Offset uint64 `json:"offset"`
}

// 同步本地的日志到远程日志服务器
// 监听本地各个服务的状态

// 枚举本地日志
func listLocalLogs()[]string{
	return nil
}

func RegistryTask()  {
	
}

type task struct {
	path string
	reader *os.File
	record *Record
	remoteWriter io.Writer
	ctx context.Context
	module string
}

func (t *task)flashRecord(){

}

func (t *task) sync() {
	runtime.LockOSThread()
	runtime.UnlockOSThread()
	for true {
		// 每间隔2倍的心跳时间刷新一次
		timer := time.After(2*gGrpcHeartTime)
		select {
		case <-timer:
			// 判断writer是否可用
			//remoteWriter := *(*io.Writer)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&pd.remoteWriter))))
			if t.remoteWriter == nil {
				if gGrpcConnection != nil && gGrpcConnIsUseFul.Load().(bool){
					if newWriter, err := operatorGrpc(t.module); err != nil{
						continue
					}else {
						// 更新writer
						t.remoteWriter = newWriter
					}
				}else {
					continue
				}
			}

			var dataLen int
			data := make([]byte, dataLen, dataLen + 1)
			t.reader.Read(data)
			if _, err := t.remoteWriter.Write(data); err != nil{
				t.remoteWriter = nil
			}else {
				t.flashRecord()
			}
		case <-t.ctx.Done():
			return
		}
	}
}
