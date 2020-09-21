#ifndef H_BT_
#define H_BT_

#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>

#define ADV_FIELDS_PWR_LVL_NONE         (-128)
#define ADV_FIELDS_CONNECTABLE_NONE     (-1)

struct byte_arr {
    const uint8_t *data;
    int length;
};

struct string_arr {
    const char **strings;
    int count;
};

struct obj_arr {
    void **objs;
    int count;
};

struct bt_error {
    const char *msg;
    int code;
};

struct adv_fields {
	const char *name;
	struct byte_arr mfg_data;
    struct string_arr svc_uuids;
    struct string_arr overflow_svc_uuids;
    int8_t pwr_lvl; // -128 = not present
	int connectable; // -1 = not present
    struct string_arr svc_data_uuids;
    struct byte_arr *svc_data_values;
};

struct adv_data {
    const char *name;
    struct string_arr svc_uuids;
    struct byte_arr ibeacon_data;
};

struct scan_opts {
    bool allow_dups;
    struct string_arr sol_svc_uuids;
};

struct connect_opts {
	bool notify_on_connection;
	bool notify_on_disconnection;
	bool notify_on_notification;
	bool enable_transport_bridging;
	bool requires_ancs;
	int start_delay;
};

struct cmgr_restore_opts {
    struct obj_arr prphs;
    struct string_arr scan_svcs;
    struct scan_opts *scan_opts;
};

struct pmgr_restore_opts {
    struct obj_arr svcs;
    struct adv_data *adv_data;
};

@interface BTDlg : NSObject <CBCentralManagerDelegate, CBPeripheralManagerDelegate, CBPeripheralDelegate>
{
}
@end

// bt.m
bool bt_start();
void bt_stop();
void bt_init();

// util.m
struct byte_arr nsdata_to_byte_arr(const NSData *nsdata);
NSData *byte_arr_to_nsdata(const struct byte_arr *ba);
struct obj_arr nsarray_to_obj_arr(const NSArray *arr);
NSArray *obj_arr_to_nsarray(const struct obj_arr *oa);
NSString *str_to_nsstring(const char *s);
struct bt_error nserror_to_bt_error(const NSError *err);
struct string_arr cbuuids_to_strs(const NSArray *cbuuids);
int dict_int(NSDictionary *dict, NSString *key, int dflt);
const char *dict_string(NSDictionary *dict, NSString *key);
const void *dict_data(NSDictionary *dict, NSString *key, int *out_len);
const struct byte_arr dict_bytes(NSDictionary *dict, NSString *key);
void dict_set_bool(NSMutableDictionary *dict, NSString *key, bool val);
void dict_set_int(NSMutableDictionary *dict, NSString *key, int val);
NSUUID *str_to_nsuuid(const char *s);
CBUUID *str_to_cbuuid(const char *s);
NSArray *strs_to_nsuuids(const struct string_arr *sa);
NSArray *strs_to_cbuuids(const struct string_arr *sa);
NSArray *strs_to_nsstrings(const struct string_arr *sa);

// cmgr.m
CBCentralManager *cb_alloc_cmgr(bool pwr_alert, const char *restore_id);
void cb_cmgr_set_delegate(void *cmgr, bool set);
int cb_cmgr_state(void *cm);
void cb_cmgr_scan(void *cmgr, const struct string_arr *svc_uuids,
                  const struct scan_opts *opts);
void cb_cmgr_stop_scan(void *cm);
bool cb_cmgr_is_scanning(void *cm);
void cb_cmgr_connect(void *cmgr, void *prph, const struct connect_opts *opts);
void cb_cmgr_cancel_connect(void *cmgr, void *prph);
struct obj_arr cb_cmgr_retrieve_prphs_with_svcs(void *cmgr, const struct string_arr *svc_uuids);
struct obj_arr cb_cmgr_retrieve_prphs(void *cmgr, const struct string_arr *uuids);

const char *cb_peer_identifier(void *prph);

void cb_prph_set_delegate(void *prph, bool set);
const char *cb_prph_name(void *prph);
struct obj_arr cb_prph_services(void *prph);
void cb_prph_discover_svcs(void *prph, const struct string_arr *svc_uuid_strs);
void cb_prph_discover_included_svcs(void *prph, const struct string_arr *svc_uuid_strs, void *svc);
void cb_prph_discover_chrs(void *prph, void *svc, const struct string_arr *chr_uuid_strs);
void cb_prph_discover_dscs(void *prph, void *chr);
void cb_prph_read_chr(void *prph, void *chr);
void cb_prph_read_dsc(void *prph, void *dsc);
void cb_prph_write_chr(void *prph, void *chr, struct byte_arr *value, int type);
void cb_prph_write_dsc(void *prph, void *dsc, struct byte_arr *value);
int cb_prph_max_write_len(void *prph, int type);
void cb_prph_set_notify(void *prph, bool enabled, void *chr);
int cb_prph_state(void *prph);
bool cb_prph_can_send_write_without_rsp(void *prph);
void cb_prph_read_rssi(void *prph);
bool cb_prph_ancs_authorized(void *prph);

const char *cb_svc_uuid(void *svc);
void *cb_svc_peripheral(void *svc);
bool cb_svc_is_primary(void *svc);
struct obj_arr cb_svc_characteristics(void *svc);
struct obj_arr cb_svc_included_svcs(void *svc);

const char *cb_chr_uuid(void *chr);
void *cb_chr_service(void *chr);
struct obj_arr cb_chr_descriptors(void *chr);
struct byte_arr cb_chr_value(void *chr);
int cb_chr_properties(void *chr);
bool cb_chr_is_notifying(void *chr);

const char *cb_dsc_uuid(void *dsc);
void *cb_dsc_characteristic(void *dsc);
struct byte_arr cb_dsc_value(void *dsc);

// pmgr.m
CBPeripheralManager *cb_alloc_pmgr(bool pwr_alert, const char *restore_id);
void cb_pmgr_set_delegate(void *pmgr, bool set);
int cb_pmgr_state(void *pmgr);
void cb_pmgr_add_svc(void *pmgr, void *svc);
void cb_pmgr_remove_svc(void *pmgr, void *svc);
void cb_pmgr_remove_all_svcs(void *pmgr);
void cb_pmgr_start_adv(void *pmgr, const struct adv_data *ad);
void cb_pmgr_stop_adv(void *pmgr);
bool cb_pmgr_is_adv(void *pmgr);
bool cb_pmgr_update_val(void *pmgr, const struct byte_arr *value, void *chr, const struct obj_arr *centrals);
void cb_pmgr_respond_to_req(void *pmgr, void *req, int result);
void cb_pmgr_set_conn_latency(void *pmgr, int latency, void *central);

int cb_cent_maximum_update_len(void *cent);

CBMutableService *cb_msvc_alloc(const char *uuid, bool primary);
void cb_msvc_set_characteristics(void *msvc, const struct obj_arr *mchrs);
void cb_msvc_set_included_services(void *msvc, const struct obj_arr *msvcs);

CBMutableCharacteristic *cb_mchr_alloc(const char *uuid, int properties, const struct byte_arr *value,
                                       int permissions);
void cb_mchr_set_value(void *mchr, const struct byte_arr *val);
void cb_mchr_set_descriptors(void *mchr, const struct obj_arr *mdscs);

CBMutableDescriptor *cb_mdsc_alloc(const char *uuid, const struct byte_arr *value);

CBCentral *cb_atr_central(void *atr);
CBCharacteristic *cb_atr_characteristic(void *atr);
struct byte_arr cb_atr_value(void *atr);
void cb_atr_set_value(void *atr, const struct byte_arr *ba);
int cb_atr_offset(void *atr);

// cbhandlers.go
void BTCentralManagerDidConnectPeripheral(void *cmgr, void *prph);
void BTCentralManagerDidFailToConnectPeripheral(void *cmgr, void *prph, struct bt_error *err);
void BTCentralManagerDidDisconnectPeripheral(void *cmgr, void *prph, struct bt_error *err);
void BTCentralManagerConnectionEventDidOccur(void *cmgr, int event, void *prph);
void BTCentralManagerDidDiscoverPeripheral(void *cmgr, void *prph, struct adv_fields *advData, int rssi);
void BTCentralManagerDidUpdateState(void *cmgr);
void BTCentralManagerWillRestoreState(void *cmgr, struct cmgr_restore_opts *opts);
void BTPeripheralDidDiscoverServices(void *prph, struct bt_error *err);
void BTPeripheralDidDiscoverIncludedServices(void *prph, void *svc, struct bt_error *err);
void BTPeripheralDidDiscoverCharacteristics(void *prph, void *svc, struct bt_error *err);
void BTPeripheralDidDiscoverDescriptors(void *prph, void *chr, struct bt_error *err);
void BTPeripheralDidUpdateValueForCharacteristic(void *prph, void *chr, struct bt_error *err);
void BTPeripheralDidUpdateValueForDescriptor(void *prph, void *dsc, struct bt_error *err);
void BTPeripheralDidWriteValueForCharacteristic(void *prph, void *chr, struct bt_error *err);
void BTPeripheralDidWriteValueForDescriptor(void *prph, void *dsc, struct bt_error *err);
void BTPeripheralIsReadyToSendWriteWithoutResponse(void *prph);
void BTPeripheralDidUpdateNotificationState(void *prph, void *chr, struct bt_error *err);
void BTPeripheralDidReadRSSI(void *prph, int rssi, struct bt_error *err);
void BTPeripheralDidUpdateName(void *prph);
void BTPeripheralDidModifyServices(void *prph, struct obj_arr *inv_svcs);

void BTPeripheralManagerDidUpdateState(void *pmgr);
void BTPeripheralManagerWillRestoreState(void *pmgr, struct pmgr_restore_opts *opts);
void BTPeripheralManagerDidAddService(void *pmgr, void *svc, struct bt_error *err);
void BTPeripheralManagerDidStartAdvertising(void *pmgr, struct bt_error *err);
void BTPeripheralManagerCentralDidSubscribe(void *pmgr, void *cent, void *chr);
void BTPeripheralManagerCentralDidUnsubscribe(void *pmgr, void *cent, void *chr);
void BTPeripheralManagerIsReadyToUpdateSubscribers(void *pmgr);
void BTPeripheralManagerDidReceiveReadRequest(void *pmgr, void *req);
void BTPeripheralManagerDidReceiveWriteRequests(void *pmgr, struct obj_arr *oa);

extern dispatch_queue_t bt_queue;
extern BTDlg *bt_dlg;

#endif
