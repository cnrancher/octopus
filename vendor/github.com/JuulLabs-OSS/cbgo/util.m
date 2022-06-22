#import "bt.h"

struct byte_arr
nsdata_to_byte_arr(const NSData *nsdata)
{
    return (struct byte_arr) {
        .data = [nsdata bytes],
        .length = [nsdata length],
    };
}

NSData *
byte_arr_to_nsdata(const struct byte_arr *ba)
{
    if (ba->length == 0) {
        return nil;
    } else {
        return [NSData dataWithBytes: ba->data length: ba->length];
    }
}

struct obj_arr
nsarray_to_obj_arr(const NSArray *arr)
{
    struct obj_arr oa = {0};

    if (arr != nil && [arr count] > 0) {
        oa.count = [arr count],
        oa.objs = malloc(oa.count * sizeof (void *)),
        assert(oa.objs != NULL);

        for (int i = 0; i < oa.count; i++) {
            oa.objs[i] = [arr objectAtIndex:i];
        }
    }

    return oa;
}

NSArray *
obj_arr_to_nsarray(const struct obj_arr *oa)
{
    NSMutableArray *nsa = [[NSMutableArray alloc] init];
    [nsa autorelease];

    for (int i = 0; i < oa->count; i++) {
        [nsa addObject:oa->objs[i]];
    }

    return nsa;
}

NSString *
str_to_nsstring(const char *s)
{
    return [[NSString alloc] initWithCString:s encoding:NSUTF8StringEncoding];
}

struct bt_error
nserror_to_bt_error(const NSError *err)
{
    if (err == NULL) {
        return (struct bt_error) {0};
    } else {
        return (struct bt_error) {
            .msg = [err.localizedDescription UTF8String],
            .code = err.code,
        };
    }
}

NSUUID *
str_to_nsuuid(const char *s)
{
    NSString *nss = str_to_nsstring(s);
    return [[NSUUID alloc] initWithUUIDString:nss];
}

CBUUID *
str_to_cbuuid(const char *s)
{
    NSString *nss = str_to_nsstring(s);
    return [CBUUID UUIDWithString:nss];
}

NSArray *
strs_to_nsuuids(const struct string_arr *sa)
{
    if (sa == nil || sa->count == 0) {
        return nil;
    };

    NSMutableArray *arr = [[NSMutableArray alloc] init]; 
    [arr autorelease];

    for (int i = 0; i < sa->count; i++) {
        [arr addObject:str_to_nsuuid(sa->strings[i])];
    }

    return arr;
}

NSArray *
strs_to_cbuuids(const struct string_arr *sa)
{
    if (sa == nil || sa->count == 0) {
        return nil;
    };

    NSMutableArray *arr = [[NSMutableArray alloc] init]; 
    [arr autorelease];

    for (int i = 0; i < sa->count; i++) {
        [arr addObject:str_to_cbuuid(sa->strings[i])];
    }

    return arr;
}

NSArray *
strs_to_nsstrings(const struct string_arr *sa)
{
    if (sa == nil || sa->count == 0) {
        return nil;
    };

    NSMutableArray *arr = [[NSMutableArray alloc] init]; 
    [arr autorelease];

    for (int i = 0; i < sa->count; i++) {
        [arr addObject:str_to_nsstring(sa->strings[i])];
    }

    return arr;
}

struct string_arr
cbuuids_to_strs(const NSArray *cbuuids)
{
    struct string_arr sa = {0};

    if (cbuuids != nil && [cbuuids count] > 0) {
        sa.count = [cbuuids count];
        sa.strings = malloc(sa.count * sizeof (char *));
        assert(sa.strings != NULL);

        for (int i = 0; i < sa.count; i++) {
            sa.strings[i] = [[[cbuuids objectAtIndex:i] UUIDString] UTF8String];
        }
    }

    return sa;
}

int 
dict_int(NSDictionary *dict, NSString *key, int dflt)
{
    NSNumber *nsn = [dict objectForKey:key];
    if (nsn == nil) {
        return dflt;
    } else {
        return [nsn intValue];
    }
}

const char *
dict_string(NSDictionary *dict, NSString *key)
{
    return [[dict objectForKey:key] UTF8String];
}

const void *
dict_data(NSDictionary *dict, NSString *key, int *out_len)
{
    NSData *data;

    data = [dict objectForKey:key];

    *out_len = [data length];
    return [data bytes];
}

const struct byte_arr
dict_bytes(NSDictionary *dict, NSString *key)
{
    const NSData *nsdata = [dict objectForKey:key];
    return nsdata_to_byte_arr(nsdata);
}

void
dict_set_bool(NSMutableDictionary *dict, NSString *key, bool val)
{
    [dict setObject:[NSNumber numberWithBool:val] forKey:key];
}

void
dict_set_int(NSMutableDictionary *dict, NSString *key, int val)
{
    [dict setObject:[NSNumber numberWithInt:val] forKey:key];
}
