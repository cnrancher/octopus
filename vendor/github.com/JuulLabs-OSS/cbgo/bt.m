#import <Foundation/Foundation.h>
#import <CoreBluetooth/CoreBluetooth.h>
#import "bt.h"

dispatch_queue_t bt_queue;
static bool bt_loop_active;

/**
 * Universal delegate.  All callbacks get funnelled through this object before
 * being translated to the appropriate Go calls.
 */
BTDlg *bt_dlg;

/**
 * Starts a thread that processes CoreBluetooth events from the BT queue.
 *
 * @return                      false if the thread was already started;
 *                              true if this call started a new thread.
 */
bool
bt_start()
{
    if (bt_loop_active) {
        return false;
    }
    bt_loop_active = true;

    dispatch_async(dispatch_get_main_queue(), ^{
        NSRunLoop *rl;
        rl = [NSRunLoop currentRunLoop];

        bool done;
        do {
            done = [rl runMode:NSDefaultRunLoopMode
                    beforeDate:[NSDate distantFuture]];
        } while (bt_loop_active && !done);
    });

    return true;
}

void
bt_stop()
{
    bt_loop_active = false;
}

void
bt_init()
{
    // XXX: I have no idea why a separate queue is required here.  When I
    // attempt to use the default queue, the run loop does not receive any
    // events.
    if (bt_queue == NULL) {
        bt_queue = dispatch_queue_create("bt_queue", NULL);
    }

    bt_dlg = [[BTDlg alloc] init];
    [bt_dlg retain];
}
