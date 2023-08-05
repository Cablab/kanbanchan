package notion

import (
	"fmt"
	"kanbanchan/internal/steam"
	"os"
	"strings"
	"time"

	"github.com/jomei/notionapi"
)

const (
	StatusUnreleased = "Unreleased"
	StatusUnowned    = "Unowned"
	StatusPlaying    = "Playing"
)

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

// GetGamePages retrieves all pages in the Games DB
func (nc *NotionClient) GetGamePages(options *notionapi.DatabaseQueryRequest) (*map[string]GameProperties, error) {
	gameDB := nc.dbIDs.gameDB
	env := os.Getenv("ENVIRONMENT")
	if env == "development" || env == "dev" || env == "staging" || env == "local" {
		gameDB = nc.dbIDs.testGame
	}
	options = setQueryOptions(options)
	pages, err := nc.client.GetDatabasePages(gameDB, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get game pages from database id %s: %s", gameDB, err.Error())
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
func (nc *NotionClient) AddGame(game steam.SteamGame) error {
	gameDB := nc.dbIDs.gameDB
	env := os.Getenv("ENVIRONMENT")
	if env == "development" || env == "dev" || env == "staging" || env == "local" {
		gameDB = nc.dbIDs.testGame
	}

	properties := notionapi.Properties{}
	status := determineGameStatus(game)

	var genres []notionapi.Option
	for _, genre := range game.Genres {
		genres = append(genres, notionapi.Option{Name: genre})
	}
	notionReleaseDate := notionapi.Date(game.ReleaseDate)

	properties["Name"] = &notionapi.TitleProperty{
		Title: []notionapi.RichText{{
			Text: &notionapi.Text{
				Content: game.Name,
			},
			PlainText: game.Name,
		}},
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
		Files: []notionapi.File{{
			Name: game.HeaderImage,
			Type: notionapi.FileTypeExternal,
			External: &notionapi.FileObject{
				URL: game.HeaderImage,
			},
		}},
	}
	properties["Release Date"] = &notionapi.DateProperty{
		Date: &notionapi.DateObject{
			Start: &notionReleaseDate,
		},
	}

	_, err := nc.client.CreatePage(&notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: notionapi.DatabaseID(gameDB),
		},
		Properties: properties,
	})
	if err != nil {
		return fmt.Errorf("failed to add game %s to database id %s: %s", game.Name, gameDB, err.Error())
	}

	// gp := GameProperties{
	// 	Name:              page.Properties["Name"].(*notionapi.TitleProperty),
	// 	Status:            page.Properties["Status"].(*notionapi.StatusProperty),
	// 	Tags:              page.Properties["Tags"].(*notionapi.MultiSelectProperty),
	// 	OfficialStorePage: page.Properties["Official Store Page"].(*notionapi.URLProperty),
	// 	CompletedDate:     page.Properties["Completed Date"].(*notionapi.DateProperty),
	// 	CoverArt:          page.Properties["Cover Art"].(*notionapi.FilesProperty),
	// 	Platform:          page.Properties["Platform"].(*notionapi.MultiSelectProperty),
	// 	ReleaseDate:       page.Properties["Release Date"].(*notionapi.DateProperty),
	// 	Rating:            page.Properties["Rating"].(*notionapi.RichTextProperty),
	// 	Notes:             page.Properties["Notes"].(*notionapi.RichTextProperty),
	// }
	// PrintGameProperties(gp)

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

func determineGameStatus(game steam.SteamGame) string {
	upNextVal, upNextOk := game.Collections[steam.CollectionUpNext]
	playingVal, playingOk := game.Collections[steam.CollectionPlaying]
	finishedVal, finishedOk := game.Collections[steam.CollectionFinished]

	if finishedOk && finishedVal {
		return StatusFinished
	} else if playingOk && playingVal {
		return StatusPlaying
	} else if upNextOk && upNextVal {
		return StatusUpNext
	} else if game.ReleaseDate.Before(time.Now()) {
		return StatusUnowned
	} else {
		return StatusUnreleased
	}
}

func setQueryOptions(options *notionapi.DatabaseQueryRequest) *notionapi.DatabaseQueryRequest {
	if options == nil || options == (&notionapi.DatabaseQueryRequest{}) { // default to returning all games
		return &notionapi.DatabaseQueryRequest{
			PageSize: 100,
			Sorts:    []notionapi.SortObject{{Property: "Name", Direction: "ascending"}},
		}
	}

	if options.PageSize == 0 { // default to pages of size 100
		options.PageSize = 100
	}
	if len(options.Sorts) == 0 { // default to sort by Name ascending
		options.Sorts = []notionapi.SortObject{{Property: "Name", Direction: "ascending"}}
	}

	return options
}
