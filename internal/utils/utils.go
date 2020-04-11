package utils

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// Complexity max == O(2N)
func CompareUnorderedArray(d1, d2 []string) bool {
	if d1 == nil || d2 == nil {
		return false
	}

	if len(d1) != len(d2) {
		return false
	}

	m := make(map[string]bool, len(d1))
	for _, d := range d1 {
		m[d] = true
	}

	for _, d := range d2 {
		if _, ok := m[d]; !ok {
			return false
		}
	}

	return true
}
