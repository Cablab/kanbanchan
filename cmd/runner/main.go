package main

import (
	"context"
	"fmt"
	"kanbanchan/ext"
	"kanbanchan/internal"
)

func main() {
	// testNotionDatabaseProperties()
	testNotionDatabasePages()
}

// =======================================================

func testNotionDatabasePages() {
	secrets, err := internal.GetSecrets()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	nc, err := ext.NewNotionClient(context.Background())
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
	secrets, err := internal.GetSecrets()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	nc, err := ext.NewNotionClient(context.Background())
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
	sc, err := ext.NewSteamClient(context.Background())
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	wishlist, err := sc.GetWishlist()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println((*wishlist))
}

func testKeys() {
	keys, err := internal.GetSecrets()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Discord: %s\nGoogle: %s\nNotion: %s\nSteam: %s\n",
		keys.Discord.Key, keys.Google.Key, keys.Notion, keys.Steam)
}
