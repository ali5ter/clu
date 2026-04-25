// Package api provides types and a client for the commandlineuser.com data API.
package api

// CatalogItem represents a single entry in the CLI tool catalog.
type CatalogItem struct {
	Name        string   `json:"name"`
	Summary     string   `json:"summary"`
	SourceURL   string   `json:"source_url"`
	SourceType  string   `json:"source_type"`
	Score       float64  `json:"score"`
	Confidence  float64  `json:"confidence"`
	Maturity    string   `json:"maturity"`
	Maintenance string   `json:"maintenance"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	AuthorOrOrg string   `json:"author_or_org"`
	License     string   `json:"license"`
	Trend       *string  `json:"trend"`
	IsNew       bool     `json:"is_new"`
}

// Article represents a feature article.
type Article struct {
	Slug      string   `json:"slug"`
	Title     string   `json:"title"`
	Summary   string   `json:"summary"`
	Published string   `json:"published"`
	Tags      []string `json:"tags"`
	URL       string   `json:"url"`
	Body      string   `json:"body"`
}

// About holds site and author metadata.
type About struct {
	Site struct {
		Description    string `json:"description"`
		AboutURL       string `json:"about_url"`
		MethodologyURL string `json:"methodology_url"`
		FeaturesURL    string `json:"features_url"`
	} `json:"site"`
	Author struct {
		Bio   []string `json:"bio"`
		Links struct {
			GitHub string `json:"github"`
		} `json:"links"`
	} `json:"author"`
	Catalogue struct {
		Description string `json:"description"`
		Curation    string `json:"curation"`
	} `json:"catalogue"`
	Methodology struct {
		ScoreFormula      string            `json:"score_formula"`
		ConfidenceFormula string            `json:"confidence_formula"`
		Components        map[string]string `json:"components"`
		Selection         string            `json:"selection"`
	} `json:"methodology"`
}
