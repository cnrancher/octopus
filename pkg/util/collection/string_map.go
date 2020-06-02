package collection

func StringMapCopy(source map[string]string) map[string]string {
	return StringMapCopyInto(source, make(map[string]string, len(source)))
}

func StringMapCopyInto(source, destination map[string]string) map[string]string {
	if destination == nil {
		return nil
	}
	if len(source) == 0 {
		return destination
	}

	for k, v := range source {
		destination[k] = v
	}
	return destination
}

func DiffStringMap(left, right map[string]string) bool {
	for lk, lv := range left {
		if rv, exist := right[lk]; !exist {
			return true
		} else if lv != rv {
			return true
		}
	}
	return false
}
