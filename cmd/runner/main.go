package main

import (
	"fmt"
	"kanbanchan/internal"
)

func main() {
	keys, err := internal.GetSecrets()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Printf("Discord: %s\nGoogle: %s\nNotion: %s\nSteam: %s\n",
		keys.Discord, keys.Google, keys.Notion, keys.Steam)
}
