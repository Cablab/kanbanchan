package notion

import (
	"context"
	"fmt"
	"kanbanchan/internal/aws"
	"kanbanchan/pkg/notion"

	"github.com/jomei/notionapi"
)

const (
	statusUpNext   = "Up Next"
	statusFinished = "Finished"
)

// NotionClient contains a usable Notion client and information about
// databases in the workspace
type NotionClient struct {
	client    notion.NotionClient
	workspace string
	dbIDs     struct {
		gameDB    string
		animeDB   string
		movieDB   string
		tvDB      string
		testGame  string
		testAnime string
		testMovie string
		testTV    string
	}
}

// NewClient sets up an authenticated Notion client and user information about
// databases in the workspace
func NewClient(ctx context.Context) (*NotionClient, error) {
	var client NotionClient
	var secrets, err = aws.GetSecrets()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve secrets: %s", err.Error())
	}

	notionClient, err := notion.NewClient(ctx, secrets.Notion.AuthToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create notion client: %s", err.Error())
	}

	client.client = *notionClient
	client.workspace = secrets.Notion.Workspace
	client.dbIDs.gameDB = secrets.Notion.GameDB
	client.dbIDs.animeDB = secrets.Notion.AnimeDB
	client.dbIDs.movieDB = secrets.Notion.MovieDB
	client.dbIDs.tvDB = secrets.Notion.TVDB
	client.dbIDs.testGame = secrets.Notion.TestGame
	client.dbIDs.testAnime = secrets.Notion.TestAnime
	client.dbIDs.testMovie = secrets.Notion.TestMovie
	client.dbIDs.testTV = secrets.Notion.TestTV

	return &client, nil
}

// GetDatabase retrieves the specified database
func (nc *NotionClient) GetDatabase(databaseID string) (*notionapi.Database, error) {
	db, err := nc.client.GetDatabase(databaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database id %s: %s", databaseID, err.Error())
	}

	return db, nil
}
