package notion

import (
	"context"
	"fmt"
	"kanbanchan/internal/aws"
	"kanbanchan/internal/steam"
	"strings"
	"time"

	"github.com/jomei/notionapi"
)

type NotionClient struct {
	ctx       context.Context
	client    *notionapi.Client
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

type GameProperties struct {
	Name              *notionapi.TitleProperty       `json:"name,omitempty"`
	Status            *notionapi.StatusProperty      `json:"status,omitempty"`
	Tags              *notionapi.MultiSelectProperty `json:"tags,omitempty"`
	OfficialStorePage *notionapi.URLProperty         `json:"officialStorePage,omitempty"`
	CompletedDate     *notionapi.DateProperty        `json:"completedDate,omitempty"`
	CoverArt          *notionapi.FilesProperty       `json:"coverArt,omitempty"`
	Platform          *notionapi.MultiSelectProperty `json:"platform,omitempty"`
	ReleaseDate       *notionapi.DateProperty        `json:"releaseDate,omitempty"`
	Rating            *notionapi.RichTextProperty    `json:"rating,omitempty"`
	Notes             *notionapi.RichTextProperty    `json:"notes,omitempty"`
}

func NewClient(ctx context.Context) (*NotionClient, error) {
	var client NotionClient
	var secrets, err = aws.GetSecrets()
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		client.ctx = context.Background()
	} else {
		client.ctx = ctx
	}

	client.client = notionapi.NewClient(notionapi.Token(secrets.Notion.AuthToken))
	client.workspace = secrets.Notion.Workspace
	client.dbIDs.gameDB = secrets.Notion.GameDB
	client.dbIDs.animeDB = secrets.Notion.AnimeDB
	client.dbIDs.movieDB = secrets.Notion.MovieDB
	client.dbIDs.tvDB = secrets.Notion.TVDB
	client.dbIDs.testGame = secrets.Notion.TestGame
	client.dbIDs.testFilm = secrets.Notion.TestFilm

	return &client, nil
}

func (nc *NotionClient) GetDatabase(databaseID string) error {
	db, err := nc.client.Database.Get(nc.ctx, notionapi.DatabaseID(databaseID))
	if err != nil {
		return err
	}

	fmt.Println(db.Properties)
	for key, pc := range db.Properties {
		fmt.Printf("Key: %s, PropertyConfig: %v, PropertyConfig Type: %T\n", key, pc, pc)
	}

	return nil
}

func (nc *NotionClient) GetDatabasePages(databaseID string) error {
	var pages []notionapi.Page

	opts := &notionapi.DatabaseQueryRequest{
		PageSize: 100,
		Sorts:    []notionapi.SortObject{{Property: "Name", Direction: "ascending"}},
	}
	res, err := nc.client.Database.Query(nc.ctx, notionapi.DatabaseID(databaseID), opts)
	if err != nil {
		return err
	}
	pages = append(pages, res.Results...)
	for res.HasMore {
		opts.StartCursor = res.NextCursor
		res, err = nc.client.Database.Query(nc.ctx, notionapi.DatabaseID(databaseID), opts)
		if err != nil {
			return err
		}
		pages = append(pages, res.Results...)
	}

	fmt.Println("Total pages found:", len(pages))
	for _, page := range pages {
		gp := GameProperties{
			Name:              page.Properties["Name"].(*notionapi.TitleProperty),
			Status:            page.Properties["Status"].(*notionapi.StatusProperty),
			Tags:              page.Properties["Tags"].(*notionapi.MultiSelectProperty),
			OfficialStorePage: page.Properties["Official Store Page"].(*notionapi.URLProperty),
			CompletedDate:     page.Properties["Completed Date"].(*notionapi.DateProperty),
			CoverArt:          page.Properties["Cover Art"].(*notionapi.FilesProperty),
			Platform:          page.Properties["Platform"].(*notionapi.MultiSelectProperty),
			ReleaseDate:       page.Properties["Release Date"].(*notionapi.DateProperty),
			Rating:            page.Properties["Rating"].(*notionapi.RichTextProperty),
			Notes:             page.Properties["Notes"].(*notionapi.RichTextProperty),
		}
		printGameProperties(gp)
		fmt.Println()
	}
	return nil
}

func (nc *NotionClient) AddGame(databaseID string, game steam.SteamApp) error {
	// Name (game name)
	// Status (Unreleased, Unowned, Backlog, Playing, Completed)
	// Rating (text "Score: x.y")
	// Platform (tags like PC, Steam, Nintendo, PlayStation, Epic Games, Xbox, GoG, Ubisoft, EA, Mobile)
	// Tags (tags for genre)
	// Official Store Page (link to steam store page)
	// Cover art (URL to header image)
	// Release Date (game release date)
	// Completed Date
	properties := notionapi.Properties{}
	properties["Name"] = &notionapi.TitleProperty{
		Title: []notionapi.RichText{{PlainText: game.Data.Name}},
	}
	releaseDate, err := steam.ParseSteamDate(game.Data.ReleaseDate.Date)
	if err != nil {
		return err
	}
	notionReleaseDate := notionapi.Date(releaseDate)
	properties["Release Date"] = &notionapi.DateProperty{
		Date: &notionapi.DateObject{
			Start: &notionReleaseDate,
		},
	}
	status := "Unreleased"
	if releaseDate.Before(time.Now()) {
		status = "Unowned"
	}
	properties["Status"] = &notionapi.StatusProperty{
		Status: notionapi.Option{
			Name: status,
		},
	}
	properties["Platform"] = &notionapi.MultiSelectProperty{
		MultiSelect: []notionapi.Option{{Name: "Steam"}},
	}
	var genres []notionapi.Option
	for _, genre := range game.Data.Genres {
		genres = append(genres, notionapi.Option{Name: genre.Description})
	}
	properties["Tags"] = &notionapi.MultiSelectProperty{
		MultiSelect: genres,
	}

	page, err := nc.client.Page.Create(nc.ctx, &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseID),
		},
		Properties: properties,
	})
	if err != nil {
		return err
	}

	gp := GameProperties{
		Name:              page.Properties["Name"].(*notionapi.TitleProperty),
		Status:            page.Properties["Status"].(*notionapi.StatusProperty),
		Tags:              page.Properties["Tags"].(*notionapi.MultiSelectProperty),
		OfficialStorePage: page.Properties["Official Store Page"].(*notionapi.URLProperty),
		CompletedDate:     page.Properties["Completed Date"].(*notionapi.DateProperty),
		CoverArt:          page.Properties["Cover Art"].(*notionapi.FilesProperty),
		Platform:          page.Properties["Platform"].(*notionapi.MultiSelectProperty),
		ReleaseDate:       page.Properties["Release Date"].(*notionapi.DateProperty),
		Rating:            page.Properties["Rating"].(*notionapi.RichTextProperty),
		Notes:             page.Properties["Notes"].(*notionapi.RichTextProperty),
	}
	printGameProperties(gp)

	return nil
}

func printGameProperties(gp GameProperties) {
	builder := strings.Builder{}

	builder.WriteString(fmt.Sprintf("Name: %s\n", gp.Name.Title[0].PlainText))
	builder.WriteString(fmt.Sprintf("Status: %s\n", gp.Status.Status.Name))
	if len(gp.Rating.RichText) != 0 {
		builder.WriteString(fmt.Sprintf("Rating: %s\n", gp.Rating.RichText[0].PlainText))
	} else {
		builder.WriteString("Rating: <empty>\n")
	}
	var platforms []string
	for _, platform := range gp.Platform.MultiSelect {
		platforms = append(platforms, platform.Name)
	}
	builder.WriteString(fmt.Sprintf("Platforms: %s\n", strings.Join(platforms[:], ", ")))
	var tags []string
	for _, tag := range gp.Tags.MultiSelect {
		tags = append(tags, tag.Name)
	}
	builder.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(tags[:], ", ")))
	if gp.OfficialStorePage != nil {
		builder.WriteString(fmt.Sprintf("Official Store Page: %s\n", gp.OfficialStorePage.URL))
	} else {
		builder.WriteString("Official Store Page: <empty>\n")
	}
	if len(gp.CoverArt.Files) != 0 {
		builder.WriteString(fmt.Sprintf("Cover Art: %s\n", gp.CoverArt.Files[0].Name))
	} else {
		builder.WriteString("Cover Art: <empty>\n")
	}
	if gp.ReleaseDate.Date != nil {
		builder.WriteString(fmt.Sprintf("Release Date: %s\n", gp.ReleaseDate.Date.Start.String()))
	} else {
		builder.WriteString("Release Date: <empty>\n")
	}
	if gp.CompletedDate.Date != nil {

		builder.WriteString(fmt.Sprintf("Completed Date: %s\n", gp.CompletedDate.Date.Start.String()))
	} else {
		builder.WriteString("Completed Date: <empty>\n")
	}
	if len(gp.Notes.RichText) != 0 {
		builder.WriteString(fmt.Sprintf("Notes: %s\n", gp.Notes.RichText[0].PlainText))
	} else {
		builder.WriteString("Notes: <empty>\n")
	}
	fmt.Print(builder.String())
}
