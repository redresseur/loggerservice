package client

import (
	"errors"
	"github.com/redresseur/loggerservice/common"
	"github.com/redresseur/loggerservice/utils/ioutils"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"
)

var (
	testGrpcAddr = ":10040"
	testGrpcAddr1 = ":10041"
	testModule = "sdk"
	once sync.Once = sync.Once{}
)

func TestInitSDKWithLocal(t *testing.T) {
	defer ReleaseSDK()
	InitSDK()
}

func TestOpenLoggerWithLocal(t *testing.T) {

	InitSDK()
	writer, _ := OpenChannelIoLogger(testModule)
	for i :=0 ; i < 10000; i++{
		data := "hello world: " + strconv.Itoa(i) +"\n"
		writer.Write([]byte(data))
	}
	ReleaseSDK()
}

func TestInitSDKWithGrpc(t *testing.T) {
	defer ReleaseSDK()
	InitSDK(WithLoggerServerAddr(testGrpcAddr), WithLoggerType(LoggerGrpc))
}

func TestOpenLoggerWithGrpc(t *testing.T) {
	InitSDK(WithLoggerServerAddr(testGrpcAddr), WithLoggerType(LoggerGrpc))
	if loggerIn, err  := OpenChannelIoLogger(testModule); err != nil{
		t.Fatalf("TestOpenLogger %v", err)
	}else {
		for i :=0 ; i < 10000; i++{
			data := "hello world: " + strconv.Itoa(i) +"\n"
			if _, err := loggerIn.Write([]byte(data)); err != nil{
				t.Fatalf("WriteFile %v", err)
			}
		}

		CloseLogger(loggerIn)
		t.Logf("Passed")
	}

	ReleaseSDK()
}

func TestReconnect(t *testing.T)  {
	InitSDK(WithLoggerServerAddr(testGrpcAddr),WithLoggerServerAddr(testGrpcAddr1), WithLoggerType(LoggerGrpc))
	defer ReleaseSDK()
	if loggerIn, err  := OpenChannelIoLogger(testModule); err != nil{
		t.Fatalf("TestOpenLogger %v", err)
	}else {
		defer CloseLogger(loggerIn)
		for i :=0 ; i < 10000; i++{
			time.Sleep(time.Millisecond)
			data := "hello world: " + strconv.Itoa(i) +"\n"
			if _, err := loggerIn.Write([]byte(data)); err != nil{
				t.Fatalf("WriteFile %v", err)
			}
		}

		for i := 10000 ; i < 20000; i++{
			time.Sleep(time.Millisecond)
			data := "hello world: " + strconv.Itoa(i) +"\n"
			if _, err := loggerIn.Write([]byte(data)); err != nil{
				t.Fatalf("WriteFile %v", err)
			}
		}


		t.Logf("Passed")
	}

	time.Sleep(5*time.Second)

}

func TestOpenAccidentWithGrpc(t *testing.T) {
	InitSDK(WithLoggerServerAddr(testGrpcAddr), WithLoggerType(LoggerGrpc))
	if loggerIn, err  := OpenChannelIoLogger(testModule); err != nil{
		t.Fatalf("OpenLogger %v", err)
	}else {
		defer CloseLogger(loggerIn)
		acc, _ := OpenAccident(loggerIn)
		acc(errors.New("Accident!!!!!"))
		t.Logf("Passed")
	}

	ReleaseSDK()
}

func BenchmarkWithOpenLoggerWithGrpc1(b *testing.B) {
	once.Do(func() {
		InitSDK(WithLoggerServerAddr(testGrpcAddr), WithLoggerType(LoggerGrpc))
	})

	if loggerIn, err  := openLogger(testModule); err != nil{
		b.Fatalf("TestOpenLogger %v", err)
	}else {
		defer CloseLogger(loggerIn)
		for i :=0 ; i < b.N; i++{
			data := "hello world: " + strconv.Itoa(i) +"\n"
			if _, err := loggerIn.Write([]byte(data)); err != nil{
				b.Fatalf("WriteFile %v", err)
			}
		}

		b.Logf("Passed")
	}

	//defer ReleaseSDK()
}

func BenchmarkWithOpenLoggerWithGrpc(b *testing.B) {
	once.Do(func() {
		InitSDK(WithLoggerServerAddr(testGrpcAddr), WithLoggerType(LoggerGrpc))
	})

	if loggerIn, err  := OpenChannelIoLogger(testModule); err != nil{
		b.Fatalf("TestOpenLogger %v", err)
	}else {
		defer CloseLogger(loggerIn)
		for i :=0 ; i < b.N; i++{
			data := "hello world: " + strconv.Itoa(i) +"\n"
			if _, err := loggerIn.Write([]byte(data)); err != nil{
				b.Fatalf("WriteFile %v", err)
			}
		}

		b.Logf("Passed")
	}

	//defer ReleaseSDK()
}

func BenchmarkWithOpenLoggerWithFile(b *testing.B) {
	path := filepath.Join(ioutils.TempDir(), common.NewLogFileID("test") )
	if loggerIn, err  := ioutils.OpenFile(path, ""); err != nil{
		b.Fatalf("TestOpenLogger %v", err)
	}else {
		defer CloseLogger(loggerIn)
		for i :=0 ; i < b.N; i++{
			data := "hello world: " + strconv.Itoa(i) +"\n"
			if _, err := loggerIn.Write([]byte(data)); err != nil{
				b.Fatalf("WriteFile %v", err)
			}
		}

		b.Logf("Passed")
	}

	//defer ReleaseSDK()
}

func BenchmarkWithOpenLoggerWithLocal(b *testing.B) {
	once.Do(func() {
		InitSDK(WithLoggerServerAddr(testGrpcAddr), WithLoggerType(LoggerLocal))
	})

	if loggerIn, err  := OpenChannelIoLogger(testModule); err != nil{
		b.Fatalf("TestOpenLogger %v", err)
	}else {
		defer CloseLogger(loggerIn)
		for i :=0 ; i < b.N; i++{
			data := "hello world: " + strconv.Itoa(i) +"\n"
			if _, err := loggerIn.Write([]byte(data)); err != nil{
				b.Fatalf("WriteFile %v", err)
			}
		}
		//CloseLogger(loggerIn)
		b.Logf("Passed")
	}

	//defer ReleaseSDK()
}
