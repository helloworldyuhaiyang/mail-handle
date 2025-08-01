package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func String[T any](m *T) string {
	b, err := json.Marshal(*m)
	if err != nil {
		return fmt.Sprintf("%+v", *m)
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		return fmt.Sprintf("%+v", *m)
	}
	return out.String()
}

func StringSlice[T any](m []*T) string {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%+v", m)
	}
	var out bytes.Buffer
	err = json.Indent(&out, b, "", "    ")
	if err != nil {
		return fmt.Sprintf("%+v", m)
	}
	return out.String()
}

func StringMap[K comparable, V any](m map[K]V) string {
	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Sprintf("%+v", m)
	}
	return string(b)
}

func StringSliceContains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}
