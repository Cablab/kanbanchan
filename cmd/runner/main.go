package main

import (
	"context"
	"fmt"
	"kanbanchan/internal/notion"
	"kanbanchan/internal/steam"
	pkgnotion "kanbanchan/pkg/notion"
	"time"

	"github.com/jomei/notionapi"
)

type clients struct {
	steamClient  *steam.SteamClient
	notionClient *notion.NotionClient
}

func main() {
	// testSuite() // quick output sanity check testing stuff
	// =======================================================

	nc, err := notion.NewClient(context.Background())
	if err != nil {
		fmt.Printf("failed to create notion client: %s", err.Error())
		return
	}

	sc, err := steam.NewClient(context.Background())
	if err != nil {
		fmt.Printf("failed to create steam client: %s", err.Error())
		return
	}

	runner := clients{
		steamClient:  sc,
		notionClient: nc,
	}

	// err = runner.syncGames()
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	return
	// }

	err = runner.transitionGames()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func (c *clients) syncGames() error {
	library, err := c.steamClient.GetLibrary()
	if err != nil {
		return fmt.Errorf("failed to get steam library: %s", err.Error())
	}

	wishlist, err := c.steamClient.GetWishlist()
	if err != nil {
		return fmt.Errorf("failed to get steam wishlist: %s", err.Error())
	}

	notionGames, err := c.notionClient.GetGamePages(nil)
	if err != nil {
		return fmt.Errorf("failed to get notion games: %s", err.Error())
	}

	// For every game in library, check to see if its in notionGames. If not, add
	for _, game := range *library {
		_, ok := (*notionGames)[game.Name]
		if !ok {
			err := c.notionClient.AddGame(game)
			if err != nil {
				return fmt.Errorf("failed to add game %s: %s", game.Name, err.Error())
			}
		}
	}

	// For every game in wishlist, check to see if its in notionGames. If not, add
	for _, game := range *wishlist {
		_, ok := (*notionGames)[game.Name]
		if !ok {
			err := c.notionClient.AddGame(game)
			if err != nil {
				return fmt.Errorf("failed to add game %s: %s", game.Name, err.Error())
			}
		}
	}

	return nil
}

func (c *clients) transitionGames() error {
	options := &notionapi.DatabaseQueryRequest{
		Filter: notionapi.PropertyFilter{
			Property: "Status",
			Status: &notionapi.StatusFilterCondition{
				Equals: notion.StatusUnreleased,
			},
		},
	}

	games, err := c.notionClient.GetGamePages(options)
	if err != nil {
		return fmt.Errorf("failed to get unreleased games: %s", err.Error())
	}

	nc := pkgnotion.NotionClient{}
	for title, game := range *games {
		if game.ReleaseDate == nil || game.ReleaseDate.Date == nil || game.ReleaseDate.Date.Start == nil {
			continue
		}

		releaseDate, err := nc.ParseNotionDate(*game.ReleaseDate.Date.Start)
		if err != nil {
			return fmt.Errorf("failed to parse release date for \"%s\": %s", title, err.Error())
		}
		if time.Now().Before(releaseDate) {
			gamePage, err := c.notionClient.GetGamePageByID(game.PageID)
			if err != nil {
				return fmt.Errorf("failed to retrieve game \"%s\" by id %s: %s", game.Name.Title[0].PlainText, game.PageID, err.Error())
			}
			gamePage.Properties["Status"].(*notionapi.StatusProperty).Status.Name = notion.StatusUnowned

			err = c.notionClient.UpdateGame(game.PageID, gamePage.Properties)
			if err != nil {
				return fmt.Errorf("failed to transition %s to status Unowned: %s", game.Name.Title[0].PlainText, err.Error())
			}
		}
	}
	return nil
}
