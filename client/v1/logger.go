package v1

import (
	"context"
	"errors"
	"github.com/redresseur/loggerservice/protos/v1"
	"io"
)

type LoggerIOV1 struct {
	grpcClient v1.LoggerV1Client
	ctx context.Context
}

func NewLogger(client v1.LoggerV1Client, ctx context.Context) io.Writer {
	return &LoggerIOV1{ grpcClient: client, ctx: ctx}
}

func (io *LoggerIOV1)Write(p []byte) (n int, err error){
	msg := v1.Message{
		Version: 1.0,
		Message: p,
		Tag: v1.LogMessageTag_COMMON,
	}

	if rsp, err :=io.grpcClient.Commit(io.ctx, &msg); err != nil{
		return -1, err
	}else {
		if rsp.Status != 200{
			return -1, errors.New(string(rsp.Payload))
		}
	}

	return len(p), nil
}

func (io *LoggerIOV1)Accident(err error) error {
	msg := v1.Message{
		Version: 1.0,
		Message: []byte(err.Error()),
		Tag: v1.LogMessageTag_COMMON,
	}

	if rsp, err :=io.grpcClient.Commit(io.ctx, &msg); err != nil{
		return err
	}else {
		if rsp.Status != 200{
			return errors.New(string(rsp.Payload))
		}
	}

	return nil
}