package client

import (
	"context"

	"github.com/redresseur/loggerservice/utils/ioutils"
	"github.com/redresseur/loggerservice/utils/sturcture"
	"io"
	"os"
	"runtime"
	"time"
)



// 可以用于实际生产的IO接口
// 要保证健壮性
// 日志优先写入本地文件
type productionIO struct {
	queue *structure.Queue
	localWriter io.Writer
	remoteWriter io.Writer
	localReader io.Reader
	createTime time.Time
	module string
	ctx context.Context
}

func (pd *productionIO)Write(p []byte) (n int, err error)  {
	pd.queue.Push(p)
	pd.queue.SingleUP(false)
	return len(p), nil
}

func (pd *productionIO)flushRecord(){

}

func (pd *productionIO)checkTime()(err error){
	//
	createDate :=  pd.createTime.Unix()/0x15180
	nowDate := time.Now().Unix()/0x15180
	if createDate != nowDate{
		// 更新localWriter
		pd.localWriter, _ = operatorLocal(pd.module)
	}
	pd.createTime = time.Now()
	pd.flushRecord()
	return
}


func (pd *productionIO)checkSize()(err error){
	fd, _ := pd.localWriter.(*ioutils.FileMutexIO)
	var fdInfo os.FileInfo
	fdInfo, err = fd.Stat()

	// 大小限制在10M
	if fdInfo.Size() > 0xc00000{
		pd.localWriter,_ = operatorLocal(pd.module)
	}
	pd.createTime = time.Now()
	pd.flushRecord()
	return
}

func (pd *productionIO)check() error {
	if err := pd.checkTime(); err != nil{
		return err
	}

	if err := pd.checkTime(); err != nil{
		return err
	}

	return nil
}

func (pd *productionIO)readAll()[]byte{
	res := []byte{}
	for true {
		data, isOK := pd.queue.Pop().([]byte)
		if data == nil || !isOK{
			break
		}
		res = append(res, data ...)
	}

	return res
}


func (pd *productionIO)worker(){
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for true{
		select{
		case _, isOK := <- pd.queue.Single() :
			if !isOK{
				return
			}

			//ch.buffer.Write(data)
			//应该无脑写内存
			data := pd.readAll()
			// 有可能读出空数据
			if len(data) == 0{
				pd.queue.SingleDown()
				continue
			}

			// TODO: 为了防止写数据过大，此处应该分块处理比较好
			// 先同步到本地
			//dataLen, _ := pd.localWriter.Write(data)

			// 再推送到服务器


			//if buffData := ch.buffer.Bytes(); len(buffData) != 0{
			//	data = append(buffData, data...)
			//	ch.buffer = bytes.NewBuffer(nil)
			//}
			//
			//if _ , err := ch.realWriter.Write(data); err!= nil{
			//	// 如果写入错误，就暂时缓存在内存中
			//	ch.buffer.Write(data)
			//	ch.redirect.Store(true)
			//	ch.reopen()
			//}

			pd.queue.SingleDown()
		case <-pd.ctx.Done():
			return
		}
	}
}