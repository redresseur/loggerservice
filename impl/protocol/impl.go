package protocol

import (
	"context"

	"github.com/redresseur/loggerservice/protos/protocol"
)

type ProtocolServerImpl struct {
	protocols []float32
}

func RegistryProtocol(ps *ProtocolServerImpl, protocol ...float32) {
	ps.protocols = append(ps.protocols, protocol...)
}

// FetchProtocolInfo 用于协议磋商
// 目前仅仅支持 1.0 协议
func (ps *ProtocolServerImpl) FetchProtocolInfo(ctx context.Context, req *protocol.ProtocolRequest) (*protocol.ProtocolRespond, error) {
	rsp := protocol.ProtocolRespond{}
	rsp.SupportProtocol = append(rsp.SupportProtocol, ps.protocols...)
	return &rsp, nil
}

type PingPongImpl struct {

}

func (pp *PingPongImpl)PingIng(ctx context.Context, req *protocol.Ping) (*protocol.Pong, error)  {
	return &protocol.Pong{
		Counter: req.Counter,
	}, nil
}

