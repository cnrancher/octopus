#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
#import <CoreLocation/CoreLocation.h>
#import "bt.h"

// cmgr.m: C functions for interfacing with CoreBluetooth central
// functionality.  This is necessary because Go code cannot execute some
// objective C constructs directly.

CBCentralManager *
cb_alloc_cmgr(bool pwr_alert, const char *restore_id)
{
    // Ensure queue is initialized.
    bt_init();

    NSMutableDictionary *opts = [[NSMutableDictionary alloc] init];
    [opts autorelease];

    if (pwr_alert) {
        [opts setObject:[NSNumber numberWithBool:pwr_alert]
                 forKey:CBCentralManagerOptionShowPowerAlertKey];
    }
    if (restore_id != NULL) {
        [opts setObject:str_to_nsstring(restore_id)
                 forKey:CBCentralManagerOptionRestoreIdentifierKey];
    }

    CBCentralManager *cm = [[CBCentralManager alloc] initWithDelegate:bt_dlg
                                                                queue:bt_queue
                                                              options:opts];
    [cm retain];
    return cm;
}

void
cb_cmgr_set_delegate(void *cmgr, bool set)
{
    BTDlg *del;
    if (set) {
        del = bt_dlg;
    } else {
        del = nil;
    }

    ((CBCentralManager *)cmgr).delegate = del;
}

int
cb_cmgr_state(void *cmgr)
{
    CBCentralManager *cm = cmgr;
    return [cm state];
}

void
cb_cmgr_scan(void *cmgr, const struct string_arr *svc_uuids,
             const struct scan_opts *opts)
{
    NSArray *arr_svc_uuids = strs_to_cbuuids(svc_uuids);
    [arr_svc_uuids autorelease];

    NSMutableDictionary *dict = [[NSMutableDictionary alloc] init];
    [dict autorelease];

    if (opts->allow_dups) {
        [dict setObject:[NSNumber numberWithBool:YES]
                 forKey:CBCentralManagerScanOptionAllowDuplicatesKey];
    }

    if (opts->sol_svc_uuids.count > 0) {
        NSArray *arr_sol_svc_uuids = strs_to_cbuuids(&opts->sol_svc_uuids);
        [arr_sol_svc_uuids autorelease];

        [dict setObject:arr_sol_svc_uuids
                 forKey:CBCentralManagerScanOptionSolicitedServiceUUIDsKey];
    }

    CBCentralManager *cm = cmgr;
    [cm scanForPeripheralsWithServices:arr_svc_uuids options:dict];
}

void
cb_cmgr_stop_scan(void *cmgr)
{
    [(CBCentralManager *)cmgr stopScan];
}

bool
cb_cmgr_is_scanning(void *cmgr)
{
    return ((CBCentralManager *)cmgr).isScanning;
}

void
cb_cmgr_connect(void *cmgr, void *prph, const struct connect_opts *opts)
{
    CBCentralManager *cm = cmgr;
    CBPeripheral *pr = prph;

    NSMutableDictionary *dict = [[NSMutableDictionary alloc] init];
    [dict autorelease];

    if (opts->notify_on_connection) {
        dict_set_bool(dict, CBConnectPeripheralOptionNotifyOnConnectionKey, true);
    }
    if (opts->notify_on_disconnection) {
        dict_set_bool(dict, CBConnectPeripheralOptionNotifyOnDisconnectionKey, true);
    }
    if (opts->notify_on_notification) {
        dict_set_bool(dict, CBConnectPeripheralOptionNotifyOnNotificationKey, true);
    }

    // XXX: These don't seem to be supported in all versions of macOS.
#if 0
    if (opts->enable_transport_bridging) {
        dict_set_bool(dict, CBConnectPeripheralOptionEnableTransportBridgingKey, true);
    }
    if (opts->requires_ancs) {
        dict_set_bool(dict, CBConnectPeripheralOptionRequiresANCS, true);
    }
    if (opts->start_delay > 0) {
        dict_set_int(dict, CBConnectPeripheralOptionStartDelayKey, opts->start_delay);
    }
#endif

    [cm connectPeripheral:pr options:dict];
}

void
cb_cmgr_cancel_connect(void *cmgr, void *prph)
{
    CBCentralManager *cm = cmgr;
    CBPeripheral *pr = prph;

    [cm cancelPeripheralConnection:pr];
}

struct obj_arr
cb_cmgr_retrieve_prphs_with_svcs(void *cmgr, const struct string_arr *uuids)
{
    CBCentralManager *cm = cmgr;

    NSArray *cbuuids = strs_to_cbuuids(uuids);
    NSArray *prphs = [cm retrieveConnectedPeripheralsWithServices:cbuuids];

    if ([prphs count] <= 0) {
        return (struct obj_arr) {0};
    }

    void **objs = malloc([prphs count] * sizeof *objs);
    assert(objs != NULL);

    for (int i = 0; i < [prphs count]; i++) {
        objs[i] = [prphs objectAtIndex:i];
    }

    return (struct obj_arr) {
        .objs = objs,
        .count = [prphs count],
    };
}

struct obj_arr
cb_cmgr_retrieve_prphs(void *cmgr, const struct string_arr *uuids)
{
    CBCentralManager *cm = cmgr;

    NSArray *nsuuids = strs_to_nsuuids(uuids);
    NSArray *prphs = [cm retrievePeripheralsWithIdentifiers:nsuuids];

    if ([prphs count] <= 0) {
        return (struct obj_arr) {0};
    }

    void **objs = malloc([prphs count] * sizeof *objs);
    assert(objs != NULL);

    for (int i = 0; i < [prphs count]; i++) {
        objs[i] = [prphs objectAtIndex:i];
    }

    return (struct obj_arr) {
        .objs = objs,
        .count = [prphs count],
    };
}

const char *
cb_peer_identifier(void *peer)
{
    NSUUID *uuid = [(CBPeer *)peer identifier];
    return [[uuid UUIDString] UTF8String];
}

void
cb_prph_set_delegate(void *prph, bool set)
{
    BTDlg *del;
    if (set) {
        del = bt_dlg;
    } else {
        del = nil;
    }

    ((CBPeripheral *)prph).delegate = del;
}

const char *
cb_prph_name(void *prph)
{
    NSString *nss = [(CBPeripheral *)prph name];
    return [nss UTF8String];
}

struct obj_arr
cb_prph_services(void *prph)
{
    NSArray *svcs = ((CBPeripheral *)prph).services;
    return nsarray_to_obj_arr(svcs);
}

void
cb_prph_discover_svcs(void *prph, const struct string_arr *svc_uuid_strs)
{
    NSArray *svc_uuids = strs_to_cbuuids(svc_uuid_strs);
    [(CBPeripheral *)prph discoverServices:svc_uuids];
}

void
cb_prph_discover_included_svcs(void *prph, const struct string_arr *svc_uuid_strs, void *svc)
{
    NSArray *svc_uuids = strs_to_cbuuids(svc_uuid_strs);
    [(CBPeripheral *)prph discoverIncludedServices:svc_uuids forService:svc];
}

void
cb_prph_discover_chrs(void *prph, void *svc, const struct string_arr *chr_uuid_strs)
{
    NSArray *chr_uuids = strs_to_cbuuids(chr_uuid_strs);
    [(CBPeripheral *)prph discoverCharacteristics:chr_uuids forService:svc];
}

void
cb_prph_discover_dscs(void *prph, void *chr)
{
    [(CBPeripheral *)prph discoverDescriptorsForCharacteristic:chr];
}

void
cb_prph_read_chr(void *prph, void *chr)
{
    [(CBPeripheral *)prph readValueForCharacteristic:chr];
}

void
cb_prph_read_dsc(void *prph, void *dsc)
{
    [(CBPeripheral *)prph readValueForDescriptor:dsc];
}

void
cb_prph_write_chr(void *prph, void *chr, struct byte_arr *value, int type)
{
    NSData *nsd = byte_arr_to_nsdata(value);
    [(CBPeripheral *)prph writeValue:nsd
                   forCharacteristic:chr
                                type:type];
}

void
cb_prph_write_dsc(void *prph, void *dsc, struct byte_arr *value)
{
    NSData *nsd = byte_arr_to_nsdata(value);
    [(CBPeripheral *)prph writeValue:nsd
                       forDescriptor:dsc];
}

int
cb_prph_max_write_len(void *prph, int type)
{
    return [(CBPeripheral *)prph maximumWriteValueLengthForType:type];
}

void
cb_prph_set_notify(void *prph, bool enabled, void *chr)
{
    [(CBPeripheral *)prph setNotifyValue:enabled forCharacteristic:chr];
}

int
cb_prph_state(void *prph)
{
    return ((CBPeripheral *) prph).state;
}

bool
cb_prph_can_send_write_without_rsp(void *prph)
{
    return ((CBPeripheral *)prph).canSendWriteWithoutResponse;
}

void
cb_prph_read_rssi(void *prph)
{
    [(CBPeripheral *)prph readRSSI];
}

// error: 'openL2CAPChannel:' is unavailable: not available on macOS
#if 0
void
cb_prph_open_l2cap_channel(void *prph, uint16_t cid)
{
    [(CBPeripheral *)prph openL2CAPChannel:cid];
}
#endif

// error: property 'ancsAuthorized' not found on object of type 'CBPeripheral *'
#if 0
bool
cb_prph_ancs_authorized(void *prph)
{
    return ((CBPeripheral *)prph).ancsAuthorized;
}
#endif

const char *
cb_svc_uuid(void *svc)
{
    CBUUID *cbuuid = ((CBService *)svc).UUID;
    return [[cbuuid UUIDString] UTF8String];
}

void *
cb_svc_peripheral(void *svc)
{
    return ((CBService *)svc).peripheral;
}

bool
cb_svc_is_primary(void *svc)
{
    return ((CBService *)svc).isPrimary;
}

struct obj_arr
cb_svc_characteristics(void *svc)
{
    NSArray *chrs = ((CBService *)svc).characteristics;
    return nsarray_to_obj_arr(chrs);
}

struct obj_arr
cb_svc_included_svcs(void *svc)
{
    NSArray *svcs = ((CBService *)svc).includedServices;
    return nsarray_to_obj_arr(svcs);
}

const char *
cb_chr_uuid(void *chr)
{
    CBUUID *cbuuid = ((CBCharacteristic *)chr).UUID;
    return [[cbuuid UUIDString] UTF8String];
}

void *
cb_chr_service(void *chr)
{
    return ((CBCharacteristic *)chr).service;
}

struct obj_arr
cb_chr_descriptors(void *chr)
{
    NSArray *dscs = ((CBCharacteristic *)chr).descriptors;
    return nsarray_to_obj_arr(dscs);
}

struct byte_arr
cb_chr_value(void *chr)
{
    NSData *nsd = ((CBCharacteristic *)chr).value;
    return nsdata_to_byte_arr(nsd);
}

int
cb_chr_properties(void *chr)
{
    return ((CBCharacteristic *)chr).properties;
}

bool
cb_chr_is_notifying(void *chr)
{
    return ((CBCharacteristic *)chr).isNotifying;
}

const char *
cb_dsc_uuid(void *dsc)
{
    CBUUID *cbuuid = ((CBDescriptor *)dsc).UUID;
    return [[cbuuid UUIDString] UTF8String];
}

void *
cb_dsc_characteristic(void *dsc)
{
    return ((CBDescriptor *)dsc).characteristic;
}

struct byte_arr
cb_dsc_value(void *dsc)
{
    id val = ((CBDescriptor *)dsc).value;

    // `value` returns a different type depending on the descriptor being
    // queried!  Convert whatever got returned to a byte array.

    if ([val isKindOfClass:[NSData class]]) {
        return nsdata_to_byte_arr(val);
    }

    if ([val isKindOfClass:[NSString class]]) {
        NSData *nsd = [val dataUsingEncoding:NSUTF8StringEncoding];
        return nsdata_to_byte_arr(nsd);
    }

    // Unknown type.
    return (struct byte_arr){0};
}
