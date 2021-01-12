package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

import (
	"unsafe"
)

// CentralManagerRestoreOpts: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/central_manager_state_restoration_options
type CentralManagerRestoreOpts struct {
	Peripherals            []Peripheral
	ScanServices           []UUID
	CentralManagerScanOpts *CentralManagerScanOpts // nil if none
}

// CentralManagerScanOpts: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/peripheral_scanning_options
type CentralManagerScanOpts struct {
	AllowDuplicates       bool
	SolicitedServiceUUIDs []UUID
}

// DfltCentralManagerConnectOpts is the set of options that gets used when nil
// is passed to `Connect()`.
var DfltCentralManagerConnectOpts = CentralManagerConnectOpts{
	NotifyOnConnection:    true,
	NotifyOnDisconnection: true,
	NotifyOnNotification:  true,
}

// DfltCentralManagerScanOpts is the set of options that gets used when nil is
// passed to `Scan()`.
var DfltCentralManagerScanOpts = CentralManagerScanOpts{}

// CentralManagerConnectOpts: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/peripheral_connection_options
type CentralManagerConnectOpts struct {
	NotifyOnConnection      bool
	NotifyOnDisconnection   bool
	NotifyOnNotification    bool
	EnableTransportBridging bool
	RequiresANCS            bool
	StartDelay              int
}

// CentralManager: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager?language=objc
type CentralManager struct {
	ptr unsafe.Pointer
}

var cmgrPtrMap = newPtrMap()

func findCentralManagerDlg(ptr unsafe.Pointer) CentralManagerDelegate {
	itf := cmgrPtrMap.find(ptr)
	if itf == nil {
		return nil
	}

	return itf.(CentralManagerDelegate)
}

// NewCentralManager creates a central manager.  Specify a nil `opts` value for
// defaults.  Don't forget to call `SetDelegate()` afterwards!
func NewCentralManager(opts *ManagerOpts) CentralManager {
	if opts == nil {
		opts = &DfltManagerOpts
	}

	pwrAlert := C.bool(opts.ShowPowerAlert)

	restoreID := (*C.char)(nil)
	if opts.RestoreIdentifier != "" {
		restoreID = C.CString(opts.RestoreIdentifier)
		defer C.free(unsafe.Pointer(restoreID))
	}

	return CentralManager{
		ptr: unsafe.Pointer(C.cb_alloc_cmgr(pwrAlert, restoreID)),
	}
}

// SetDelegate configures a receiver for a central manager's asynchronous
// callbacks.
func (cm CentralManager) SetDelegate(d CentralManagerDelegate) {
	if d != nil {
		cmgrPtrMap.add(cm.ptr, d)
	}
	C.cb_cmgr_set_delegate(cm.ptr, C.bool(d != nil))
}

// State: https://developer.apple.com/documentation/corebluetooth/cbmanager/1648600-state
func (cm CentralManager) State() ManagerState {
	return ManagerState(C.cb_cmgr_state(cm.ptr))
}

// Connect: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/1518766-connectperipheral
func (cm CentralManager) Connect(prph Peripheral, opts *CentralManagerConnectOpts) {
	if opts == nil {
		opts = &DfltCentralManagerConnectOpts
	}

	copts := C.struct_connect_opts{
		notify_on_connection:      C.bool(opts.NotifyOnConnection),
		notify_on_disconnection:   C.bool(opts.NotifyOnDisconnection),
		notify_on_notification:    C.bool(opts.NotifyOnNotification),
		enable_transport_bridging: C.bool(opts.EnableTransportBridging),
		requires_ancs:             C.bool(opts.RequiresANCS),
		start_delay:               C.int(opts.StartDelay),
	}

	C.cb_cmgr_connect(cm.ptr, prph.ptr, &copts)
}

// CancelConnect: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/1518952-cancelperipheralconnection
func (cm CentralManager) CancelConnect(prph Peripheral) {
	C.cb_cmgr_cancel_connect(cm.ptr, prph.ptr)
}

// RetrieveConnectedPeripheralsWithServices: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/1518924-retrieveconnectedperipheralswith
func (cm CentralManager) RetrieveConnectedPeripheralsWithServices(uuids []UUID) []Peripheral {
	strs := uuidsToStrArr(uuids)
	defer freeStrArr(&strs)

	var prphs []Peripheral

	prphPtrs := C.cb_cmgr_retrieve_prphs_with_svcs(cm.ptr, &strs)
	defer C.free(unsafe.Pointer(prphPtrs.objs))

	for i := 0; i < int(prphPtrs.count); i++ {
		ptr := getObjArrElem(&prphPtrs, i)
		prphs = append(prphs, Peripheral{ptr})
	}

	return prphs
}

// RetrievePeripheralsWithIdentifiers: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/1519127-retrieveperipheralswithidentifie
func (cm CentralManager) RetrievePeripheralsWithIdentifiers(uuids []UUID) []Peripheral {
	strs := uuidsToStrArr(uuids)
	defer freeStrArr(&strs)

	var prphs []Peripheral

	prphPtrs := C.cb_cmgr_retrieve_prphs(cm.ptr, &strs)
	defer C.free(unsafe.Pointer(prphPtrs.objs))

	for i := 0; i < int(prphPtrs.count); i++ {
		ptr := getObjArrElem(&prphPtrs, i)
		prphs = append(prphs, Peripheral{ptr})
	}

	return prphs
}

// Scan: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/1518986-scanforperipheralswithservices
func (cm CentralManager) Scan(serviceUUIDs []UUID, opts *CentralManagerScanOpts) {
	arrSvcUUIDs := uuidsToStrArr(serviceUUIDs)
	defer freeStrArr(&arrSvcUUIDs)

	if opts == nil {
		opts = &DfltCentralManagerScanOpts
	}

	arrSolSvcUUIDs := uuidsToStrArr(opts.SolicitedServiceUUIDs)
	defer freeStrArr(&arrSolSvcUUIDs)

	copts := C.struct_scan_opts{
		allow_dups:    C.bool(opts.AllowDuplicates),
		sol_svc_uuids: arrSolSvcUUIDs,
	}

	C.cb_cmgr_scan(cm.ptr, &arrSvcUUIDs, &copts)
}

// StopScan: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/1518984-stopscan
func (cm CentralManager) StopScan() {
	C.cb_cmgr_stop_scan(cm.ptr)
}

// IsScanning: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/1620640-isscanning
func (cm CentralManager) IsScanning() bool {
	return bool(C.cb_cmgr_is_scanning(cm.ptr))
}
