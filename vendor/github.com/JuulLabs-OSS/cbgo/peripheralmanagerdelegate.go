package cbgo

// https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate
type PeripheralManagerDelegate interface {
	// PeripheralManagerDidUpdateState: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393271-peripheralmanagerdidupdatestate
	PeripheralManagerDidUpdateState(pmgr PeripheralManager)

	// PeripheralManagerWillRestoreState: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393317-peripheralmanager
	PeripheralManagerWillRestoreState(pmgr PeripheralManager, opts PeripheralManagerRestoreOpts)

	// DidAddService: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393279-peripheralmanager
	DidAddService(pmgr PeripheralManager, svc Service, err error)

	// DidStartAdvertising: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393321-peripheralmanagerdidstartadverti
	DidStartAdvertising(pmgr PeripheralManager, err error)

	// CentralDidSubscribe: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393261-peripheralmanager
	CentralDidSubscribe(pmgr PeripheralManager, cent Central, chr Characteristic)

	// CentralDidUnsubscribe: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393289-peripheralmanager
	CentralDidUnsubscribe(pmgr PeripheralManager, cent Central, chr Characteristic)

	// IsReadyToUpdateSubscribers: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393248-peripheralmanagerisreadytoupdate
	IsReadyToUpdateSubscribers(pmgr PeripheralManager)

	// DidReceiveReadRequest: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393257-peripheralmanager
	DidReceiveReadRequest(pmgr PeripheralManager, req ATTRequest)

	// DidReceiveWriteRequests: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/1393315-peripheralmanager
	DidReceiveWriteRequests(pmgr PeripheralManager, reqs []ATTRequest)
}

// PeripheralManagerDelegateBase implements the PeripheralManagerDelegate
// interface with stub functions.  Embed this in your delegate type if you only
// want to define a subset of the PeripheralManagerDelegate interface.
type PeripheralManagerDelegateBase struct {
}

func (b *PeripheralManagerDelegateBase) PeripheralManagerDidUpdateState(pmgr PeripheralManager) {
}
func (b *PeripheralManagerDelegateBase) PeripheralManagerWillRestoreState(pmgr PeripheralManager, opts PeripheralManagerRestoreOpts) {
}
func (b *PeripheralManagerDelegateBase) DidAddService(pmgr PeripheralManager, svc Service, err error) {
}
func (b *PeripheralManagerDelegateBase) DidStartAdvertising(pmgr PeripheralManager, err error) {
}
func (b *PeripheralManagerDelegateBase) CentralDidSubscribe(pmgr PeripheralManager, cent Central, chr Characteristic) {
}
func (b *PeripheralManagerDelegateBase) CentralDidUnsubscribe(pmgr PeripheralManager, cent Central, chr Characteristic) {
}
func (b *PeripheralManagerDelegateBase) IsReadyToUpdateSubscribers(pmgr PeripheralManager) {
}
func (b *PeripheralManagerDelegateBase) DidReceiveReadRequest(pmgr PeripheralManager, req ATTRequest) {
}
func (b *PeripheralManagerDelegateBase) DidReceiveWriteRequests(pmgr PeripheralManager, reqs []ATTRequest) {
}
