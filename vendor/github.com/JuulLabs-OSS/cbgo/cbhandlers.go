package cbgo

// CBHandlers: Go handlers for asynchronous CoreBluetooth callbacks.

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"
import (
	"unsafe"
)

//export BTCentralManagerDidConnectPeripheral
func BTCentralManagerDidConnectPeripheral(cmgr unsafe.Pointer, prph unsafe.Pointer) {
	btlog.Debugf("CentralManagerDidConnectPeripheral: cmgr=%v prph=%v", cmgr, prph)

	dlg := findCentralManagerDlg(cmgr)
	if dlg != nil {
		dlg.DidConnectPeripheral(CentralManager{cmgr}, Peripheral{prph})
	}
}

//export BTCentralManagerDidFailToConnectPeripheral
func BTCentralManagerDidFailToConnectPeripheral(cmgr unsafe.Pointer, prph unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("CentralManagerDidFailToConnectPeripheral: cmgr=%v prph=%v err=%v", cmgr, prph, nserr)

	dlg := findCentralManagerDlg(cmgr)
	if dlg != nil {
		dlg.DidFailToConnectPeripheral(CentralManager{cmgr}, Peripheral{prph}, nserr)
	}
}

//export BTCentralManagerDidDisconnectPeripheral
func BTCentralManagerDidDisconnectPeripheral(cmgr unsafe.Pointer, prph unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("CentralManagerDidDisconnectPeripheral: cmgr=%v prph=%v err=%v", cmgr, prph, nserr)

	dlg := findCentralManagerDlg(cmgr)
	if dlg != nil {
		dlg.DidDisconnectPeripheral(CentralManager{cmgr}, Peripheral{prph}, nserr)
	}
}

// macOS 10.15+
/*
//export BTCentralManagerConnectionEventDidOccur
func BTCentralManagerConnectionEventDidOccur(cmgr unsafe.Pointer, event C.int, prph unsafe.Pointer) {
	dlg := findCentralManagerDlg(cmgr)
	if dlg == nil {
		return
	}

	dlg.ConnectionEventDidOccur(CentralManager{cmgr}, ConnectionEvent(event), Peripheral{prph})
}
*/

//export BTCentralManagerDidDiscoverPeripheral
func BTCentralManagerDidDiscoverPeripheral(cmgr unsafe.Pointer, prph unsafe.Pointer, advData *C.struct_adv_fields,
	rssi C.int) {

	af := AdvFields{}
	af.LocalName = C.GoString(advData.name)
	af.ManufacturerData = byteArrToByteSlice(&advData.mfg_data)
	if advData.pwr_lvl != C.ADV_FIELDS_PWR_LVL_NONE {
		i := int(advData.pwr_lvl)
		af.TxPowerLevel = &i
	}
	if advData.connectable != C.ADV_FIELDS_CONNECTABLE_NONE {
		b := advData.connectable != 0
		af.Connectable = &b
	}

	af.ServiceUUIDs = mustStrArrToUUIDs(&advData.svc_uuids)
	svcDataUUIDs := mustStrArrToUUIDs(&advData.svc_data_uuids)

	for i, u := range svcDataUUIDs {
		elemPtr := getArrElemAddr(unsafe.Pointer(advData.svc_data_values), unsafe.Sizeof(*advData.svc_data_values), i)
		elem := (*C.struct_byte_arr)(elemPtr)
		af.ServiceData = append(af.ServiceData, AdvServiceData{
			UUID: u,
			Data: byteArrToByteSlice(elem),
		})

	}

	btlog.Debugf("CentralManagerDidDiscoverPeripheral: cmgr=%v prph=%v af=%+v rssi=%v", cmgr, prph, af, rssi)

	dlg := findCentralManagerDlg(cmgr)
	if dlg != nil {
		dlg.DidDiscoverPeripheral(CentralManager{cmgr}, Peripheral{prph}, af, int(rssi))
	}
}

//export BTCentralManagerDidUpdateState
func BTCentralManagerDidUpdateState(cmgr unsafe.Pointer) {
	btlog.Debugf("CentralManagerDidUpdateState: cmgr=%v", cmgr)

	dlg := findCentralManagerDlg(cmgr)
	if dlg != nil {
		dlg.CentralManagerDidUpdateState(CentralManager{cmgr})
	}
}

//export BTCentralManagerWillRestoreState
func BTCentralManagerWillRestoreState(cmgr unsafe.Pointer, opts *C.struct_cmgr_restore_opts) {
	ropts := CentralManagerRestoreOpts{}

	ropts.Peripherals = make([]Peripheral, opts.prphs.count)
	for i, _ := range ropts.Peripherals {
		ropts.Peripherals[i].ptr = getObjArrElem(&opts.prphs, i)
	}

	ropts.ScanServices = mustStrArrToUUIDs(&opts.scan_svcs)

	if opts.scan_opts != nil {
		ropts.CentralManagerScanOpts = &CentralManagerScanOpts{
			AllowDuplicates:       bool(opts.scan_opts.allow_dups),
			SolicitedServiceUUIDs: mustStrArrToUUIDs(&opts.scan_opts.sol_svc_uuids),
		}
	}

	btlog.Debugf("CentralManagerWillRestoreState: cmgr=%v opts=%+v", cmgr, ropts)

	dlg := findCentralManagerDlg(cmgr)
	if dlg != nil {
		dlg.CentralManagerWillRestoreState(CentralManager{cmgr}, ropts)
	}
}

//export BTPeripheralDidDiscoverServices
func BTPeripheralDidDiscoverServices(prph unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidDiscoverServices: prph=%v err=%v", prph, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidDiscoverServices(Peripheral{prph}, nserr)
	}
}

//export BTPeripheralDidDiscoverIncludedServices
func BTPeripheralDidDiscoverIncludedServices(prph unsafe.Pointer, svc unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidDiscoverIncludedServices: prph=%v svc=%v err=%v", prph, svc, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidDiscoverIncludedServices(Peripheral{prph}, Service{svc}, nserr)
	}
}

//export BTPeripheralDidDiscoverCharacteristics
func BTPeripheralDidDiscoverCharacteristics(prph unsafe.Pointer, svc unsafe.Pointer, err *C.struct_bt_error) {

	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidDiscoverCharacteristics: prph=%v svc=%v err=%v", prph, svc, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidDiscoverCharacteristics(Peripheral{prph}, Service{svc}, nserr)
	}
}

//export BTPeripheralDidDiscoverDescriptors
func BTPeripheralDidDiscoverDescriptors(prph unsafe.Pointer, chr unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidDiscoverDescriptors: prph=%v chr=%v err=%v", prph, chr, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidDiscoverDescriptors(Peripheral{prph}, Characteristic{chr}, nserr)
	}
}

//export BTPeripheralDidUpdateValueForCharacteristic
func BTPeripheralDidUpdateValueForCharacteristic(prph unsafe.Pointer, chr unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidUpdateValueForCharacteristic: prph=%v chr=%v err=%v", prph, chr, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidUpdateValueForCharacteristic(Peripheral{prph}, Characteristic{chr}, nserr)
	}
}

//export BTPeripheralDidUpdateValueForDescriptor
func BTPeripheralDidUpdateValueForDescriptor(prph unsafe.Pointer, dsc unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidUpdateValueForDescriptor: prph=%v dsc=%v err=%v", prph, dsc, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidUpdateValueForDescriptor(Peripheral{prph}, Descriptor{dsc}, nserr)
	}
}

//export BTPeripheralDidWriteValueForCharacteristic
func BTPeripheralDidWriteValueForCharacteristic(prph unsafe.Pointer, chr unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidWriteValueForCharacteristic: prph=%v chr=%v err=%v", prph, chr, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidWriteValueForCharacteristic(Peripheral{prph}, Characteristic{chr}, nserr)
	}
}

//export BTPeripheralDidWriteValueForDescriptor
func BTPeripheralDidWriteValueForDescriptor(prph unsafe.Pointer, dsc unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidWriteValueForDescriptor: prph=%v dsc=%v err=%v", prph, dsc, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidWriteValueForDescriptor(Peripheral{prph}, Descriptor{dsc}, nserr)
	}
}

//export BTPeripheralIsReadyToSendWriteWithoutResponse
func BTPeripheralIsReadyToSendWriteWithoutResponse(prph unsafe.Pointer) {
	btlog.Debugf("PeripheralIsReadyToSendWriteWithoutResponse: prph=%v", prph)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.IsReadyToSendWriteWithoutResponse(Peripheral{prph})
	}
}

//export BTPeripheralDidUpdateNotificationState
func BTPeripheralDidUpdateNotificationState(prph unsafe.Pointer, chr unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidUpdateNotificationState: prph=%v chr=%v err=%v", prph, chr, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidUpdateNotificationState(Peripheral{prph}, Characteristic{chr}, nserr)
	}
}

//export BTPeripheralDidReadRSSI
func BTPeripheralDidReadRSSI(prph unsafe.Pointer, rssi C.int, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralDidReadRSSI: prph=%v rssi=%v err=%v", prph, rssi, nserr)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidReadRSSI(Peripheral{prph}, int(rssi), nserr)
	}
}

//export BTPeripheralDidUpdateName
func BTPeripheralDidUpdateName(prph unsafe.Pointer) {
	btlog.Debugf("PeripheralDidUpdateName: prph=%v", prph)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidUpdateName(Peripheral{prph})
	}
}

//export BTPeripheralDidModifyServices
func BTPeripheralDidModifyServices(prph unsafe.Pointer, inv_svcs *C.struct_obj_arr) {
	svcs := make([]Service, inv_svcs.count)
	for i, _ := range svcs {
		elem := getObjArrElem(inv_svcs, i)
		svc := Service{elem}
		svcs = append(svcs, svc)
	}

	btlog.Debugf("PeripheralDidModifyServices: prph=%v inv_svcs=%+v", prph, svcs)

	dlg := findPeripheralDlg(prph)
	if dlg != nil {
		dlg.DidModifyServices(Peripheral{prph}, svcs)
	}
}

//export BTPeripheralManagerDidUpdateState
func BTPeripheralManagerDidUpdateState(pmgr unsafe.Pointer) {
	btlog.Debugf("PeripheralManagerDidUpdateState: pmgr=%v", pmgr)

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.PeripheralManagerDidUpdateState(PeripheralManager{pmgr})
	}
}

//export BTPeripheralManagerWillRestoreState
func BTPeripheralManagerWillRestoreState(pmgr unsafe.Pointer, opts *C.struct_pmgr_restore_opts) {
	ropts := PeripheralManagerRestoreOpts{}

	if opts.svcs.count > 0 {
		ropts.Services = make([]MutableService, opts.svcs.count)
		for i, _ := range ropts.Services {
			ropts.Services[i] = MutableService{getObjArrElem(&opts.svcs, i)}
		}
	}

	if opts.adv_data != nil {
		ropts.AdvertisementData = &AdvData{
			LocalName:    C.GoString(opts.adv_data.name),
			ServiceUUIDs: mustStrArrToUUIDs(&opts.adv_data.svc_uuids),
		}
	}

	btlog.Debugf("PeripheralManagerWillRestoreState: pmgr=%v opts=%+v", pmgr, ropts)

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.PeripheralManagerWillRestoreState(PeripheralManager{pmgr}, ropts)
	}
}

//export BTPeripheralManagerDidAddService
func BTPeripheralManagerDidAddService(pmgr unsafe.Pointer, svc unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralManagerDidAddService: pmgr=%v err=%v", pmgr, nserr)

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.DidAddService(PeripheralManager{pmgr}, Service{svc}, nserr)
	}
}

//export BTPeripheralManagerDidStartAdvertising
func BTPeripheralManagerDidStartAdvertising(pmgr unsafe.Pointer, err *C.struct_bt_error) {
	nserr := btErrorToNSError(err)

	btlog.Debugf("PeripheralManagerDidStartAdvertising: pmgr=%v err=%+v", pmgr, nserr)

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.DidStartAdvertising(PeripheralManager{pmgr}, nserr)
	}
}

//export BTPeripheralManagerCentralDidSubscribe
func BTPeripheralManagerCentralDidSubscribe(pmgr unsafe.Pointer, cent unsafe.Pointer, chr unsafe.Pointer) {
	btlog.Debugf("PeripheralManagerCentralDidSubscribe: pmgr=%v cent=%v chr=%v", pmgr, cent, chr)

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.CentralDidSubscribe(PeripheralManager{pmgr}, Central{cent}, Characteristic{chr})
	}
}

//export BTPeripheralManagerCentralDidUnsubscribe
func BTPeripheralManagerCentralDidUnsubscribe(pmgr unsafe.Pointer, cent unsafe.Pointer, chr unsafe.Pointer) {
	btlog.Debugf("PeripheralManagerCentralDidUnsubscribe: pmgr=%v cent=%v chr=%v", pmgr, cent, chr)

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.CentralDidUnsubscribe(PeripheralManager{pmgr}, Central{cent}, Characteristic{chr})
	}
}

//export BTPeripheralManagerIsReadyToUpdateSubscribers
func BTPeripheralManagerIsReadyToUpdateSubscribers(pmgr unsafe.Pointer) {
	btlog.Debugf("PeripheralManagerIsReadyToUpdateSubscribers: pmgr=%v")

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.IsReadyToUpdateSubscribers(PeripheralManager{pmgr})
	}
}

//export BTPeripheralManagerDidReceiveReadRequest
func BTPeripheralManagerDidReceiveReadRequest(pmgr unsafe.Pointer, req unsafe.Pointer) {
	btlog.Debugf("PeripheralManagerDidReceiveReadRequest: pmgr=%v req=%+v", pmgr, req)

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.DidReceiveReadRequest(PeripheralManager{pmgr}, ATTRequest{req})
	}
}

//export BTPeripheralManagerDidReceiveWriteRequests
func BTPeripheralManagerDidReceiveWriteRequests(pmgr unsafe.Pointer, oa *C.struct_obj_arr) {
	reqs := make([]ATTRequest, oa.count)
	for i, _ := range reqs {
		ptr := getObjArrElem(oa, i)
		reqs[i] = ATTRequest{ptr}
	}

	btlog.Debugf("PeripheralManagerDidReceiveWriteRequests: pmgr=%v reqs=%v", pmgr, reqs)

	dlg := findPeripheralManagerDlg(pmgr)
	if dlg != nil {
		dlg.DidReceiveWriteRequests(PeripheralManager{pmgr}, reqs)
	}
}
