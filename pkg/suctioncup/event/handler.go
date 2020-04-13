package event

type AdaptorHandler interface {
	ReceiveAdaptorStatus(req RequestAdaptorStatus) (Response, error)
}

type ConnectionHandler interface {
	ReceiveConnectionStatus(req RequestConnectionStatus) (Response, error)
}
