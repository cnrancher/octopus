package collection

func StringSliceContain(slice []string, target string) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

func StringSliceRemove(slice []string, target string) (result []string) {
	for _, item := range slice {
		if item == target {
			continue
		}
		result = append(result, item)
	}
	return
}
