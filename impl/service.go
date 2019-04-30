package impl

import (
	"context"

	v1 "github.com/redresseur/loggerservice/protos/v1"
)

type LoggerServerImpl struct {
	ioSet map[string]string // the log files io sets
}

func (ls *LoggerServerImpl) Commit(context.Context, *v1.LogMessageRequest) (*v1.LogMessageReply, error) {
	return nil, nil
}
