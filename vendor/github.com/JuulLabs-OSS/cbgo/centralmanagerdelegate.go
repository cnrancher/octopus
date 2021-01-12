package cbgo

// CentralManagerDelegate: https://developer.apple.com/documentation/corebluetooth/cbcentralmanagerdelegate
type CentralManagerDelegate interface {
	// DidConnectPeripheral: https://developer.apple.com/documentation/corebluetooth/cbcentralmanagerdelegate/1518969-centralmanager
	DidConnectPeripheral(cmgr CentralManager, prph Peripheral)

	// DidDisconnectPeripheral: https://developer.apple.com/documentation/corebluetooth/cbcentralmanagerdelegate/1518791-centralmanager
	DidDisconnectPeripheral(cmgr CentralManager, prph Peripheral, err error)

	// DidFailToConnectPeripheral: https://developer.apple.com/documentation/corebluetooth/cbcentralmanagerdelegate/1518988-centralmanager
	DidFailToConnectPeripheral(cmgr CentralManager, prph Peripheral, err error)

	// DidDiscoverPeripheral: https://developer.apple.com/documentation/corebluetooth/cbcentralmanagerdelegate/1518937-centralmanager
	DidDiscoverPeripheral(cmgr CentralManager, prph Peripheral, advFields AdvFields, rssi int)

	// CentralManagerDidUpdateState: https://developer.apple.com/documentation/corebluetooth/cbcentralmanagerdelegate/1518888-centralmanagerdidupdatestate
	CentralManagerDidUpdateState(cmgr CentralManager)

	// CentralManagerWillRestoreState: https://developer.apple.com/documentation/corebluetooth/cbcentralmanagerdelegate/1518819-centralmanager
	CentralManagerWillRestoreState(cmgr CentralManager, opts CentralManagerRestoreOpts)
}

// CentralManagerDelegateBase implements the CentralManagerDelegate interface
// with stub functions.  Embed this in your delegate type if you only want to
// define a subset of the CentralManagerDelegate interface.
type CentralManagerDelegateBase struct {
}

func (b *CentralManagerDelegateBase) DidConnectPeripheral(cmgr CentralManager, prph Peripheral) {
}
func (b *CentralManagerDelegateBase) DidFailToConnectPeripheral(cmgr CentralManager, prph Peripheral, err error) {
}
func (b *CentralManagerDelegateBase) DidDisconnectPeripheral(cmgr CentralManager, prph Peripheral, err error) {
}
func (b *CentralManagerDelegateBase) CentralManagerDidUpdateState(cmgr CentralManager) {
}
func (b *CentralManagerDelegateBase) CentralManagerWillRestoreState(cmgr CentralManager, opts CentralManagerRestoreOpts) {
}
func (b *CentralManagerDelegateBase) DidDiscoverPeripheral(cmgr CentralManager, prph Peripheral, advFields AdvFields, rssi int) {
}
