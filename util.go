package main

import "fmt"

func GetMapStrKeys(m interface{}) ([]string, error) {
	var keys []string
	switch m.(type) {
	case map[string]struct{}:
		mm := m.(map[string]struct{})
		for key := range mm {
			keys = append(keys, key)
		}
	default:
		return nil, fmt.Errorf("not map[string] type")
	}
	return keys, nil
}
