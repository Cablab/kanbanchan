package ext

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kanbanchan/internal"
	"net/http"
)

const (
	baseURL = "https://store.steampowered.com/"
)

type SteamInterface interface {
	GetWishlist() error
}

type SteamClient struct {
	ctx      context.Context
	steamKey string
	steamID  string
}

type WishlistApp struct {
	Name        string      `json:"name"`
	Capsule     string      `json:"capsule"`
	ReleaseDate json.Number `json:"release_date"`
	Type        string      `json:"type"`
	Tags        []string    `json:"tags"`
}

func NewSteamClient(ctx context.Context) (*SteamClient, error) {
	var client SteamClient
	var secrets, err = internal.GetSecrets()
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
	endpoint := fmt.Sprintf("wishlist/profiles/%s/wishlistdata/", sc.steamID)
	resp, err := http.Get(fmt.Sprintf("%s%s", baseURL, endpoint))
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

func getHeaderImage(appID string) string {
	return fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/steam/apps/%s/header.jpg", appID)
}
