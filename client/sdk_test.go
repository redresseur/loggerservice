package client

import (
	"errors"
	"testing"
)

var (
	testGrpcAddr = "192.168.1.160:10040"
	testModule = "sdk"
)

func TestInitSDK(t *testing.T) {
	defer ReleaseSDK()
	InitSDK(testGrpcAddr)
}

func TestOpenLogger(t *testing.T) {
	defer ReleaseSDK()
	InitSDK(testGrpcAddr)
	if loggerIn, err  := OpenLogger(testModule); err != nil{
		t.Fatalf("TestOpenLogger %v", err)
	}else {
		if _, err := loggerIn.Write([]byte("hello world!!")); err != nil{
			t.Fatalf("WriteFile %v", err)
		}
		t.Logf("Passed")
	}
}

func TestOpenAccident(t *testing.T) {
	defer ReleaseSDK()
	InitSDK(testGrpcAddr)
	if loggerIn, err  := OpenLogger(testModule); err != nil{
		t.Fatalf("OpenLogger %v", err)
	}else {
		acc:= OpenAccident(loggerIn)
		acc(errors.New("Accident!!!!!"))
		t.Logf("Passed")
	}
}