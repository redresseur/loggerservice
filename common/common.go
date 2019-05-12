package common

import (
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

func NewClientUUID()string  {
	return uuid.Must(uuid.NewV1()).String()
}

func NewLogFileID(ID string) string{
	res := ID + "-" + time.Now().Format("2006-01-02-04-05")+ "-" + uuid.Must(uuid.NewV1()).String() + ".log"
	return strings.NewReplacer("-", "_").Replace(res)
}