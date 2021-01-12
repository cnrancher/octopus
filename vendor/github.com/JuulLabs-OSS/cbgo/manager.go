package cbgo

/*
// See cutil.go for C compiler flags.
#import "bt.h"
*/
import "C"

// ManagerState: https://developer.apple.com/documentation/corebluetooth/cbmanagerstate
type ManagerState int

const (
	ManagerStatePoweredOff   = ManagerState(C.CBManagerStatePoweredOff)
	ManagerStatePoweredOn    = ManagerState(C.CBManagerStatePoweredOn)
	ManagerStateResetting    = ManagerState(C.CBManagerStateResetting)
	ManagerStateUnauthorized = ManagerState(C.CBManagerStateUnauthorized)
	ManagerStateUnknown      = ManagerState(C.CBManagerStateUnknown)
	ManagerStateUnsupported  = ManagerState(C.CBManagerStateUnsupported)
)

// ManagerOpts: https://developer.apple.com/documentation/corebluetooth/cbcentralmanager/central_manager_initialization_options
type ManagerOpts struct {
	ShowPowerAlert    bool
	RestoreIdentifier string
}

// DfltManagerOpts is the set of options that gets used when nil is
// passed to `NewCentralManager()`.
var DfltManagerOpts = ManagerOpts{
	ShowPowerAlert:    false,
	RestoreIdentifier: "",
}
