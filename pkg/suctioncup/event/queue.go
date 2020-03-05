package event

import (
	"sync"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
)

type Queue interface {
	ShutDown()
	Reconcile()
	RegisterAdaptorHandler(handler AdaptorHandler)
	RegisterConnectionHandler(handler ConnectionHandler)
	GetAdaptorNotifier() AdaptorNotifier
	GetConnectionNotifier() ConnectionNotifier
}

func NewQueue(name string) Queue {
	var q = workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		name,
	)
	return &queue{
		queue: q,
	}
}

type queue struct {
	queue             workqueue.RateLimitingInterface
	adaptorHandler    AdaptorHandler
	connectionHandler ConnectionHandler

	receivedDataCache  sync.Map
	receivedErrorCache sync.Map
}

func (q *queue) ShutDown() {
	q.queue.ShutDown()
}

func (q *queue) Reconcile() {
	for q.reconcileNext() {
	}
}

func (q *queue) RegisterAdaptorHandler(handler AdaptorHandler) {
	q.adaptorHandler = handler
}

func (q *queue) RegisterConnectionHandler(handler ConnectionHandler) {
	q.connectionHandler = handler
}

func (q *queue) GetConnectionNotifier() ConnectionNotifier {
	return q
}

func (q *queue) GetAdaptorNotifier() AdaptorNotifier {
	return q
}

func (q *queue) NoticeConnectionReceivedData(adaptorName string, name types.NamespacedName, data []byte) {
	var key = connectionReceivedData{
		adaptorName: adaptorName,
		name:        name,
	}
	q.receivedDataCache.Store(key, data)
	q.queue.AddRateLimited(key)
}

func (q *queue) NoticeConnectionReceivedError(adaptorName string, name types.NamespacedName, err error) {
	var key = connectionReceivedError{
		adaptorName: adaptorName,
		name:        name,
	}
	q.receivedErrorCache.Store(key, err)
	q.queue.AddRateLimited(key)
}

func (q *queue) NoticeConnectionClosed(adaptorName string, name types.NamespacedName) {
	var key = connectionClosed{
		adaptorName: adaptorName,
		name:        name,
	}
	q.queue.AddRateLimited(key)
}

func (q *queue) NoticeAdaptorRegistered(name string) {
	var key = adaptorRegistered{
		name: name,
	}
	q.queue.AddRateLimited(key)
}

func (q *queue) NoticeAdaptorUnregistered(name string) {
	var key = adaptorUnregistered{
		name: name,
	}
	q.queue.AddRateLimited(key)
}

func (q *queue) reconcileNext() bool {
	obj, shutdown := q.queue.Get()
	if shutdown {
		return false
	}

	defer q.queue.Done(obj)
	return q.reconcileHandler(obj)
}

func (q *queue) reconcileHandler(obj interface{}) bool {
	var (
		resp Response
		err  error
	)

	switch req := obj.(type) {
	case adaptorRegistered:
		resp, err = q.ReceiveAdaptorStatus(RequestAdaptorStatus{
			Name:       req.name,
			Registered: true,
		})
	case adaptorUnregistered:
		resp, err = q.ReceiveAdaptorStatus(RequestAdaptorStatus{
			Name:       req.name,
			Registered: false,
		})
	case connectionReceivedError:
		if obj, ok := q.receivedErrorCache.Load(req); ok {
			resp, err = q.ReceiveConnectionStatus(RequestConnectionStatus{
				AdaptorName: req.adaptorName,
				Name:        req.name,
				Error:       obj.(error),
			})

			if resp.RequeueAfter <= 0 && !resp.Requeue {
				q.receivedErrorCache.Delete(req)
			}
		}
	case connectionReceivedData:
		if obj, ok := q.receivedDataCache.Load(req); ok {
			resp, err = q.ReceiveConnectionStatus(RequestConnectionStatus{
				AdaptorName: req.adaptorName,
				Name:        req.name,
				Data:        obj.([]byte),
			})

			if resp.RequeueAfter <= 0 && !resp.Requeue {
				q.receivedDataCache.Delete(req)
			}
		}
	case connectionClosed:
		resp, err = q.ReceiveConnectionStatus(RequestConnectionStatus{
			AdaptorName: req.adaptorName,
			Name:        req.name,
			Closed:      true,
		})
	default:
		q.queue.Forget(obj)
		// Return true, don't take a break
		return true
	}

	if err != nil {
		q.queue.AddRateLimited(obj)
		return false
	} else if resp.RequeueAfter > 0 {
		q.queue.Forget(obj)
		q.queue.AddAfter(obj, resp.RequeueAfter)
		return true
	} else if resp.Requeue {
		q.queue.AddRateLimited(obj)
		return true
	}

	q.queue.Forget(obj)
	return true
}

// proxy
func (q *queue) ReceiveConnectionStatus(req RequestConnectionStatus) (Response, error) {
	if q.connectionHandler == nil {
		return Response{}, nil
	}
	return q.connectionHandler.ReceiveConnectionStatus(req)
}

// proxy
func (q *queue) ReceiveAdaptorStatus(req RequestAdaptorStatus) (Response, error) {
	if q.adaptorHandler == nil {
		return Response{}, nil
	}
	return q.adaptorHandler.ReceiveAdaptorStatus(req)
}
