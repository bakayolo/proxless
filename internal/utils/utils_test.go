package utils

import "testing"

func Test_Contains(t *testing.T) {
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

func Test_CompareUnorderedArrays(t *testing.T) {
	arrays := []string{"helloworld.io", "helloworld.com"}

	arraysUnordered := []string{"helloworld.com", "helloworld.io"}

	arraysDiff := []string{"diff.io", "diff.com"}

	arraysMissing := []string{"helloworld.com"}

	testCases := []struct {
		a1, a2 []string
		want   bool
	}{
		{
			arrays,
			arrays,
			true,
		},
		{
			arrays,
			arraysUnordered,
			true,
		},
		{
			arrays,
			arraysDiff,
			false,
		},
		{
			arrays,
			arraysMissing,
			false,
		},
		{
			arrays,
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		got := CompareUnorderedArray(tc.a1, tc.a2)

		if got != tc.want {
			t.Errorf("CompareUnorderedArray(%s, %s) = %t; want %t", tc.a1, tc.a2, got, tc.want)
		}
	}
}
