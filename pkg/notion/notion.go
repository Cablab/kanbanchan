package notion

import (
	"context"
	"fmt"

	"github.com/jomei/notionapi"
)

// NotionClient contains an authenticated Notion client
type NotionClient struct {
	ctx    context.Context
	client *notionapi.Client
}

// NewClient creates an authenticated Notion client
func NewClient(ctx context.Context, authToken string) (*NotionClient, error) {
	var client NotionClient
	if ctx == nil {
		client.ctx = context.Background()
	} else {
		client.ctx = ctx
	}
	client.client = notionapi.NewClient(notionapi.Token(authToken))
	return &client, nil
}

// GetDatabase retrieves the specified database
func (nc *NotionClient) GetDatabase(databaseID string) (*notionapi.Database, error) {
	db, err := nc.client.Database.Get(nc.ctx, notionapi.DatabaseID(databaseID))
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %s", err.Error())
	}

	return db, nil
}

// GetDatabasePages retrieves all pages from the specified database
func (nc *NotionClient) GetDatabasePages(databaseID string, opts *notionapi.DatabaseQueryRequest) ([]notionapi.Page, error) {
	var pages []notionapi.Page

	res, err := nc.client.Database.Query(nc.ctx, notionapi.DatabaseID(databaseID), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve database pages: %s", err.Error())
	}
	pages = append(pages, res.Results...)
	for res.HasMore {
		opts.StartCursor = res.NextCursor
		res, err = nc.client.Database.Query(nc.ctx, notionapi.DatabaseID(databaseID), opts)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve database pages: %s", err.Error())
		}
		pages = append(pages, res.Results...)
	}

	return pages, nil
}

// CreatePage creates the specified page
func (nc *NotionClient) CreatePage(opts *notionapi.PageCreateRequest) (*notionapi.Page, error) {
	page, err := nc.client.Page.Create(nc.ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %s", err.Error())
	}

	return page, nil
}
