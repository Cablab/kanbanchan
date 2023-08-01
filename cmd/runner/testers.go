package main

import (
	"context"
	"fmt"
	"kanbanchan/internal/aws"
	"kanbanchan/internal/notion"
	"kanbanchan/internal/steam"
)

func testSuite() {
	// testNotionDatabaseProperties()
	testNotionDatabasePages()
	// testSteamWishlist()
	// testSteamLibrary()
	// testSteamApp()
}

func testNotionDatabasePages() {
	nc, err := notion.NewClient(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	games, err := nc.GetGamePages()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("Number of found games:", len(games))
	notion.PrintGameProperties(games[0])
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

	db, err := nc.GetDatabase(secrets.Notion.GameDB)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(db.Properties)
	for key, pc := range db.Properties {
		fmt.Printf("Key: %s, PropertyConfig: %v, PropertyConfig Type: %T\n", key, pc, pc)
	}
}

func testSteamWishlist() {
	sc, err := steam.NewClient(context.Background())
	if err != nil {
		fmt.Println("error making steam client:", err.Error())
		return
	}

	wishlist, err := sc.GetWishlist()
	if err != nil {
		fmt.Println("error getting steam wishlist:", err.Error())
		return
	}

	// fmt.Println((*wishlist))
	fmt.Println(len(wishlist))
	fmt.Println(wishlist[0])
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

	game := gameLibrary[0]

	fmt.Printf("Release Date: %s\n", game.ReleaseDate)
	fmt.Println("Games in library:", len(gameLibrary))
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
