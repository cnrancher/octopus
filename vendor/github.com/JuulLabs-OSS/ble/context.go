package ble

// ContextKey is a type used for keys of a context
type ContextKey string

var (
	// ContextKeySig for SigHandler context
	ContextKeySig = ContextKey("sig")
	// ContextKeyCCC for per connection contexts
	ContextKeyCCC = ContextKey("ccc")
)
