package notion

import (
	"context"
	"fmt"
	"kanbanchan/internal/aws"
	"kanbanchan/internal/steam"
	"kanbanchan/pkg/notion"
	"strings"
	"time"

	"github.com/jomei/notionapi"
)

// NotionClient contains a usable Notion client and information about
// databases in the workspace
type NotionClient struct {
	client    notion.NotionClient
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

// GameProperties contains info about pages in the Games database
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
	client.dbIDs.testFilm = secrets.Notion.TestFilm

	return &client, nil
}

// GetDatabase retrieves the specified database
func (nc *NotionClient) GetDatabase(databaseID string) (*notionapi.Database, error) {
	db, err := nc.client.GetDatabase(databaseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %s", err.Error())
	}

	return db, nil
}

// GetGamePages retrieves all pages in the Games DB
func (nc *NotionClient) GetGamePages() (*map[string]GameProperties, error) {
	opts := &notionapi.DatabaseQueryRequest{
		PageSize: 100,
		Sorts:    []notionapi.SortObject{{Property: "Name", Direction: "ascending"}},
	}
	pages, err := nc.client.GetDatabasePages(nc.dbIDs.gameDB, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get game pages: %s", err.Error())
	}

	games := make(map[string]GameProperties)
	for _, page := range pages {
		game := GameProperties{
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
		_, ok := games[game.Name.Title[0].PlainText]
		if !ok {
			games[game.Name.Title[0].PlainText] = game
		}
	}
	return &games, nil
}

// AddGame adds a game to the Games DB
func (nc *NotionClient) AddGame(databaseID string, game steam.SteamGame) error {
	properties := notionapi.Properties{}
	status := "Unreleased"
	upNextVal, upNextOk := game.Collections["UpNext"]
	playingVal, playingOk := game.Collections["Playing"]
	completedVal, completedOk := game.Collections["Completed"]
	if game.ReleaseDate.Before(time.Now()) {
		status = "Unowned"
	} else if upNextOk && upNextVal {
		status = "Up Next"
	} else if playingOk && playingVal {
		status = "Playing"
	} else if completedOk && completedVal {
		status = "Completed"
	}
	var genres []notionapi.Option
	for _, genre := range game.Genres {
		genres = append(genres, notionapi.Option{Name: genre})
	}
	notionReleaseDate := notionapi.Date(game.ReleaseDate)

	properties["Name"] = &notionapi.TitleProperty{
		Title: []notionapi.RichText{{PlainText: game.Name}},
	}
	properties["Status"] = &notionapi.StatusProperty{
		Status: notionapi.Option{
			Name: status,
		},
	}
	properties["Platform"] = &notionapi.MultiSelectProperty{
		MultiSelect: []notionapi.Option{{Name: "Steam"}, {Name: "kanbanchan"}},
	}
	properties["Tags"] = &notionapi.MultiSelectProperty{
		MultiSelect: genres,
	}
	properties["Official Store Page"] = &notionapi.URLProperty{URL: fmt.Sprintf("https://store.steampowered.com/app/%s", game.ID)}
	properties["Cover Art"] = &notionapi.FilesProperty{
		Files: []notionapi.File{{Name: game.HeaderImage}},
	}
	properties["Release Date"] = &notionapi.DateProperty{
		Date: &notionapi.DateObject{
			Start: &notionReleaseDate,
		},
	}

	page, err := nc.client.CreatePage(&notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(databaseID),
		},
		Properties: properties,
	})
	if err != nil {
		return fmt.Errorf("failed to add game: %s", err.Error())
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
	PrintGameProperties(gp)

	return nil
}

// PrintGameProperties prints the columns of a game from the Games DB in readable format
func PrintGameProperties(gp GameProperties) {
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
