#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
#import <CoreLocation/CoreLocation.h>
#import "bt.h"

// pmgr.m: C functions for interfacing with CoreBluetooth peripheral
// functionality.  This is necessary because Go code cannot execute some
// objective C constructs directly.

CBPeripheralManager *
cb_alloc_pmgr(bool pwr_alert, const char *restore_id)
{
    // Ensure queue is initialized.
    bt_init();

    NSMutableDictionary *opts = [[NSMutableDictionary alloc] init];

    if (pwr_alert) {
        [opts setObject:[NSNumber numberWithBool:pwr_alert]
                 forKey:CBPeripheralManagerOptionShowPowerAlertKey];
    }
    if (restore_id != NULL) {
        [opts setObject:str_to_nsstring(restore_id)
                 forKey:CBPeripheralManagerOptionRestoreIdentifierKey];
    }

    CBPeripheralManager *pm = [[CBPeripheralManager alloc] initWithDelegate:bt_dlg
                                                                      queue:bt_queue
                                                                    options:opts];
    [pm retain];
    return pm;
}

void
cb_pmgr_set_delegate(void *pmgr, bool set)
{
    BTDlg *del;
    if (set) {
        del = bt_dlg;
    } else {
        del = nil;
    }

    ((CBPeripheralManager *)pmgr).delegate = del;
}

int
cb_pmgr_state(void *pmgr)
{
    CBPeripheralManager *pm = pmgr;
    return [pm state];
}

void
cb_pmgr_add_svc(void *pmgr, void *svc)
{
    [(CBPeripheralManager *)pmgr addService:svc];
}

void
cb_pmgr_remove_svc(void *pmgr, void *svc)
{
    [(CBPeripheralManager *)pmgr removeService:svc];
}

void
cb_pmgr_remove_all_svcs(void *pmgr)
{
    [(CBPeripheralManager *)pmgr removeAllServices];
}

void
cb_pmgr_start_adv(void *pmgr, const struct adv_data *ad)
{
    NSMutableDictionary *dict = [[NSMutableDictionary alloc] init];
    [dict autorelease];

    // iBeacon data is mutually exclusing with the rest of the fields.
    if (ad->ibeacon_data.length > 0) {
        NSData *nsd = byte_arr_to_nsdata(&ad->ibeacon_data);
        [nsd autorelease];

        [dict setObject:nsd forKey:@"kCBAdvDataAppleBeaconKey"];
    } else {
        if (ad->name != NULL) {
            NSString *nss = str_to_nsstring(ad->name);
            [nss autorelease];

            [dict setObject:nss
                     forKey:CBAdvertisementDataLocalNameKey];
        }

        if (ad->svc_uuids.count > 0) {
            NSArray *arr_svc_uuids = strs_to_cbuuids(&ad->svc_uuids);
            [arr_svc_uuids autorelease];

            [dict setObject:arr_svc_uuids
                     forKey:CBAdvertisementDataServiceUUIDsKey];
        }
    }

    [(CBPeripheralManager *)pmgr startAdvertising:dict];
}

void
cb_pmgr_stop_adv(void *pmgr)
{
    [(CBPeripheralManager *)pmgr stopAdvertising];
}

bool
cb_pmgr_is_adv(void *pmgr)
{
    return [(CBPeripheralManager *)pmgr isAdvertising];
}

bool
cb_pmgr_update_val(void *pmgr, const struct byte_arr *value, void *chr, const struct obj_arr *centrals)
{
    NSData *nsd = byte_arr_to_nsdata(value);
    [nsd autorelease];

    NSArray *nsa = obj_arr_to_nsarray(centrals);
    [nsa autorelease];

    return [(CBPeripheralManager *)pmgr updateValue:nsd
                                  forCharacteristic:chr
                               onSubscribedCentrals:nsa];
}

void
cb_pmgr_respond_to_req(void *pmgr, void *req, int result)
{
    [(CBPeripheralManager *)pmgr respondToRequest:req withResult:result];
}

void
cb_pmgr_set_conn_latency(void *pmgr, int latency, void *central)
{
    [(CBPeripheralManager *)pmgr setDesiredConnectionLatency:latency forCentral:central];
}

int
cb_cent_maximum_update_len(void *cent)
{
    return ((CBCentral *)cent).maximumUpdateValueLength;
}

CBMutableService *
cb_msvc_alloc(const char *uuid, bool primary)
{
    CBUUID *cbuuid = str_to_cbuuid(uuid);

    CBMutableService *svc = [[CBMutableService alloc] initWithType:cbuuid primary:primary];

    [svc retain];
    return svc;
}

void
cb_msvc_set_characteristics(void *msvc, const struct obj_arr *mchrs)
{
    NSArray *nsa = obj_arr_to_nsarray(mchrs);
    [nsa autorelease];

    ((CBMutableService *)msvc).characteristics = nsa;
}

void
cb_msvc_set_included_services(void *msvc, const struct obj_arr *msvcs)
{
    NSArray *nsa = obj_arr_to_nsarray(msvcs);
    [nsa autorelease];

    ((CBMutableService *)msvc).includedServices = nsa;
}

CBMutableCharacteristic *
cb_mchr_alloc(const char *uuid, int properties, const struct byte_arr *value, int permissions)
{
    CBUUID *cbuuid = str_to_cbuuid(uuid);
    NSData *nsd = byte_arr_to_nsdata(value);

    CBMutableCharacteristic *chr = [[CBMutableCharacteristic alloc] initWithType:cbuuid
                                                                      properties:properties
                                                                           value:nsd
                                                                     permissions:permissions];

    [chr retain];
    return chr;
}

void
cb_mchr_set_descriptors(void *mchr, const struct obj_arr *mdscs)
{
    NSArray *nsa = obj_arr_to_nsarray(mdscs);
    [nsa autorelease];

    ((CBMutableCharacteristic *)mchr).descriptors = nsa;
}

void
cb_mchr_set_value(void *mchr, const struct byte_arr *val)
{
    ((CBMutableCharacteristic *)mchr).value = byte_arr_to_nsdata(val);
}


CBMutableDescriptor *
cb_mdsc_alloc(const char *uuid, const struct byte_arr *value)
{
    CBUUID *cbuuid = str_to_cbuuid(uuid);
    NSData *nsd = byte_arr_to_nsdata(value);

    CBMutableDescriptor *dsc = [[CBMutableDescriptor alloc] initWithType:cbuuid
                                                                   value:nsd];

    [dsc retain];
    return dsc;
}

CBCentral *
cb_atr_central(void *atr)
{
    return ((CBATTRequest *)atr).central;
}

CBCharacteristic *
cb_atr_characteristic(void *atr)
{
    return ((CBATTRequest *)atr).characteristic;
}

struct byte_arr
cb_atr_value(void *atr)
{
    NSData *nsd = ((CBATTRequest *)atr).value;
    return nsdata_to_byte_arr(nsd);
}

void
cb_atr_set_value(void *atr, const struct byte_arr *ba)
{
    NSData *nsd = byte_arr_to_nsdata(ba);
    ((CBATTRequest *)atr).value = nsd;
}

int
cb_atr_offset(void *atr)
{
    return ((CBATTRequest *)atr).offset;
}
