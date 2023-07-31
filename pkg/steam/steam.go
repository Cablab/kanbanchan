package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	steamAPIURL = "https://api.steampowered.com"
	steamURL    = "https://store.steampowered.com"
)

type SteamClient struct {
	ctx      context.Context
	steamKey string
}

type WishlistApp struct {
	Name        string      `json:"name"`
	Capsule     string      `json:"capsule"`
	ReleaseDate json.Number `json:"release_date"`
	Type        string      `json:"type"`
	Tags        []string    `json:"tags"`
}

type OwnedApps struct {
	Response struct {
		GameCount json.Number `json:"game_count"`
		Games     []struct {
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
		} `json:"games"`
	} `json:"response"`
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

func NewClient(ctx context.Context, steamKey string) (*SteamClient, error) {
	var client SteamClient
	if ctx == nil {
		client.ctx = context.Background()
	} else {
		client.ctx = ctx
	}
	key := strings.TrimSpace(steamKey)
	if key == "" {
		return nil, fmt.Errorf("empty steamKey provided")
	}
	client.steamKey = key
	return &client, nil
}

// TODO paginate with ?p=0 (page 0), ?p=1 (page 1), etc. at the end of the URL until nothing is returned (still 200, just empty array)
func (sc *SteamClient) GetUserWishlist(steamUserID string) (*map[string]WishlistApp, error) {
	var wishlist map[string]WishlistApp
	endpoint := fmt.Sprintf("/wishlist/profiles/%s/wishlistdata/", steamUserID)
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

func (sc *SteamClient) GetUserOwnedGames(steamUserID string) (*OwnedApps, error) {
	// Optional URL Params: &skip_unvetted_apps=false | &include_played_free_games=1 | &include_appinfo=1
	var ownedApps OwnedApps
	endpoint := fmt.Sprintf("/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&include_appinfo=1&include_played_free_games=1&skip_unvetted_apps=false&format=json", sc.steamKey, steamUserID)
	resp, err := http.Get(fmt.Sprintf("%s%s", steamAPIURL, endpoint))
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &ownedApps)
	if err != nil {
		return nil, err
	}

	return &ownedApps, nil
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
