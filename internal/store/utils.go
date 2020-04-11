package store

// cannot believe go does not have a built-in function for that
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}
