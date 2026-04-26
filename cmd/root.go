// Package cmd defines the root Cobra command and CLI flags.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ali5ter/clu/config"
	"github.com/ali5ter/clu/internal/api"
	"github.com/ali5ter/clu/internal/cache"
	"github.com/ali5ter/clu/internal/pipeline"
	"github.com/ali5ter/clu/internal/tui"
)

var (
	flagJSON    bool
	flagOffline bool
	flagNoTTY   bool
	version     = "0.4.0"
)

var rootCmd = &cobra.Command{
	Use:     "clu",
	Short:   "Explore the commandlineuser.com CLI catalog from your terminal",
	Version: version,
	RunE:    run,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVar(&flagJSON, "json", false, "emit catalog as JSON to stdout")
	rootCmd.Flags().BoolVar(&flagOffline, "offline", false, "use cached data, skip live fetch")
	rootCmd.Flags().BoolVar(&flagNoTTY, "no-tty", false, "force pipeline mode even in a terminal")
}

func run(cmd *cobra.Command, _ []string) error {
	_, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
	}

	client := api.NewClient()

	// Pipeline mode: --json flag or stdout not a TTY — load data synchronously
	if flagJSON || flagNoTTY || !isTTY() {
		catalog, _, _, err := loadData(client)
		if err != nil {
			return err
		}
		return pipeline.EmitCatalog(catalog)
	}

	// TUI mode — pass fetch function so TUI can show loading screen
	finalModel, err := tui.Run(
		func() ([]api.CatalogItem, []api.Article, *api.About, error) {
			return loadData(client)
		},
		version,
		func() string {
			v, _ := client.FetchLatestVersion()
			return v
		},
	)
	if err != nil {
		return err
	}

	// Handle ^J: emit selected item JSON after TUI exits
	if it := finalModel.SelectedCatalogItem(); it != nil {
		return pipeline.EmitItem(it)
	}

	return nil
}

func loadData(client *api.Client) ([]api.CatalogItem, []api.Article, *api.About, error) {
	var (
		catalog  []api.CatalogItem
		articles []api.Article
		about    *api.About
	)

	if flagOffline {
		if err := cache.Read("catalog", &catalog); err != nil {
			return nil, nil, nil, fmt.Errorf("offline: no cached catalog: %w", err)
		}
		if err := cache.Read("articles", &articles); err != nil {
			return nil, nil, nil, fmt.Errorf("offline: no cached articles: %w", err)
		}
		var a api.About
		if err := cache.Read("about", &a); err != nil {
			return nil, nil, nil, fmt.Errorf("offline: no cached about: %w", err)
		}
		about = &a
		return catalog, articles, about, nil
	}

	var err error
	catalog, err = client.FetchCatalog()
	if err != nil {
		return nil, nil, nil, err
	}
	articles, err = client.FetchArticles()
	if err != nil {
		return nil, nil, nil, err
	}
	about, err = client.FetchAbout()
	if err != nil {
		return nil, nil, nil, err
	}

	// Persist for offline use
	_ = cache.Write("catalog", catalog)
	_ = cache.Write("articles", articles)
	_ = cache.Write("about", about)

	return catalog, articles, about, nil
}
