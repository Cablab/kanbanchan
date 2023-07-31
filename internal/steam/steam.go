package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kanbanchan/internal/aws"
	"net/http"
	"time"
)

const (
	steamAPIURL        = "https://api.steampowered.com"
	steamURL           = "https://store.steampowered.com"
	steamDateFormat    = "Jan 2, 2006"
	steamDateAltFormat = "2 Jan, 2006"
)

type SteamClient struct {
	ctx         context.Context
	steamKey    string
	steamID     string
	collections struct {
		completed []string
		upNext    []string
		playing   []string
	}
}

type SteamGame struct {
	ID                       string          `json:"id,omitempty"`
	Name                     string          `json:"name,omitempty"`
	HeaderImage              string          `json:"header_image,omitempty"`
	Genres                   []string        `json:"genres,omitempty"`
	ReleaseDate              time.Time       `json:"releaseDate,omitempty"`
	Playtime                 time.Time       `json:"playtime_forever,omitempty"`
	PlaytimeWindows          time.Time       `json:"playtime_windows_forever,omitempty"`
	PlaytimeMac              time.Time       `json:"playtime_mac_forever,omitempty"`
	PlaytimeLinux            time.Time       `json:"playtime_linux_forever,omitempty"`
	PlaytimeDisconnected     time.Time       `json:"playtime_disconnected,omitempty"`
	LastPlayed               time.Time       `json:"rtime_last_played,omitempty"`
	HasCommunityVisibleStats bool            `json:"has_community_visible_stats,omitempty"`
	Collections              map[string]bool `json:"collections,omitempty"`
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
	AppID                    json.Number     `json:"appid"`
	Name                     string          `json:"name"`
	Playtime                 json.Number     `json:"playtime_forever"`
	PlaytimeWindows          json.Number     `json:"playtime_windows_forever"`
	PlaytimeMac              json.Number     `json:"playtime_mac_forever"`
	PlaytimeLinux            json.Number     `json:"playtime_linux_forever"`
	PlaytimeDisconnected     json.Number     `json:"playtime_disconnected"`
	IconURL                  string          `json:"img_icon_url"`
	LastPlayed               json.Number     `json:"rtime_last_played"`
	HasCommunityVisibleStats bool            `json:"has_community_visible_stats,omitempty"`
	Collections              map[string]bool `json:"collections,omitempty"`
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
	var completed, upNext, playing []string
	for _, jnum := range secrets.Steam.Collections.Completed {
		completed = append(completed, jnum.String())
	}
	for _, jnum := range secrets.Steam.Collections.UpNext {
		upNext = append(upNext, jnum.String())
	}
	for _, jnum := range secrets.Steam.Collections.Playing {
		playing = append(playing, jnum.String())
	}
	client.collections.completed = completed
	client.collections.upNext = upNext
	client.collections.playing = playing
	return &client, nil
}

func (sc *SteamClient) GetWishlist() (*[]SteamGame, error) {
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

	var games []SteamGame
	for id, _ := range wishlist {
		steamApp, err := sc.GetApp(string(id))
		if err != nil {
			return nil, err
		}
		var genres []string
		for _, genre := range steamApp.Data.Genres {
			genres = append(genres, genre.Description)
		}
		releaseDate, err := ParseSteamDate(steamApp.Data.ReleaseDate.Date)
		if err != nil {
			return nil, err
		}
		game := SteamGame{
			ID:          id,
			Name:        steamApp.Data.Name,
			HeaderImage: steamApp.Data.HeaderImage,
			Genres:      genres,
			ReleaseDate: releaseDate,
		}
		games = append(games, game)
	}

	return &games, nil
}

func (sc *SteamClient) GetLibrary() (*[]SteamGame, error) {
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

	for i, game := range library.Response.Games {
		collections, err := libraryCollectionCheck(sc, string(game.AppID))
		if err != nil {
			return nil, err
		}
		library.Response.Games[i].Collections = collections
	}

	// LibraryApp Only: Playtime, PlaytimeWindows, PlaytimeMac, PlaytimeLinux, PlaytimeDisconnected, LastPlayed, HasCommunityVisibleStats, Collections
	// SteamApp Only: HeaderImage, Genres, Release Date
	// Both (default SteamApp): ID, Name
	var games []SteamGame
	for _, game := range library.Response.Games {
		if len(game.Collections) > 0 {
			steamApp, err := sc.GetApp(string(game.AppID))
			if err != nil {
				return nil, err
			}
			var genres []string
			for _, genre := range steamApp.Data.Genres {
				genres = append(genres, genre.Description)
			}
			releaseDate, err := ParseSteamDate(steamApp.Data.ReleaseDate.Date)
			if err != nil {
				return nil, err
			}
			iPlaytime, err := game.Playtime.Int64()
			if err != nil {
				return nil, err
			}
			iPlaytimeWindows, err := game.PlaytimeWindows.Int64()
			if err != nil {
				return nil, err
			}
			iPlaytimeMac, err := game.PlaytimeMac.Int64()
			if err != nil {
				return nil, err
			}
			iPlaytimeLinux, err := game.PlaytimeLinux.Int64()
			if err != nil {
				return nil, err
			}
			iPlaytimeDisconnected, err := game.PlaytimeDisconnected.Int64()
			if err != nil {
				return nil, err
			}
			iLastPlayed, err := game.LastPlayed.Int64()
			if err != nil {
				return nil, err
			}
			game := SteamGame{
				ID:                       string(steamApp.Data.AppID),
				Name:                     steamApp.Data.Name,
				HeaderImage:              steamApp.Data.HeaderImage,
				Genres:                   genres,
				ReleaseDate:              releaseDate,
				Playtime:                 ParsePlaytime(iPlaytime),
				PlaytimeWindows:          ParsePlaytime(iPlaytimeWindows),
				PlaytimeMac:              ParsePlaytime(iPlaytimeMac),
				PlaytimeLinux:            ParsePlaytime(iPlaytimeLinux),
				PlaytimeDisconnected:     ParsePlaytime(iPlaytimeDisconnected),
				LastPlayed:               ParsePlaytime(iLastPlayed),
				HasCommunityVisibleStats: game.HasCommunityVisibleStats,
				Collections:              game.Collections,
			}
			games = append(games, game)
		}
	}

	// fmt.Printf("Game Count: %s\n", library.Response.GameCount)
	// for _, game := range library.Response.Games {
	// 	fmt.Printf("App ID: %s\tGame Name: %s\tPlaytime: %s minutes\n", game.AppID, game.Name, game.Playtime)
	// 	// fmt.Printf("%s , %s\n", game.Name, game.AppID) // Use for easy app ID searching for manual collections
	// }
	return &games, nil
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

func ParseSteamDate(steamDate string) (time.Time, error) {
	date, err := time.Parse(steamDateFormat, steamDate)
	if err != nil {
		date, err = time.Parse(steamDateAltFormat, steamDate)
		if err != nil {
			return time.Time{}, err
		}
	}
	return date, nil
}

func ParsePlaytime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func libraryCollectionCheck(sc *SteamClient, appID string) (map[string]bool, error) {
	collections := make(map[string]bool)
	for _, id := range sc.collections.completed {
		if id == appID {
			collections["Completed"] = true
			break
		}
	}
	for _, id := range sc.collections.upNext {
		if id == appID {
			collections["UpNext"] = true
			break
		}
	}
	for _, id := range sc.collections.playing {
		if id == appID {
			collections["Playing"] = true
			break
		}
	}
	return collections, nil
}

func getHeaderImage(appID string) string {
	return fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/steam/apps/%s/header.jpg", appID)
}
