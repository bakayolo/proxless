package utils

func Contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

// Return true if same elements in a1 and a2
func CompareUnorderedArray(a1, a2 []string) bool {
	if a1 == nil && a2 != nil || a1 != nil && a2 == nil {
		return false
	}

	if len(a1) != len(a2) {
		return false
	}

	m := make(map[string]bool, len(a1))
	for _, el := range a1 {
		m[el] = true
	}

	for _, el := range a2 {
		if _, ok := m[el]; !ok {
			return false
		}
	}

	return true
}

// TODO review complexity
// Return an array with the difference between a1 and a2
func DiffUnorderedArray(a1, a2 []string) []string {
	if a1 == nil {
		return a2
	}

	if a2 == nil {
		return a1
	}

	m1 := arrayToMap(a1)
	m2 := arrayToMap(a2)

	diff1 := keysNotInMap(a1, m2) // values of a1 that are not in m2
	diff2 := keysNotInMap(a2, m1) // values of a2 that are not in m1

	diff := append(diff1, diff2...)

	if diff == nil {
		return []string{}
	}

	return diff
}

func arrayToMap(array []string) map[string]bool {
	m := make(map[string]bool, len(array))
	for _, el := range array {
		m[el] = true
	}
	return m
}

// return list of keys that are not in the map
func keysNotInMap(keys []string, m map[string]bool) []string {
	var diff []string
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			diff = append(diff, k)
		}
	}
	return diff
}

func IsArrayEmpty(array []string) bool {
	return array == nil || len(array) == 0
}

func Int32Ptr(i int32) *int32 {
	return &i
}

// return the union of map A and B
// /!\ if B has A values, the output will have B values
func MergeMap(mapA, mapB map[string]string) map[string]string {
	output := make(map[string]string)

	for k, v := range mapA {
		output[k] = v
	}

	for k, v := range mapB {
		output[k] = v
	}

	return output
}

func CompareMap(m1, m2 map[string]string) bool {
	if m1 == nil && m2 != nil || m1 != nil && m2 == nil {
		return false
	}

	if len(m1) != len(m2) {
		return false
	}

	for k, v := range m1 {
		if m2[k] != v {
			return false
		}
	}

	return true
}
