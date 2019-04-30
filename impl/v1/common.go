package v1

import (
	"errors"
	"github.com/gogo/protobuf/proto"
	"github.com/redresseur/loggerservice/protos/v1"
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

const ProtocolVersion  = 1.0

var (
	ErrorIDIsEmpty  = errors.New("the id is empty")
	ErrorLoggerIDIsNotValid = errors.New(" the logger id is invalid")
)

func NewClientUUID()string  {
	return uuid.Must(uuid.NewV1()).String()
}

func NewLogFileID(ID string) string{
	res := ID + "-" + time.Now().Format("2006-01-02")+ "-" + uuid.Must(uuid.NewV1()).String() + ".log"
	return strings.NewReplacer("-", "_").Replace(res)
}

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