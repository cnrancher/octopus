package cbgo

type AdvServiceData struct {
	UUID UUID
	Data []byte
}

// AdvFields represents the contents of an advertisement received during
// scanning.
type AdvFields struct {
	LocalName        string
	ManufacturerData []byte
	TxPowerLevel     *int
	Connectable      *bool
	ServiceUUIDs     []UUID
	ServiceData      []AdvServiceData
}

// AdvData represents the configurable portion of outgoing advertisements.
type AdvData struct {
	LocalName    string
	ServiceUUIDs []UUID

	// If len>0, this overrides the other fields.
	IBeaconData []byte
}
