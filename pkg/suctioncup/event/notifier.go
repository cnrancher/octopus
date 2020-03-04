package event

import (
	"k8s.io/apimachinery/pkg/types"
)

type AdaptorNotifier interface {
	NoticeAdaptorRegistered(name string)
	NoticeAdaptorUnregistered(name string)
}

type ConnectionNotifier interface {
	NoticeConnectionReceivedData(adaptorName string, name types.NamespacedName, data []byte)
	NoticeConnectionReceivedError(adaptorName string, name types.NamespacedName, err error)
	NoticeConnectionClosed(adaptorName string, name types.NamespacedName)
}
