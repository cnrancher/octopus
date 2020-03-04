package connection

import (
	"io"

	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

func isActiveClosed(err error) bool {
	if err == io.EOF {
		return true
	}
	if status, ok := grpcstatus.FromError(err); ok {
		return status.Code() == grpccodes.Canceled && status.Message() == "context canceled"
	}
	return false
}

func isPassiveClosed(err error) bool {
	if err == io.EOF {
		return true
	}
	if status, ok := grpcstatus.FromError(err); ok {
		return status.Code() == grpccodes.Unavailable && status.Message() == "transport is closing"
	}
	return false
}
