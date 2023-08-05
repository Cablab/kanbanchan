package steam

import (
	"context"
	"encoding/json"
	"fmt"
	"kanbanchan/internal/aws"
	"kanbanchan/pkg/steam"
	"strings"
	"time"
)

const (
	CollectionFinished = "Finished"
	CollectionUpNext   = "Up Next"
	CollectionPlaying  = "Playing"

	steamAPIURL         = "https://api.steampowered.com"
	steamURL            = "https://store.steampowered.com"
	steamDateFormat     = "Jan 2, 2006"
	steamDateAltFormat  = "2 Jan, 2006"
	steamDateYearFormat = "2006"
)

// SteamClient contains auth info, a client, and manually tracked Library Collections
type SteamClient struct {
	steam       steam.SteamClient
	steamKey    string
	steamID     string
	collections struct {
		finished []string
		upNext   []string
		playing  []string
	}
}

// SteamGame contains info about a Steam Game
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

// SteamApp contains info about a Steam App
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

// WishlistApp contains info about an app on a Wishlist
type WishlistApp struct {
	Name        string      `json:"name"`
	Capsule     string      `json:"capsule"`
	ReleaseDate json.Number `json:"release_date"`
	Type        string      `json:"type"`
	Tags        []string    `json:"tags"`
}

// LibraryApp contains info about an app in a user's Library
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

// Library contains info about games owned by a user
type Library struct {
	Response struct {
		GameCount json.Number  `json:"game_count"`
		Games     []LibraryApp `json:"games"`
	} `json:"response"`
}

// NewClient creates an authenticated client and sets up manually tracked
// library Collections (since those can't be retrieved via API yet)
func NewClient(ctx context.Context) (*SteamClient, error) {
	var client SteamClient
	var secrets, err = aws.GetSecrets()
	if err != nil {
		return nil, err
	}
	client.steamID = secrets.Steam.ID
	client.steamKey = secrets.Steam.Key
	steamClient, err := steam.NewClient(context.Background(), client.steamKey)
	if err != nil {
		return nil, err
	}
	client.steam = *steamClient

	var finished, upNext, playing []string
	for _, jnum := range secrets.Steam.Collections.Finished {
		finished = append(finished, jnum.String())
	}
	for _, jnum := range secrets.Steam.Collections.UpNext {
		upNext = append(upNext, jnum.String())
	}
	for _, jnum := range secrets.Steam.Collections.Playing {
		playing = append(playing, jnum.String())
	}
	client.collections.finished = finished
	client.collections.upNext = upNext
	client.collections.playing = playing
	return &client, nil
}

// GetWishlist gets all games on the authenticated user's wishlist with extra
// data populated by getting the Steam App info
func (sc *SteamClient) GetWishlist() (*map[string]SteamGame, error) {
	wishlist, err := sc.steam.GetUserWishlist(sc.steamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get wishlist for user id %s: %s", sc.steamID, err.Error())
	}

	games := make(map[string]SteamGame)
	for _, wishlistApp := range wishlist {
		steamApp, err := sc.steam.GetApp(wishlistApp.ID)
		if err != nil {
			return nil, err
		}
		var genres []string
		for _, genre := range steamApp.Data.Genres {
			genres = append(genres, genre.Description)
		}
		game := SteamGame{
			ID:          wishlistApp.ID,
			Name:        steamApp.Data.Name,
			HeaderImage: steamApp.Data.HeaderImage,
			Genres:      genres,
		}
		if strings.ToLower(steamApp.Data.ReleaseDate.Date) != "to be announced" &&
			steamApp.Data.ReleaseDate.Date != "" {
			releaseDate, err := ParseSteamDate(steamApp.Data.ReleaseDate.Date)
			if err != nil {
				return nil, err
			}
			game.ReleaseDate = releaseDate
		}
		_, ok := games[game.Name]
		if !ok {
			games[game.Name] = game
		}
	}

	return &games, nil
}

// GetLibrary gets all games owned by the authenticated user with extra
// data populated by getting the Steam App info
func (sc *SteamClient) GetLibrary() (*map[string]SteamGame, error) {
	library, err := sc.steam.GetUserOwnedGames(sc.steamID)
	if err != nil {
		return nil, err
	}

	collectionMap := make(map[string]map[string]bool)
	for _, game := range library.Response.Games {
		collections, err := libraryCollectionCheck(sc, game.AppID.String())
		if err != nil {
			return nil, err
		}
		collectionMap[game.AppID.String()] = collections
	}

	games := make(map[string]SteamGame)
	for _, game := range library.Response.Games {
		if len(collectionMap[game.AppID.String()]) > 0 {
			steamApp, err := sc.steam.GetApp(string(game.AppID))
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
				Collections:              collectionMap[game.AppID.String()],
			}
			_, ok := games[game.Name]
			if !ok {
				games[game.Name] = game
			}
		}
	}

	return &games, nil
}

// GetApp gets a Steam App
func (sc *SteamClient) GetApp(appID string) (*steam.SteamApp, error) {
	steamApp, err := sc.steam.GetApp(appID)
	if err != nil {
		return nil, err
	}
	return steamApp, nil
}

// GetAppByName gets a Steam App
func (sc *SteamClient) GetAppByName(appName string) (*steam.SteamApp, error) {
	steamApp, err := sc.steam.GetAppByName(appName)
	if err != nil {
		return nil, err
	}
	return steamApp, nil
}

// ParseSteamDate attempts to parse various formats of dates used by Steam apps
func ParseSteamDate(steamDate string) (time.Time, error) {
	date, err := time.Parse(steamDateFormat, steamDate)
	if err != nil {
		// fmt.Println("failed to parse release date in standard format; attempting alt format...")
	} else {
		return date, nil
	}

	date, err = time.Parse(steamDateAltFormat, steamDate)
	if err != nil {
		// fmt.Println("failed to parse release date in alt format; attempting year format...")
	} else {
		return date, nil
	}

	date, err = time.Parse(steamDateYearFormat, steamDate)
	if err != nil {
		// fmt.Println("failed to parse release date in year format")
		return time.Time{}, nil
	} else {
		return date, nil
	}
}

// ParsePlaytime parses a JSON timestamp into a Golang time.Time type
func ParsePlaytime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// libraryCollectionCheck checks which manually tracked Library Collections a game is in
func libraryCollectionCheck(sc *SteamClient, appID string) (map[string]bool, error) {
	collections := make(map[string]bool)
	for _, id := range sc.collections.finished {
		if id == appID {
			collections[CollectionFinished] = true
			break
		}
	}
	for _, id := range sc.collections.upNext {
		if id == appID {
			collections[CollectionUpNext] = true
			break
		}
	}
	for _, id := range sc.collections.playing {
		if id == appID {
			collections[CollectionPlaying] = true
			break
		}
	}
	return collections, nil
}
