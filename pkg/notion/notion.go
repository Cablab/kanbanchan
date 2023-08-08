package notion

import (
	"context"
	"fmt"
	"strings"
	"time"

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
		return nil, fmt.Errorf("failed to get database id %s: %s", databaseID, err.Error())
	}

	return db, nil
}

// GetPageByID fetches a single page by ID
func (nc *NotionClient) GetPageByID(ctx context.Context, pageID string) (*notionapi.Page, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(pageID) == "" {
		return nil, fmt.Errorf("request made with empty page id")
	}

	page, err := nc.client.Page.Get(ctx, notionapi.PageID(pageID))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve page id %s: %s", pageID, err.Error())
	}

	return page, nil
}

// GetDatabasePages retrieves all pages from the specified database
func (nc *NotionClient) GetDatabasePages(databaseID string, opts *notionapi.DatabaseQueryRequest) ([]notionapi.Page, error) {
	var pages []notionapi.Page

	res, err := nc.client.Database.Query(nc.ctx, notionapi.DatabaseID(databaseID), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve pages from database id %s: %s", databaseID, err.Error())
	}
	pages = append(pages, res.Results...)
	for res.HasMore {
		opts.StartCursor = res.NextCursor
		res, err = nc.client.Database.Query(nc.ctx, notionapi.DatabaseID(databaseID), opts)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve pages from database id %s: %s", databaseID, err.Error())
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

// UpdatePage updates the specified page
func (nc *NotionClient) UpdatePage(ctx context.Context, pageID string, opts *notionapi.PageUpdateRequest) (*notionapi.Page, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if strings.TrimSpace(pageID) == "" {
		return nil, fmt.Errorf("request made with empty page id")
	}

	page, err := nc.client.Page.Update(ctx, notionapi.PageID(pageID), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to update page id %s: %s", pageID, err.Error())
	}

	return page, nil
}

// ParseNotionDate takes a Notion date and parses it to Golang time
func (nc *NotionClient) ParseNotionDate(date notionapi.Date) (time.Time, error) {
	timestamp, err := time.Parse(time.RFC3339, date.String())
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse notion date timestamp: %s", err.Error())
	}
	return timestamp, nil
}
