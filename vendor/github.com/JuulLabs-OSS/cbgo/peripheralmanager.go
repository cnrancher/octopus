package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

import (
	"unsafe"
)

// PeripheralManagerConnectionLatency: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerconnectionlatency
type PeripheralManagerConnectionLatency int

const (
	PeripheralManagerConnectionLatencyLow    = PeripheralManagerConnectionLatency(C.CBPeripheralManagerConnectionLatencyLow)
	PeripheralManagerConnectionLatencyMedium = PeripheralManagerConnectionLatency(C.CBPeripheralManagerConnectionLatencyMedium)
	PeripheralManagerConnectionLatencyHigh   = PeripheralManagerConnectionLatency(C.CBPeripheralManagerConnectionLatencyHigh)
)

// PeripheralManagerRestoreOpts: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanagerdelegate/peripheral_manager_state_restoration_options
type PeripheralManagerRestoreOpts struct {
	Services          []MutableService
	AdvertisementData *AdvData
}

// PeripheralManager: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager
type PeripheralManager struct {
	ptr unsafe.Pointer
}

var pmgrPtrMap = newPtrMap()

func findPeripheralManagerDlg(ptr unsafe.Pointer) PeripheralManagerDelegate {
	itf := pmgrPtrMap.find(ptr)
	if itf == nil {
		return nil
	}

	return itf.(PeripheralManagerDelegate)
}

// NewPeripheralManager creates a peripheral manager.  Specify a nil `opts`
// value for defaults.  Don't forget to call `SetDelegate()` afterwards!
func NewPeripheralManager(opts *ManagerOpts) PeripheralManager {
	if opts == nil {
		opts = &DfltManagerOpts
	}

	pwrAlert := C.bool(opts.ShowPowerAlert)

	restoreID := (*C.char)(nil)
	if opts.RestoreIdentifier != "" {
		restoreID = C.CString(opts.RestoreIdentifier)
		defer C.free(unsafe.Pointer(restoreID))
	}

	return PeripheralManager{
		ptr: unsafe.Pointer(C.cb_alloc_pmgr(pwrAlert, restoreID)),
	}
}

// SetDelegate configures a receiver for a central manager's asynchronous
// callbacks.
func (pm PeripheralManager) SetDelegate(d PeripheralManagerDelegate) {
	if d != nil {
		pmgrPtrMap.add(pm.ptr, d)
	}
	C.cb_pmgr_set_delegate(pm.ptr, C.bool(d != nil))
}

// State: https://developer.apple.com/documentation/corebluetooth/cbmanager/1648600-state
func (pm PeripheralManager) State() ManagerState {
	return ManagerState(C.cb_pmgr_state(pm.ptr))
}

// AddService: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393255-addservice
func (pm PeripheralManager) AddService(svc MutableService) {
	C.cb_pmgr_add_svc(pm.ptr, svc.ptr)
}

// RemoveService: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393287-removeservice
func (pm PeripheralManager) RemoveService(svc MutableService) {
	C.cb_pmgr_remove_svc(pm.ptr, svc.ptr)
}

// RemoveAllServices: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393269-removeallservices
func (pm PeripheralManager) RemoveAllServices() {
	C.cb_pmgr_remove_all_svcs(pm.ptr)
}

// StartAdvertising: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393252-startadvertising
func (pm PeripheralManager) StartAdvertising(ad AdvData) {
	cad := C.struct_adv_data{}

	if len(ad.IBeaconData) > 0 {
		cad.ibeacon_data = byteSliceToByteArr(ad.IBeaconData)
		defer C.free(unsafe.Pointer(cad.ibeacon_data.data))
	} else {
		if len(ad.LocalName) > 0 {
			cad.name = C.CString(ad.LocalName)
			defer C.free(unsafe.Pointer(cad.name))
		}

		cad.svc_uuids = uuidsToStrArr(ad.ServiceUUIDs)
		defer freeStrArr(&cad.svc_uuids)
	}

	C.cb_pmgr_start_adv(pm.ptr, &cad)
}

// StopAdvertising: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393275-stopadvertising
func (pm PeripheralManager) StopAdvertising() {
	C.cb_pmgr_stop_adv(pm.ptr)
}

// IsAdvertising: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393291-isadvertising
func (pm PeripheralManager) IsAdvertising() bool {
	return bool(C.cb_pmgr_is_adv(pm.ptr))
}

// UpdateValue: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393281-updatevalue
func (pm PeripheralManager) UpdateValue(value []byte, chr Characteristic, centrals []Central) bool {
	ba := byteSliceToByteArr(value)
	defer C.free(unsafe.Pointer(ba.data))

	oa := mallocObjArr(len(centrals))
	defer C.free(unsafe.Pointer(oa.objs))

	for i, c := range centrals {
		setObjArrElem(&oa, i, c.ptr)
	}

	return bool(C.cb_pmgr_update_val(pm.ptr, &ba, chr.ptr, &oa))
}

// RespondToRequest: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393293-respondtorequest
func (pm PeripheralManager) RespondToRequest(req ATTRequest, result ATTError) {
	C.cb_pmgr_respond_to_req(pm.ptr, req.ptr, C.int(result))
}

// SetDesiredConnectionLatency: https://developer.apple.com/documentation/corebluetooth/cbperipheralmanager/1393277-setdesiredconnectionlatency
func (pm PeripheralManager) SetDesiredConnectionLatency(latency PeripheralManagerConnectionLatency, central Central) {
	C.cb_pmgr_set_conn_latency(pm.ptr, C.int(latency), central.ptr)
}
