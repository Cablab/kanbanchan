package ext

import (
	"context"
	"fmt"
	"io"
	"kanbanchan/internal"
	"net/http"
)

const (
	notionURL          = "https://api.notion.com/v1/"
	headerAuthKey      = "Authorization"
	headerVersionKey   = "Notion-Version"
	headerVersionValue = "2022-06-28"
)

type NotionInterface interface {
	GetDatabase(databaseID string) error
}

type NotionClient struct {
	ctx       context.Context
	apiToken  string
	workspace string
	dbIDs     struct {
		gameDB  string
		animeDB string
		movieDB string
		tvDB    string
	}
}

func NewNotionClient(ctx context.Context) (*NotionClient, error) {
	var client NotionClient
	var secrets, err = internal.GetSecrets()
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		client.ctx = context.Background()
	} else {
		client.ctx = ctx
	}

	client.apiToken = secrets.Notion.AuthToken
	client.workspace = secrets.Notion.Workspace
	client.dbIDs.gameDB = secrets.Notion.GameDB
	client.dbIDs.animeDB = secrets.Notion.AnimeDB
	client.dbIDs.movieDB = secrets.Notion.MovieDB
	client.dbIDs.tvDB = secrets.Notion.TVDB

	return &client, nil
}

func (nc *NotionClient) GetDatabase(databaseID string) error {
	client := &http.Client{}
	endpoint := fmt.Sprintf("databases/%s", databaseID)

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", notionURL, endpoint), nil)
	if err != nil {
		return err
	}
	req.Header.Set(headerVersionKey, headerVersionValue)
	req.Header.Set(headerAuthKey, nc.apiToken)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))
	return nil
}
