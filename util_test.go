package main

import (
	"reflect"
	"testing"
)

func TestGetMapStrKeys(t *testing.T) {
	var testCases = []struct {
		in       interface{}
		expected []string
	}{
		{
			in: map[string]struct{}{
				"a": struct{}{},
				"b": struct{}{},
			},
			expected: []string{"a", "b"},
		},
	}
	for _, testCase := range testCases {
		out, err := GetMapStrKeys(testCase.in)
		if err != nil {
			t.Errorf("not nil")
		}
		if !reflect.DeepEqual(testCase.expected, out) {
			t.Errorf("not equal")
		}
	}
}
