package main

import (
	"context"
	"fmt"
	"kanbanchan/internal/aws"
	"kanbanchan/internal/notion"
	"kanbanchan/internal/steam"
)

func main() {
	// testNotionDatabaseProperties()
	// testNotionDatabasePages()
	testSteamWishlist()
	// testSteamLibrary()
	// testSteamApp()
}

// =======================================================

func testNotionDatabasePages() {
	secrets, err := aws.GetSecrets()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	nc, err := notion.NewClient(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = nc.GetDatabasePages(secrets.Notion.GameDB)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func testNotionDatabaseProperties() {
	secrets, err := aws.GetSecrets()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	nc, err := notion.NewClient(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	err = nc.GetDatabase(secrets.Notion.GameDB)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func testSteamWishlist() {
	sc, err := steam.NewClient(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wishlist, err := sc.GetWishlist()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	// fmt.Println((*wishlist))
	fmt.Println(len(*wishlist))
}

func testSteamLibrary() {
	sc, err := steam.NewClient(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	gameLibrary, err := sc.GetLibrary()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	game := (*gameLibrary)[0]

	fmt.Printf("Release Date: %s\n", game.ReleaseDate)
	fmt.Println("Games in library:", len(*gameLibrary))
}

func testSteamApp() {
	sc, err := steam.NewClient(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	app, err := sc.GetApp("445980")
	fmt.Printf("Name: %s\nRelease Date: %s\nGenres: ", app.Data.Name, app.Data.ReleaseDate.Date)
	for _, genre := range app.Data.Genres {
		fmt.Printf("%s, ", genre.Description)
	}
	fmt.Println()
}

func testKeys() {
	keys, err := aws.GetSecrets()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Discord: %s\nGoogle: %s\nNotion: %s\nSteam: %s\n",
		keys.Discord.Key, keys.Google.Key, keys.Notion, keys.Steam)
}
