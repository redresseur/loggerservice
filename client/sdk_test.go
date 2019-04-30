package client

import "testing"

var (
	testGrpcAddr = ":10040"
)

func TestInitSDK(t *testing.T) {
	InitSDK(testGrpcAddr)
}
