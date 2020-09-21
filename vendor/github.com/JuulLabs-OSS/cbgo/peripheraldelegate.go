package cbgo

// PeripheralDelegate: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate
type PeripheralDelegate interface {
	// DidDiscoverServices: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518744-peripheral
	DidDiscoverServices(prph Peripheral, err error)

	// DidDiscoverIncludedServices: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1519124-peripheral
	DidDiscoverIncludedServices(prph Peripheral, svc Service, err error)

	// DidDiscoverCharacteristics: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518821-peripheral
	DidDiscoverCharacteristics(prph Peripheral, svc Service, err error)

	// DidDiscoverDescriptors: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518785-peripheral
	DidDiscoverDescriptors(prph Peripheral, chr Characteristic, err error)

	// DidUpdateValueForCharacteristic: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518708-peripheral
	DidUpdateValueForCharacteristic(prph Peripheral, chr Characteristic, err error)

	// DidUpdateValueForDescriptor: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518929-peripheral
	DidUpdateValueForDescriptor(prph Peripheral, dsc Descriptor, err error)

	// DidWriteValueForCharacteristic: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518823-peripheral
	DidWriteValueForCharacteristic(prph Peripheral, chr Characteristic, err error)

	// DidWriteValueForDescriptor: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1519062-peripheral
	DidWriteValueForDescriptor(prph Peripheral, dsc Descriptor, err error)

	// IsReadyToSendWriteWithoutResponse: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/2874034-peripheralisreadytosendwritewith
	IsReadyToSendWriteWithoutResponse(prph Peripheral)

	// DidUpdateNotificationState: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518768-peripheral
	DidUpdateNotificationState(prph Peripheral, chr Characteristic, err error)

	// DidReadRSSI: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1620304-peripheral
	DidReadRSSI(prph Peripheral, rssi int, err error)

	// DidUpdateName: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518801-peripheraldidupdatename
	DidUpdateName(prph Peripheral)

	// DidModifyServices: https://developer.apple.com/documentation/corebluetooth/cbperipheraldelegate/1518865-peripheral
	DidModifyServices(prph Peripheral, invSvcs []Service)
}

// PeripheralDelegateBase implements the PeripheralDelegate interface with stub
// functions.  Embed this in your delegate type if you only want to define a
// subset of the PeripheralDelegate interface.
type PeripheralDelegateBase struct {
}

func (b *PeripheralDelegateBase) DidDiscoverServices(prph Peripheral, err error) {
}
func (b *PeripheralDelegateBase) DidDiscoverIncludedServices(prph Peripheral, svc Service, err error) {
}
func (b *PeripheralDelegateBase) DidDiscoverCharacteristics(prph Peripheral, svc Service, err error) {
}
func (b *PeripheralDelegateBase) DidDiscoverDescriptors(prph Peripheral, chr Characteristic, err error) {
}
func (b *PeripheralDelegateBase) DidUpdateValueForCharacteristic(prph Peripheral, chr Characteristic, err error) {
}
func (b *PeripheralDelegateBase) DidUpdateValueForDescriptor(prph Peripheral, dsc Descriptor, err error) {
}
func (b *PeripheralDelegateBase) DidWriteValueForCharacteristic(prph Peripheral, chr Characteristic, err error) {
}
func (b *PeripheralDelegateBase) DidWriteValueForDescriptor(prph Peripheral, dsc Descriptor, err error) {
}
func (b *PeripheralDelegateBase) IsReadyToSendWriteWithoutResponse(prph Peripheral) {
}
func (b *PeripheralDelegateBase) DidUpdateNotificationState(prph Peripheral, chr Characteristic, err error) {
}
func (b *PeripheralDelegateBase) DidReadRSSI(prph Peripheral, rssi int, err error) {
}
func (b *PeripheralDelegateBase) DidUpdateName(prph Peripheral) {
}
func (b *PeripheralDelegateBase) DidModifyServices(prph Peripheral, invSvcs []Service) {
}
