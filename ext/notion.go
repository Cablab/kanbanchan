package ext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kanbanchan/internal"
	"net/http"
)

const (
	notionURL          = "https://api.notion.com/v1"
	headerAuthKey      = "Authorization"
	headerVersionKey   = "Notion-Version"
	headerVersionValue = "2022-06-28"
	contentTypeKey     = "Content-Type"
	contentTypeJSON    = "application/json"
)

type NotionInterface interface {
	GetDatabaseProperties(databaseID string) error
	GetDatabasePages(databaseID string) error
}

type NotionClient struct {
	ctx       context.Context
	apiToken  string
	workspace string
	dbIDs     struct {
		gameDB   string
		animeDB  string
		movieDB  string
		tvDB     string
		testGame string
		testFilm string
	}
}

type DatabaseQueryRequest struct {
	Filter      interface{}    `json:"filter,omitempty"`
	Sorts       []DatabaseSort `json:"sorts,omitempty"`
	StartCursor string         `json:"start_cursor,omitempty"`
	PageSize    int32          `json:"page_size,omitempty"`
}

type DatabaseSort struct {
	Property  string `json:"property"`
	Direction string `json:"direction"`
}

type DatabaseQueryResponse struct {
	Object     string        `json:"object,omitempty"`
	Results    []interface{} `json:"results,omitempty"`
	NextCursor string        `json:"next_cursor,omitempty"`
	HasMore    bool          `json:"has_more,omitempty"`
	Type       string        `json:"type,omitempty"`
	Page       interface{}   `json:"page,omitempty"`
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
	client.dbIDs.testGame = secrets.Notion.TestGame
	client.dbIDs.testFilm = secrets.Notion.TestFilm

	return &client, nil
}

func (nc *NotionClient) GetDatabaseProperties(databaseID string) error {
	client := &http.Client{}
	endpoint := fmt.Sprintf("%s/databases/%s", notionURL, databaseID)

	req, err := http.NewRequest("GET", endpoint, nil)
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

func (nc *NotionClient) GetDatabasePages(databaseID string) error {
	var pages []interface{}
	reqBodyData := DatabaseQueryRequest{
		PageSize: 100,
		Sorts:    []DatabaseSort{{Property: "Name", Direction: "ascending"}},
	}
	respBody, err := queryDatabase(nc, databaseID, reqBodyData)
	if err != nil {
		return err
	}
	pages = append(pages, respBody.Results...)
	for respBody.HasMore {
		reqBodyData.StartCursor = respBody.NextCursor
		respBody, err = queryDatabase(nc, databaseID, reqBodyData)
		if err != nil {
			return err
		}
		pages = append(pages, respBody.Results...)
	}

	fmt.Println(pages)
	fmt.Println("Total pages found:", len(pages))
	return nil
}

func queryDatabase(nc *NotionClient, databaseID string, options DatabaseQueryRequest) (*DatabaseQueryResponse, error) {
	var respBody DatabaseQueryResponse
	client := &http.Client{}
	endpoint := fmt.Sprintf("%s/databases/%s/query", notionURL, databaseID)

	reqBody, err := json.Marshal(options)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request body: %s", err.Error())
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %s", err.Error())
	}
	req.Header.Set(headerVersionKey, headerVersionValue)
	req.Header.Set(headerAuthKey, nc.apiToken)
	req.Header.Set(contentTypeKey, contentTypeJSON)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending POST request: %s", err.Error())
	}
	defer resp.Body.Close()

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %s", err.Error())
	}

	err = json.Unmarshal(resBody, &respBody)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling response body: %s", err.Error())
	}

	return &respBody, nil
}
