package v1alpha1

// ReferencesHandler is a convenient `map[string]*ConnectRequestReferenceEntry` handler for obtaining data while avoiding the nil pointer error.
type ReferencesHandler map[string]*ConnectRequestReferenceEntry

// GetData returns the data of specified name and itemName,
// it's always return nil if the data bytes is not existed or empty.
func (h ReferencesHandler) GetData(name, itemName string) []byte {
	if len(h) == 0 {
		return nil
	}

	var refItems, refExist = h[name]
	if !refExist || len(refItems.Items) == 0 {
		return nil
	}

	var refItem, refItemExist = refItems.Items[itemName]
	if !refItemExist {
		return nil
	}

	if len(refItem) == 0 {
		return nil
	}
	return refItem
}

// ToDataMap returns a `map[string]map[string][]byte` that constructed the references data,
// it's always return nil if the references is nil or empty.
func (h ReferencesHandler) ToDataMap() map[string]map[string][]byte {
	if len(h) == 0 {
		return nil
	}

	var refMap = make(map[string]map[string][]byte, len(h))
	for refKey, refValue := range h {
		if refValue == nil || len(refValue.Items) == 0 {
			continue
		}
		var refItems = make(map[string][]byte, len(refValue.Items))
		for refItemKey, refItemValue := range refValue.Items {
			refItems[refItemKey] = refItemValue
		}
		refMap[refKey] = refItems
	}
	return refMap
}

// GetReferencesHandler returns a ReferencesHandler for obtaining data.
func (m *ConnectRequest) GetReferencesHandler() ReferencesHandler {
	if m != nil {
		return m.References
	}
	return nil
}
