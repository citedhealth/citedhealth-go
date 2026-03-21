package citedhealth

import "encoding/json"

// Ingredient represents a supplement ingredient.
type Ingredient struct {
	ID                int               `json:"id"`
	Name              string            `json:"name"`
	Slug              string            `json:"slug"`
	Category          string            `json:"category"`
	Mechanism         string            `json:"mechanism"`
	RecommendedDosage map[string]string `json:"recommended_dosage"`
	Forms             []string          `json:"forms"`
	IsFeatured        bool              `json:"is_featured"`
}

// Condition represents a health condition.
type Condition struct {
	Slug            string   `json:"slug"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	MetaDescription string   `json:"meta_description"`
	Prevalence      string   `json:"prevalence"`
	Symptoms        []string `json:"symptoms"`
	RiskFactors     []string `json:"risk_factors"`
	IsFeatured      bool     `json:"is_featured"`
}

// GlossaryTerm represents a glossary entry.
type GlossaryTerm struct {
	Slug            string `json:"slug"`
	Term            string `json:"term"`
	ShortDefinition string `json:"short_definition"`
	Definition      string `json:"definition"`
	Abbreviation    string `json:"abbreviation"`
	Category        string `json:"category"`
}

// Guide represents an educational guide article.
type Guide struct {
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	Content         string `json:"content"`
	Category        string `json:"category"`
	MetaDescription string `json:"meta_description"`
}

// Paper represents a PubMed-indexed paper.
type Paper struct {
	ID              int    `json:"id"`
	PMID            string `json:"pmid"`
	Title           string `json:"title"`
	Journal         string `json:"journal"`
	PublicationYear *int   `json:"publication_year"`
	StudyType       string `json:"study_type"`
	CitationCount   int    `json:"citation_count"`
	IsOpenAccess    bool   `json:"is_open_access"`
	PubMedLink      string `json:"pubmed_link"`
}

// NestedIngredient is a compact ingredient reference within evidence links.
type NestedIngredient struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// EvidenceLink represents the evidence for an ingredient-condition pair.
type EvidenceLink struct {
	ID                int              `json:"id"`
	Ingredient        NestedIngredient `json:"ingredient"`
	Condition         Condition        `json:"condition"`
	Grade             string           `json:"grade"`
	GradeLabel        string           `json:"grade_label"`
	Summary           string           `json:"summary"`
	Direction         string           `json:"direction"`
	TotalStudies      int              `json:"total_studies"`
	TotalParticipants int              `json:"total_participants"`
}

// PaginatedResponse represents a paginated API response.
type PaginatedResponse struct {
	Count    int              `json:"count"`
	Next     *string          `json:"next"`
	Previous *string          `json:"previous"`
	Results  json.RawMessage  `json:"results"`
}
