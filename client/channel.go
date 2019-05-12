package client

import (
	"bytes"
	"context"
	"github.com/redresseur/loggerservice/utils/ioutils"
	"github.com/redresseur/loggerservice/utils/sturcture"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var(
	gChannelIO *channelIO
)

// 还要再封装一层
// 用channel封装一层
type channelIO struct {
	// 加一个大的缓存队列
	io *structure.Queue
	ctx context.Context
	realWriter io.Writer
	writerType LoggerType
	buffer *bytes.Buffer
	module string
	redirect *atomic.Value
	cancel func()
}

func NewChannelIO(writer io.Writer, module string)*channelIO{
	ch := channelIO{}
	ch.ctx, ch.cancel = context.WithCancel(gSdkCtx)
	// 设置缓存4096
	ch.io = structure.New()
	ch.realWriter = writer
	ch.buffer = bytes.NewBuffer(nil)
	ch.module = module
	ch.redirect = &atomic.Value{}
	ch.redirect.Store(false)
	ch.writerType = gLoggerType

	go ch.worker()
	return &ch
}

func (ch *channelIO)Write(p []byte) (n int, err error)  {
	ch.io.Push(p)
	ch.io.SingleUP(false)
	return len(p), nil
}

// 当grpc切换到本地的时候，
// 不断的轮训，尝试切换回来
func (ch *channelIO)switchBack(){
	// realType := ch.writerType
	for true {
		timer := time.After(gGrpcHeartTime)
		select {
		case <-ch.ctx.Done():
			return
		case <-timer:
			if gGrpcConnection == nil || !gGrpcConnIsUseFul.Load().(bool){
				continue
			}

			newWrite, err := operatorGrpc(ch.module)
			if err != nil{
				continue
			}
			// 日志暂存在内存
			ch.redirect.Store(true)
			// 切换write
			oldWriter := ch.realWriter
			// 首先将缓存在本地的数据读取出来
			if ch.writerType == LoggerLocal{
				fd, _ := oldWriter.(*ioutils.FileMutexIO)

				// 定位到文档起点
				fd.Sync()
				fd.Seek(0, 0)
				// TODO: 此处将来要优化，数据太大写入容易崩溃
				if data, err := ioutil.ReadAll(fd); err == nil{
					newWrite.Write(data)
				}

				defer func() {
					fd.Close()
					os.Remove(fd.Path())
				}()
			}
			// 切换到新的write上
			ch.realWriter = newWrite
			ch.redirect.Store(false)
			ch.io.SingleUP(true)
			ch.writerType = LoggerGrpc
			return
		}
	}
}

// TODO: 后面加了配置监督可能会用到这个接口
func (ch *channelIO)reopen(){
	// 重新打开一个接口
	// 如果打开失败
	switch ch.writerType {
	case LoggerGrpc:
		// 转存本地
		newWriter, err := openLogger(ch.module)
		if err == nil{
			ch.realWriter = newWriter

			// 先从重定向状态中切换过来
			// 然后发出信号
			ch.redirect.Store(false)
			ch.io.SingleUP(true)
		}else {
			ch.writerType = LoggerLocal
			// 尝试切换回来
			go ch.switchBack()
			ch.reopen()
		}
	case LoggerLocal:
		// 直接打印到控制台
		var err error
		if ch.realWriter, err = operatorLocal(ch.module); err != nil{
			ch.writerType = LoggerStd
			ch.reopen()
		}else {
			ch.redirect.Store(false)
			ch.io.SingleUP(true)
		}
	case LoggerStd:
		ch.realWriter = os.Stdout
		ch.redirect.Store(false)
		ch.io.SingleUP(true)
	}
}

func (ch *channelIO)readAll()[]byte{
	res := []byte{}
	for true {
		data, isOK := ch.io.Pop().([]byte)
		if data == nil || !isOK{
			break
		}
		res = append(res, data ...)
	}

	return res
}

func (ch *channelIO)worker(){
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for true{
		select{
		case _, isOK := <- ch.io.Single() :
			if !isOK{
				return
			}

			//ch.buffer.Write(data)
			//应该无脑写内存
			data := ch.readAll()
			// 有可能读出空数据
			if len(data) == 0{
				ch.io.SingleDown()
				continue
			}

			if ch.redirect.Load().(bool){
				ch.buffer.Write(data)
				ch.io.SingleDown()
				continue
			}

			// TODO: 为了防止写数据过大，此处应该分块处理比较好
			if buffData := ch.buffer.Bytes(); len(buffData) != 0{
				data = append(buffData, data...)
				ch.buffer = bytes.NewBuffer(nil)
			}

			if _ , err := ch.realWriter.Write(data); err!= nil{
				// 如果写入错误，就暂时缓存在内存中
				ch.buffer.Write(data)
				ch.redirect.Store(true)
				ch.reopen()
			}

			ch.io.SingleDown()
		case <-ch.ctx.Done():
			return
		}
	}
}

func (ch* channelIO)Close(){
	ch.cancel()

	// 写入残留数据
	data := ch.buffer.Bytes()
	if qData := ch.readAll(); qData != nil{
		data = append(data, qData...)
	}

	if len(data) > 0{
		ch.realWriter.Write(data)
	}

	ch.io.Close()
	if fd, isOK := ch.realWriter.(*ioutils.FileMutexIO); isOK{
		fd.Sync()
		fd.Close()
	}
}