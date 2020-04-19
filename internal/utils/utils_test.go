package utils

import "testing"

func TestContains(t *testing.T) {
	testCases := []struct {
		array []string
		term  string
		want  bool
	}{
		{[]string{"found"}, "found", true},
		{[]string{"found"}, "notfound", false},
	}

	for _, tc := range testCases {
		got := Contains(tc.array, tc.term)
		if got != tc.want {
			t.Errorf("Contains(%s, %s) = %t; want %t", tc.array, tc.term, got, tc.want)
		}
	}
}

func TestCompareUnorderedArray(t *testing.T) {
	arrays := []string{"helloworld.io", "helloworld.com"}

	arraysUnordered := []string{"helloworld.com", "helloworld.io"}

	arraysDiff := []string{"diff.io", "diff.com"}

	arraysMissing := []string{"helloworld.com"}

	testCases := []struct {
		a1, a2 []string
		want   bool
	}{
		{arrays, arrays, true},
		{arrays, arraysUnordered, true},
		{arrays, arraysDiff, false},
		{arrays, arraysMissing, false},
		{arrays, nil, false},
	}

	for _, tc := range testCases {
		got := CompareUnorderedArray(tc.a1, tc.a2)

		if got != tc.want {
			t.Errorf("CompareUnorderedArray(%s, %s) = %t; want %t", tc.a1, tc.a2, got, tc.want)
		}
	}
}

func TestDiffUnorderedArray(t *testing.T) {
	a1 := []string{"helloworld.io", "helloworld.com"}
	a2 := []string{"helloworld.io"}
	a3 := []string{"helloworld.com"}

	testCases := []struct {
		a1, a2, want []string
	}{
		{a1, a1, []string{}},
		{a1, a2, []string{"helloworld.com"}},
		{a1, []string{}, a1},
		{nil, a2, a2},
		{a1, nil, a1},
		{a2, a3, a1},
	}

	for _, tc := range testCases {
		got := DiffUnorderedArray(tc.a1, tc.a2)

		if !CompareUnorderedArray(got, tc.want) {
			t.Errorf("DiffUnorderedArray(%s, %s) = %s; want %s", tc.a1, tc.a2, got, tc.want)
		}
	}
}

func TestIsArrayEmpty(t *testing.T) {
	testCases := []struct {
		a    []string
		want bool
	}{
		{[]string{"example.io"}, false},
		{[]string{}, true},
		{nil, true},
	}

	for _, tc := range testCases {
		got := IsArrayEmpty(tc.a)

		if got != tc.want {
			t.Errorf("isArrayEmpty(%s) = %t, want %t", tc.a, got, tc.want)
		}
	}
}

func TestInt32Ptr(t *testing.T) {
	want := int32(3)
	got := Int32Ptr(want)

	if *got != want {
		t.Errorf("Int32Ptr(%d) = %d, want %d", want, got, want)
	}
}

func TestMergeMap(t *testing.T) {
	testCases := []struct {
		a, b, want map[string]string
	}{
		{map[string]string{"a": "a"}, map[string]string{"b": "b"}, map[string]string{"a": "a", "b": "b"}},
		{map[string]string{"a": "a"}, map[string]string{}, map[string]string{"a": "a"}},
		{map[string]string{}, map[string]string{"b": "b"}, map[string]string{"b": "b"}},
		{map[string]string{"a": "a"}, map[string]string{"a": "b"}, map[string]string{"a": "b"}},
	}

	for _, tc := range testCases {
		got := MergeMap(tc.a, tc.b)

		if !CompareMap(tc.want, got) {
			t.Errorf("MergeMap(%s, %s) = %s, want %s", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCompareMap(t *testing.T) {
	testCases := []struct {
		a, b map[string]string
		want bool
	}{
		{map[string]string{"a": "a"}, map[string]string{"b": "b"}, false},
		{map[string]string{"a": "a"}, nil, false},
		{map[string]string{"a": "a"}, map[string]string{}, false},
		{map[string]string{}, map[string]string{"b": "b"}, false},
		{map[string]string{"a": "a"}, map[string]string{"a": "a", "b": "b"}, false},
		{map[string]string{"a": "a"}, map[string]string{"a": "a"}, true},
	}

	for _, tc := range testCases {
		got := CompareMap(tc.a, tc.b)

		if got != tc.want {
			t.Errorf("CompareMap(%s, %s) = %t, want %t", tc.a, tc.b, got, tc.want)
		}
	}
}
