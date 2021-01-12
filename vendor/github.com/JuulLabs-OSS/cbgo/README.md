# cbgo

cbgo implements Go bindings for [CoreBluetooth](https://developer.apple.com/documentation/corebluetooth?language=objc).

## Documentation

For documentation, see the [CoreBluetooth docs](https://developer.apple.com/documentation/corebluetooth?language=objc).

Examples are in the `examples` directory.

## Scope

cbgo aims to implement all functionality that is supported in macOS 10.13.

## Naming

Function and type names in cbgo are intended to match the corresponding CoreBluetooth functionality as closely as possible.  There are a few (consistent) deviations:

* All cbgo identifiers start with a capital letter to make them public.
* Named arguments in CoreBluetooth functions are eliminated.
* Properties are implemented as a pair of functions (`PropertyName` and `SetPropertyName`).

## Issues

There are definitely memory leaks.  ARC is not compatible with cgo, so objective C memory has to be managed manually.  I didn't see a set of consistent guidelines for object ownership in the CoreBluetooth documentation, so cbgo errs on the side of leaking.  Hopefully this is only an issue for very long running processes!  Any fixes here are much appreciated.
