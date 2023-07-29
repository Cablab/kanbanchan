package aws

import (
	"context"
	"encoding/json"
	"os"
)

type SecretsClient struct {
}

func NewClient(ctx context.Context) (*SecretsClient, error) {
	return nil, nil
}

// LocalKeys mimics the JSON structure of local key storage
type LocalSecrets struct {
	Discord struct {
		Key string `json:"key"`
	} `json:"discord"`
	Google struct {
		Key string `json:"key"`
	} `json:"google"`
	Notion struct {
		AuthToken string `json:"authToken"`
		Workspace string `json:"workspace"`
		GameDB    string `json:"gameDB"`
		AnimeDB   string `json:"animeDB"`
		MovieDB   string `json:"movieDB"`
		TVDB      string `json:"tvDB"`
		TestGame  string `json:"testGame"`
		TestFilm  string `json:"testFilm"`
	} `json:"notion"`
	Steam struct {
		ID          string `json:"id"`
		Key         string `json:"key"`
		Collections struct {
			Completed []json.Number `json:"completed"`
			Playing   []json.Number `json:"playing"`
			UpNext    []json.Number `json:"upNext"`
		} `json:"collections"`
	} `json:"steam"`
}

// GetSecrets retrieves secrets
func GetSecrets() (*LocalSecrets, error) {
	var keys LocalSecrets
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
