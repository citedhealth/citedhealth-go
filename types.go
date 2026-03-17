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
	Slug string `json:"slug"`
	Name string `json:"name"`
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
