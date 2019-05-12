package errors

import "errors"

var (
	ErrorIDIsEmpty  = errors.New("the id is empty")
	ErrorLoggerIDIsNotValid = errors.New("the logger id is invalid")
	ErrorFunctionNotSupported = errors.New("sorry, this function is not supported")
	ErrorNoMatchProtocol = errors.New("don't found matched protocol")
)

