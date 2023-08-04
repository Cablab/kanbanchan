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

// SteamClient contains authentication info for the Steam client
type SteamClient struct {
	ctx      context.Context
	steamKey string
}

// WishlistApp defines the data retrieved for an app on a user's wishlist
type WishlistApp struct {
	ID          string      `json:"id,omitempty"`
	Name        string      `json:"name"`
	Capsule     string      `json:"capsule"`
	ReleaseDate json.Number `json:"release_date"`
	Type        string      `json:"type"`
	Tags        []string    `json:"tags"`
}

// OwnedApps defines the response received from retrieving all of a user's owned apps
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

// SteamApp defines the response received from retrieving a specific Steam app by ID
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

// NewClient creates a new Steam client authenticated with the supplied steam key
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

// GetUserWishlist returns a list of apps on the specified user's wishlist
func (sc *SteamClient) GetUserWishlist(steamUserID string) ([]WishlistApp, error) {
	var wishlist []WishlistApp
	i := 0
	for {
		var wishlistPage map[string]WishlistApp
		endpoint := fmt.Sprintf("/wishlist/profiles/%s/wishlistdata/?p=%d", steamUserID, i)
		resp, err := http.Get(fmt.Sprintf("%s%s", steamURL, endpoint))
		if err != nil {
			return nil, fmt.Errorf("failed to make http request to get wishlist for user id %s: %s", steamUserID, err.Error())
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %s", err.Error())
		}

		if string(body) == "[]" { // no entries returned
			break
		}

		err = json.Unmarshal(body, &wishlistPage)
		if err != nil {
			fmt.Println(string(body), resp.Status, resp.StatusCode)
			return nil, fmt.Errorf("failed to unmarshal response body: %s", err.Error())
		}

		for k, v := range wishlistPage {
			wishlistApp := v
			wishlistApp.ID = k
			wishlist = append(wishlist, wishlistApp)
		}

		i++
	}

	return wishlist, nil
}

// GetUserOwnedGames returns information about all games owned by the specified user
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

// GetApp retrieves steam app information for the specified app ID
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
