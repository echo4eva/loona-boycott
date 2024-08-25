package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadJSONtoMap(filename string) (map[string]string, error) {
	// Read the file
	data, err := os.ReadFile(fmt.Sprintf("data/%s", filename))
	if err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	// Unmarshal JSON data into map
	var results map[string]string
	err = json.Unmarshal(data, &results)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	return results, nil
}
