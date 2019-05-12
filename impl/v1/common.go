package v1

import (
	"github.com/golang/protobuf/proto"
	"github.com/redresseur/loggerservice/protos/v1"
)

const ProtocolVersion  = 1.0

func ErrorRespond(status int32 , err error) *v1.Respond {
	return &v1.Respond{
		Status: status,
		Payload: []byte(err.Error()),
		Version: ProtocolVersion, // 暂时先写死
	}
}

func RegistryRespond(Id string)*v1.Respond{
	rsp := v1.RegistryRespond{Version: ProtocolVersion, LoggerId: Id}
	payload, _ := proto.Marshal(&rsp)
	return &v1.Respond{
		Status: 200,
		Payload: payload,
		Version:ProtocolVersion,
	}
}