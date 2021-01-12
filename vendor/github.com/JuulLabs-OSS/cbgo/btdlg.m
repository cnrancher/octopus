#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
#import "bt.h"

@implementation BTDlg

- (id)
init
{
    self = [super init];
    if (self == nil) {
        return nil;
    }

    return self;
}

/**
 * Called when the central manager successfully connects to a peripheral.
 */
- (void)
      centralManager:(CBCentralManager *)cm
didConnectPeripheral:(CBPeripheral *)prph
{
    prph.delegate = self;
    BTCentralManagerDidConnectPeripheral(cm, prph);
}

/**
 * Called when a connection to a preipheral is terminated.
 */
- (void)
         centralManager:(CBCentralManager *)cm
didDisconnectPeripheral:(CBPeripheral *)prph
                  error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTCentralManagerDidDisconnectPeripheral(cm, prph, &err);
}

/**
 * Called when the central manager fails to connect to a peripheral.
 */
- (void)
            centralManager:(CBCentralManager *)cm
didFailToConnectPeripheral:(CBPeripheral *)prph
                     error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTCentralManagerDidFailToConnectPeripheral(cm, prph, &err);
}

// macOS 10.15+
#if 0
- (void)
         centralManager:(CBCentralManager *)cm 
connectionEventDidOccur:(CBConnectionEvent)event 
          forPeripheral:(CBPeripheral *)prph
{
    BTCentralManagerConnectionEventDidOccur(cm, event, prph);
}
#endif

/**
 * Called when the central manager discovers a peripheral while scanning for
 * devices.
 */
- (void)
       centralManager:(CBCentralManager *)cm
didDiscoverPeripheral:(CBPeripheral *)prph
    advertisementData:(NSDictionary *)advData
                 RSSI:(NSNumber *)RSSI
{
    struct adv_fields af = {0};

    af.name = dict_string(advData, CBAdvertisementDataLocalNameKey);
    af.mfg_data = dict_bytes(advData, CBAdvertisementDataManufacturerDataKey);
    af.pwr_lvl = dict_int(advData, CBAdvertisementDataTxPowerLevelKey, ADV_FIELDS_PWR_LVL_NONE);
    af.connectable = dict_int(advData, CBAdvertisementDataIsConnectable, ADV_FIELDS_CONNECTABLE_NONE);

    const NSArray *arr = [advData objectForKey:CBAdvertisementDataServiceUUIDsKey];
    const char *svc_uuids[[arr count]];
    for (int i = 0; i < [arr count]; i++) {
        const CBUUID *uuid = [arr objectAtIndex:i];
        svc_uuids[i] = [[uuid UUIDString] UTF8String];
    }
    af.svc_uuids = (struct string_arr) {
        .strings = svc_uuids,
        .count = [arr count],
    };

    const NSDictionary *dict = [advData objectForKey:CBAdvertisementDataServiceDataKey];
    const NSArray *keys = [dict allKeys];

    const char *svc_data_uuids[[keys count]];
    struct byte_arr svc_data_values[[keys count]];

    for (int i = 0; i < [keys count]; i++) {
        const CBUUID *uuid = [keys objectAtIndex:i];
        svc_data_uuids[i] = [[uuid UUIDString] UTF8String];

        const NSData *data = [dict objectForKey:uuid];
        svc_data_values[i].data = [data bytes];
        svc_data_values[i].length = [data length];
    }
    af.svc_data_uuids = (struct string_arr) {
        .strings = svc_data_uuids,
        .count = [keys count],
    };
    af.svc_data_values = svc_data_values;

    prph.delegate = self;
    [prph retain];

    BTCentralManagerDidDiscoverPeripheral(cm, prph, &af, [RSSI intValue]);
}

/**
 * Called whenever the central manager's state is updated.
 */
- (void)
centralManagerDidUpdateState:(CBCentralManager *)cm
{
    BTCentralManagerDidUpdateState(cm);
}

/**
 * Called when the central manager is about to restore its state (as requested
 * by the application).
 */
- (void)
  centralManager:(CBCentralManager *)cm 
willRestoreState:(NSDictionary<NSString *,id> *)dict
{
    struct cmgr_restore_opts opts = {0};

    const NSArray *prphs = [dict objectForKey:CBCentralManagerRestoredStatePeripheralsKey];
    opts.prphs = nsarray_to_obj_arr(prphs);

    const NSArray *uuids = [dict objectForKey:CBCentralManagerRestoredStateScanServicesKey];
    opts.scan_svcs = cbuuids_to_strs(uuids);

    struct scan_opts scan_opts = {0};
    const NSDictionary *scan_dict = [dict objectForKey:CBCentralManagerRestoredStateScanOptionsKey];
    if (scan_dict != nil && [scan_dict count] > 0) {
        opts.scan_opts = &scan_opts;

        NSNumber *dups = [scan_dict objectForKey:CBCentralManagerScanOptionAllowDuplicatesKey];
        if (dups != nil && [dups boolValue]) {
            opts.scan_opts->allow_dups = true;
        }

        NSArray *sol_uuids = [scan_dict objectForKey:CBCentralManagerScanOptionSolicitedServiceUUIDsKey];
        opts.scan_opts->sol_svc_uuids = cbuuids_to_strs(sol_uuids);
    }

    BTCentralManagerWillRestoreState(cm, &opts);

    free(scan_opts.sol_svc_uuids.strings);
    free(opts.scan_svcs.strings);
    free(opts.prphs.objs);
}

/**
 * Called when the central manager successfully discovers services on a
 * peripheral.
 */
- (void) peripheral:(CBPeripheral *)prph
didDiscoverServices:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidDiscoverServices(prph, &err);
}

/**
 * Called when the central manager successfully discovers included services of
 * a peripheral's primary service.
 */
- (void)                   peripheral:(CBPeripheral *)prph
didDiscoverIncludedServicesForService:(CBService *)svc
                                error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidDiscoverIncludedServices(prph, svc, &err);
}

/**
 * Called when the central manager successfully discovers characteristics of a
 * peripheral's service.
 */
- (void)
                          peripheral:(CBPeripheral *)prph 
didDiscoverCharacteristicsForService:(CBService *)svc 
                               error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidDiscoverCharacteristics(prph, svc, &err);
}

/**
 * Called when the central manager successfully discovers descriptors of a
 * peripheral's characteristic.
 */
- (void)
                             peripheral:(CBPeripheral *)prph 
didDiscoverDescriptorsForCharacteristic:(CBCharacteristic *)chr 
             error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidDiscoverDescriptors(prph, chr, &err);
}

/**
 * Called when a connected peripheral communicates a characteristic's value
 * (via read response, notification, or indication).
 */
- (void)
                     peripheral:(CBPeripheral *)prph 
didUpdateValueForCharacteristic:(CBCharacteristic *)chr 
                          error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidUpdateValueForCharacteristic(prph, chr, &err);
}

/**
 * Called when a connected peripheral communicates a descriptors's value
 * (via read response).
 */
- (void)
                 peripheral:(CBPeripheral *)prph 
didUpdateValueForDescriptor:(CBDescriptor *)dsc 
                      error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidUpdateValueForDescriptor(prph, dsc, &err);
}

/**
 * Called when a connected peripheral responds to a Characteristic Write
 * Request.
 */
- (void)
                    peripheral:(CBPeripheral *)prph 
didWriteValueForCharacteristic:(CBCharacteristic *)chr 
                         error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidWriteValueForCharacteristic(prph, chr, &err);
}

/**
 * Called when a connected peripheral responds to a Descriptor Write Request.
 */
- (void)
                peripheral:(CBPeripheral *)prph 
didWriteValueForDescriptor:(CBDescriptor *)dsc 
                     error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidWriteValueForDescriptor(prph, dsc, &err);
}

/**
 * Called when a previously unsuccessful Write Without Response request can be
 * retried.
 */
- (void)
peripheralIsReadyToSendWriteWithoutResponse:(CBPeripheral *)prph
{
    BTPeripheralIsReadyToSendWriteWithoutResponse(prph);
}

/**
 * Called when we (un)subscribe to a connected peripheral's characteristic.
 */
- (void)
                                 peripheral:(CBPeripheral *)prph 
didUpdateNotificationStateForCharacteristic:(CBCharacteristic *)chr 
                                      error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidUpdateNotificationState(prph, chr, &err);
}

/**
 * Called when we read a connected peripheral's RSSI.
 */
- (void)
 peripheral:(CBPeripheral *)prph 
didReadRSSI:(NSNumber *)RSSI 
      error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralDidReadRSSI(prph, [RSSI intValue], &err);
}

/**
 * Called when a connected peripheral changes its name.
 */
- (void)
peripheralDidUpdateName:(CBPeripheral *)prph
{
    BTPeripheralDidUpdateName(prph);
}

/**
 * Called when a connected peripheral changes its set of services.
 */
- (void)
       peripheral:(CBPeripheral *)prph 
didModifyServices:(NSArray<CBService *> *)invSvcs
{
    struct obj_arr oa = nsarray_to_obj_arr(invSvcs);
    BTPeripheralDidModifyServices(prph, &oa);
    free(oa.objs);
}


/**
 * Called whenever the peripheral manager's state is updated.
 */
- (void)
peripheralManagerDidUpdateState:(CBPeripheralManager *)pm
{
    BTPeripheralManagerDidUpdateState(pm);
}

/**
 * Called when the peripheral manager is about to restore its state (as
 * requested by the application).
 */
- (void)
peripheralManager:(CBPeripheralManager *)pmgr 
 willRestoreState:(NSDictionary<NSString *,id> *)dict
{
    struct pmgr_restore_opts opts = {0};

    const NSArray *svcs = [dict objectForKey:CBPeripheralManagerRestoredStateServicesKey];
    opts.svcs = nsarray_to_obj_arr(svcs);

    struct adv_data adv_data = {0};
    const NSDictionary *ad_dict = [dict objectForKey:CBPeripheralManagerRestoredStateAdvertisementDataKey];
    if (ad_dict != nil && [dict count] > 0) {
        opts.adv_data = &adv_data;

        NSString *nss = [ad_dict objectForKey:CBAdvertisementDataLocalNameKey];
        if (nss != nil) {
            adv_data.name = [nss UTF8String];
        }

        NSArray *nsa = [ad_dict objectForKey:CBAdvertisementDataServiceUUIDsKey];
        if (nsa != nil) {
            adv_data.svc_uuids = cbuuids_to_strs(nsa);
        }
    }

    BTPeripheralManagerWillRestoreState(pmgr, &opts);

    free(adv_data.svc_uuids.strings);
    free(opts.svcs.objs);
}

/**
 * Called when an attempt to register a service has completed.
 */
- (void)
peripheralManager:(CBPeripheralManager *)pmgr 
    didAddService:(CBService *)svc 
            error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralManagerDidAddService(pmgr, svc, &err);
}

/**
 * Called when we start advertising or fail to advertise.
 */
- (void)
peripheralManagerDidStartAdvertising:(CBPeripheralManager *)pmgr 
                               error:(NSError *)nserr
{
    struct bt_error err = nserror_to_bt_error(nserr);
    BTPeripheralManagerDidStartAdvertising(pmgr, &err);
}

/**
 * Called when a connected central has subscribed to one of our
 * characteristics.
 */
- (void)
           peripheralManager:(CBPeripheralManager *)pmgr 
                     central:(CBCentral *)cent 
didSubscribeToCharacteristic:(CBCharacteristic *)chr
{
    BTPeripheralManagerCentralDidSubscribe(pmgr, cent, chr);
}

/**
 * Called when a connected central has unsubscribed from one of our
 * characteristics.
 */
- (void)
               peripheralManager:(CBPeripheralManager *)pmgr 
                         central:(CBCentral *)cent 
didUnsubscribeFromCharacteristic:(CBCharacteristic *)chr
{
    BTPeripheralManagerCentralDidUnsubscribe(pmgr, cent, chr);
}

/**
 * Called when a previous unsuccessful attempt to send notifications can be
 * retried.
 */
- (void)
peripheralManagerIsReadyToUpdateSubscribers:(CBPeripheralManager *)pmgr
{
    BTPeripheralManagerIsReadyToUpdateSubscribers(pmgr);
}

/**
 * Called when a connected cerntral has sent us a read request.
 */
- (void)
    peripheralManager:(CBPeripheralManager *)pmgr 
didReceiveReadRequest:(CBATTRequest *)req
{
    BTPeripheralManagerDidReceiveReadRequest(pmgr, req);
}

/**
 * Called when connected cerntrals have sent us write requests.
 */
- (void)
      peripheralManager:(CBPeripheralManager *)pmgr 
didReceiveWriteRequests:(NSArray *)reqs
{
    struct obj_arr oa = nsarray_to_obj_arr(reqs);

    BTPeripheralManagerDidReceiveWriteRequests(pmgr, &oa);
    free(oa.objs);
}

@end
