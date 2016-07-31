package utils

// StringInSlice returns true when string passed in first argument is present in slice passed as second argument
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
