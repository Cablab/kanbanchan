package main

import (
	"context"
	"fmt"
	"kanbanchan/internal/notion"
	"kanbanchan/internal/steam"
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

	err = runner.syncGames()
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

	notionGames, err := c.notionClient.GetGamePages()
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
