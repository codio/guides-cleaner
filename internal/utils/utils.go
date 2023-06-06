package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

func ContainedInArray[T comparable](array []T, value T) bool {
	for _, n := range array {
		if value == n {
			return true
		}
	}
	return false
}

func GetArrayFromJson[T interface{}](pathToFile, key string) ([]T, error) {
	var root interface{}
	jsonFile, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	bytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(bytes, &root); err != nil {
		return nil, err
	}

	records, ok := root.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("error processing file: %s", pathToFile)
	}
	array, ok := records[key].([]interface{})
	if ok {
		arrayT := make([]T, len(array))
		for i, v := range array {
			arrayT[i], ok = v.(T)
			if !ok {
				return nil, fmt.Errorf("typecast error")
			}
		}
		return arrayT, nil
	}
	return []T{}, nil
}
