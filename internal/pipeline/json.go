// Package pipeline handles non-TTY JSON output mode.
package pipeline

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ali5ter/clu/internal/api"
)

// EmitCatalog writes the full catalog as JSON to stdout.
func EmitCatalog(items []api.CatalogItem) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(items); err != nil {
		return fmt.Errorf("encode catalog: %w", err)
	}
	return nil
}

// EmitItem writes a single catalog item as JSON to stdout.
func EmitItem(item *api.CatalogItem) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(item); err != nil {
		return fmt.Errorf("encode item: %w", err)
	}
	return nil
}
