package protocol

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/jamjarlabs/jamjar-relay-server/specs/v1/transport"
	"google.golang.org/protobuf/proto"
)

const internalServerErrMessage = "An internal server error occurred"

// Fail converts an error to a payload in bytes, while logging any internal server errors
func Fail(failure *transport.Error) []byte {
	if failure.Code == http.StatusInternalServerError {
		glog.Error(failure.Message)
		failure.Message = internalServerErrMessage
	}

	failureBytes, err := proto.Marshal(failure)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	networkMessage := transport.Payload{
		Flag: transport.Payload_RESPONSE_ERROR,
		Data: failureBytes,
	}

	response, err := proto.Marshal(&networkMessage)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	return response
}

// Succeed converts a successful network message to bytes
func Succeed(networkMessage *transport.Payload) []byte {
	response, err := proto.Marshal(networkMessage)
	if err != nil {
		// Should not occur, panic
		panic(err)
	}

	return response
}
