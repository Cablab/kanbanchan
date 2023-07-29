package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kanbanchan/internal/aws"
	"net/http"
)

const (
	steamURL    = "https://store.steampowered.com"
	steamAPIURL = "https://api.steampowered.com"
)

type SteamClient struct {
	ctx      context.Context
	steamKey string
	steamID  string
}

type SteamApp struct {
	Success bool `json:"success"`
	Data    struct {
		Type        string      `json:"type"`
		Name        string      `json:"name"`
		AppID       json.Number `json:"steam_appid"`
		HeaderImage string      `json:"header_image"`
		Genres      []struct {
			ID          string `json:"id"`
			Description string `json:"description"`
		} `json:"genres"`
		ReleaseDate struct {
			Date string `json:"date"`
		} `json:"release_date"`
	} `json:"data"`
}

type WishlistApp struct {
	Name        string      `json:"name"`
	Capsule     string      `json:"capsule"`
	ReleaseDate json.Number `json:"release_date"`
	Type        string      `json:"type"`
	Tags        []string    `json:"tags"`
}

type LibraryApp struct {
	AppID                    json.Number `json:"appid"`
	Name                     string      `json:"name"`
	Playtime                 json.Number `json:"playtime_forever"`
	PlaytimeWindows          json.Number `json:"playtime_windows_forever"`
	PlaytimeMac              json.Number `json:"playtime_mac_forever"`
	PlaytimeLinux            json.Number `json:"playtime_linux_forever"`
	PlaytimeDisconnected     json.Number `json:"playtime_disconnected"`
	IconURL                  string      `json:"img_icon_url"`
	LastPlayed               json.Number `json:"rtime_last_played"`
	HasCommunityVisibleStats bool        `json:"has_community_visible_stats,omitempty"`
}

type Library struct {
	Response struct {
		GameCount json.Number  `json:"game_count"`
		Games     []LibraryApp `json:"games"`
	} `json:"response"`
}

func NewClient(ctx context.Context) (*SteamClient, error) {
	var client SteamClient
	var secrets, err = aws.GetSecrets()
	if err != nil {
		return nil, err
	}

	if ctx == nil {
		client.ctx = context.Background()
	} else {
		client.ctx = ctx
	}

	client.steamID = secrets.Steam.ID
	client.steamKey = secrets.Steam.Key
	return &client, nil
}

func (sc *SteamClient) GetWishlist() (*map[string]WishlistApp, error) {
	var wishlist map[string]WishlistApp
	endpoint := fmt.Sprintf("/wishlist/profiles/%s/wishlistdata/", sc.steamID)
	resp, err := http.Get(fmt.Sprintf("%s%s", steamURL, endpoint))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &wishlist)
	if err != nil {
		return nil, err
	}

	return &wishlist, nil
}

func (sc *SteamClient) GetLibrary() (*[]LibraryApp, error) {
	// Optional URL Params: &skip_unvetted_apps=false | &include_played_free_games=1 | &include_appinfo=1
	var library Library
	endpoint := fmt.Sprintf("/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&include_appinfo=1&include_played_free_games=1&skip_unvetted_apps=false&format=json", sc.steamKey, sc.steamID)
	resp, err := http.Get(fmt.Sprintf("%s%s", steamAPIURL, endpoint))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &library)
	if err != nil {
		return nil, err
	}

	// fmt.Printf("Game Count: %s\n", library.Response.GameCount)
	// for _, game := range library.Response.Games {
	// 	fmt.Printf("App ID: %s\tGame Name: %s\tPlaytime: %s minutes\n", game.AppID, game.Name, game.Playtime)
	// 	// fmt.Printf("%s , %s\n", game.Name, game.AppID) // Use for easy app ID searching for manual collections
	// }
	return &library.Response.Games, nil
}

func (sc *SteamClient) GetApp(appID string) (*SteamApp, error) {
	var app map[string]SteamApp
	endpoint := fmt.Sprintf("/api/appdetails?appids=%s", appID)
	resp, err := http.Get(fmt.Sprintf("%s%s", steamURL, endpoint))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &app)
	if err != nil {
		return nil, err
	}

	steamApp := app[appID]
	return &steamApp, nil
}

func getHeaderImage(appID string) string {
	return fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/steam/apps/%s/header.jpg", appID)
}
