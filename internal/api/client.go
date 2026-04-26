package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	CatalogURL  = "https://commandlineuser.com/api/catalog.json"
	ArticlesURL = "https://commandlineuser.com/api/articles.json"
	AboutURL    = "https://commandlineuser.com/api/about.json"
	VersionURL  = "https://api.github.com/repos/ali5ter/clu/releases/latest"

	httpTimeout = 8 * time.Second
)

// Client fetches data from the commandlineuser.com API.
type Client struct {
	http *http.Client
}

// NewClient returns a Client with a sensible timeout.
func NewClient() *Client {
	return &Client{http: &http.Client{Timeout: httpTimeout}}
}

func (c *Client) fetch(url string, dst any) error {
	resp, err := c.http.Get(url)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", url, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read %s: %w", url, err)
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decode %s: %w", url, err)
	}
	return nil
}

// FetchCatalog retrieves the full catalog item list.
func (c *Client) FetchCatalog() ([]CatalogItem, error) {
	var items []CatalogItem
	return items, c.fetch(CatalogURL, &items)
}

// FetchArticles retrieves the feature articles list.
func (c *Client) FetchArticles() ([]Article, error) {
	var articles []Article
	return articles, c.fetch(ArticlesURL, &articles)
}

// FetchAbout retrieves site and author metadata.
func (c *Client) FetchAbout() (*About, error) {
	var about About
	return &about, c.fetch(AboutURL, &about)
}

// FetchLatestVersion retrieves the latest published release version, e.g. "1.2.3".
// Uses the GitHub Releases API so the version is always up to date with no extra
// server infrastructure required.
func (c *Client) FetchLatestVersion() (string, error) {
	var v struct {
		TagName string `json:"tag_name"`
	}
	if err := c.fetch(VersionURL, &v); err != nil {
		return "", err
	}
	return strings.TrimPrefix(v.TagName, "v"), nil
}
