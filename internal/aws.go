package internal

import (
	"encoding/json"
	"os"
)

// LocalKeys mimics the JSON structure of local key storage
type LocalKeys struct {
	Discord string `json:"discord"`
	Google  string `json:"google"`
	Notion  string `json:"notion"`
	Steam   string `json:"steam"`
}

// GetSecrets retrieves secrets
func GetSecrets() (*LocalKeys, error) {
	var keys LocalKeys
	fileContent, err := os.ReadFile("../../local/secrets.json")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(fileContent, &keys)
	if err != nil {
		return nil, err
	}
	return &keys, nil
}
