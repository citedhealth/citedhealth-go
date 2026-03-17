package citedhealth_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	citedhealth "github.com/citedhealth/citedhealth-go"
)

// helper to create a paginated JSON response body.
func paginatedJSON(t *testing.T, results interface{}) []byte {
	t.Helper()
	raw, err := json.Marshal(results)
	if err != nil {
		t.Fatalf("marshal results: %v", err)
	}
	body := map[string]interface{}{
		"count":    1,
		"next":     nil,
		"previous": nil,
		"results":  json.RawMessage(raw),
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal paginated: %v", err)
	}
	return b
}

func TestNew(t *testing.T) {
	c := citedhealth.New()
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWithBaseURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 1, "name": "Biotin", "slug": "biotin",
		})
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	ing, err := c.GetIngredient(context.Background(), "biotin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Name != "Biotin" {
		t.Errorf("expected Biotin, got %s", ing.Name)
	}
}

func TestWithTimeout(t *testing.T) {
	c := citedhealth.New(citedhealth.WithTimeout(5 * time.Second))
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestSearchIngredients(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ingredients/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if q := r.URL.Query().Get("q"); q != "biotin" {
			t.Errorf("expected q=biotin, got q=%s", q)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedJSON(t, []map[string]interface{}{
			{"id": 1, "name": "Biotin", "slug": "biotin", "category": "vitamins"},
		}))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	results, err := c.SearchIngredients(context.Background(), "biotin", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Slug != "biotin" {
		t.Errorf("expected slug biotin, got %s", results[0].Slug)
	}
}

func TestSearchIngredientsWithCategory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cat := r.URL.Query().Get("category"); cat != "minerals" {
			t.Errorf("expected category=minerals, got category=%s", cat)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedJSON(t, []map[string]interface{}{
			{"id": 2, "name": "Magnesium", "slug": "magnesium", "category": "minerals"},
		}))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	results, err := c.SearchIngredients(context.Background(), "", "minerals")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Name != "Magnesium" {
		t.Errorf("expected Magnesium, got %s", results[0].Name)
	}
}

func TestGetIngredient(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ingredients/biotin/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":                 1,
			"name":               "Biotin",
			"slug":               "biotin",
			"category":           "vitamins",
			"mechanism":          "Cofactor for carboxylase enzymes",
			"recommended_dosage": map[string]string{"general": "2.5-5mg"},
			"forms":              []string{"capsule", "tablet"},
			"is_featured":        true,
		})
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	ing, err := c.GetIngredient(context.Background(), "biotin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ing.Name != "Biotin" {
		t.Errorf("expected Biotin, got %s", ing.Name)
	}
	if ing.Category != "vitamins" {
		t.Errorf("expected vitamins, got %s", ing.Category)
	}
	if !ing.IsFeatured {
		t.Error("expected is_featured=true")
	}
	if len(ing.Forms) != 2 {
		t.Errorf("expected 2 forms, got %d", len(ing.Forms))
	}
	if dosage, ok := ing.RecommendedDosage["general"]; !ok || dosage != "2.5-5mg" {
		t.Errorf("expected recommended_dosage general=2.5-5mg, got %v", ing.RecommendedDosage)
	}
}

func TestGetEvidence(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/evidence/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if ing := r.URL.Query().Get("ingredient"); ing != "biotin" {
			t.Errorf("expected ingredient=biotin, got %s", ing)
		}
		if cond := r.URL.Query().Get("condition"); cond != "hair-loss" {
			t.Errorf("expected condition=hair-loss, got %s", cond)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedJSON(t, []map[string]interface{}{
			{
				"id":                 1,
				"ingredient":         map[string]string{"slug": "biotin", "name": "Biotin"},
				"condition":          map[string]string{"slug": "hair-loss", "name": "Hair Loss"},
				"grade":              "A",
				"grade_label":        "Strong Evidence",
				"summary":            "Multiple RCTs show positive results",
				"direction":          "positive",
				"total_studies":      12,
				"total_participants": 1500,
			},
		}))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	ev, err := c.GetEvidence(context.Background(), "biotin", "hair-loss")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Grade != "A" {
		t.Errorf("expected grade A, got %s", ev.Grade)
	}
	if ev.TotalStudies != 12 {
		t.Errorf("expected 12 studies, got %d", ev.TotalStudies)
	}
	if ev.Ingredient.Slug != "biotin" {
		t.Errorf("expected ingredient slug biotin, got %s", ev.Ingredient.Slug)
	}
	if ev.Condition.Name != "Hair Loss" {
		t.Errorf("expected condition Hair Loss, got %s", ev.Condition.Name)
	}
}

func TestGetEvidenceNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedJSON(t, []interface{}{}))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	_, err := c.GetEvidence(context.Background(), "unknown", "unknown")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *citedhealth.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
	if notFound.Resource != "evidence" {
		t.Errorf("expected resource=evidence, got %s", notFound.Resource)
	}
}

func TestGetEvidenceByID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/evidence/5/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":          5,
			"grade":       "C",
			"grade_label": "Some Evidence",
			"ingredient":  map[string]string{"slug": "zinc", "name": "Zinc"},
			"condition":   map[string]string{"slug": "acne", "name": "Acne"},
		})
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	ev, err := c.GetEvidenceByID(context.Background(), 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev.Grade != "C" {
		t.Errorf("expected grade C, got %s", ev.Grade)
	}
	if ev.ID != 5 {
		t.Errorf("expected id 5, got %d", ev.ID)
	}
}

func TestSearchPapers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/papers/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if q := r.URL.Query().Get("q"); q != "biotin hair" {
			t.Errorf("expected q=biotin hair, got q=%s", q)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedJSON(t, []map[string]interface{}{
			{"id": 1, "pmid": "11111111", "title": "Biotin and Hair Growth"},
			{"id": 2, "pmid": "22222222", "title": "Biotin Supplementation Review"},
		}))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	papers, err := c.SearchPapers(context.Background(), "biotin hair", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(papers) != 2 {
		t.Fatalf("expected 2 papers, got %d", len(papers))
	}
	if papers[0].PMID != "11111111" {
		t.Errorf("expected pmid 11111111, got %s", papers[0].PMID)
	}
}

func TestSearchPapersWithYear(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if y := r.URL.Query().Get("year"); y != "2023" {
			t.Errorf("expected year=2023, got year=%s", y)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(paginatedJSON(t, []map[string]interface{}{
			{"id": 3, "pmid": "33333333", "title": "Recent Study"},
		}))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	year := 2023
	papers, err := c.SearchPapers(context.Background(), "", &year)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(papers) != 1 {
		t.Fatalf("expected 1 paper, got %d", len(papers))
	}
}

func TestGetPaper(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/papers/12345678/" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		pubYear := 2022
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":               1,
			"pmid":             "12345678",
			"title":            "Biotin for Hair Loss",
			"journal":          "J Dermatology",
			"publication_year": pubYear,
			"study_type":       "rct",
			"citation_count":   42,
			"is_open_access":   true,
			"pubmed_link":      "https://pubmed.ncbi.nlm.nih.gov/12345678/",
		})
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	paper, err := c.GetPaper(context.Background(), "12345678")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if paper.PMID != "12345678" {
		t.Errorf("expected pmid 12345678, got %s", paper.PMID)
	}
	if paper.CitationCount != 42 {
		t.Errorf("expected 42 citations, got %d", paper.CitationCount)
	}
	if !paper.IsOpenAccess {
		t.Error("expected is_open_access=true")
	}
	if paper.PublicationYear == nil || *paper.PublicationYear != 2022 {
		t.Errorf("expected publication_year=2022, got %v", paper.PublicationYear)
	}
}

func TestNotFoundError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"detail":"Not found."}`))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	_, err := c.GetIngredient(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var notFound *citedhealth.NotFoundError
	if !errors.As(err, &notFound) {
		t.Fatalf("expected NotFoundError, got %T: %v", err, err)
	}
}

func TestRateLimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"detail":"Rate limit exceeded."}`))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	_, err := c.SearchIngredients(context.Background(), "test", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var rateLimit *citedhealth.RateLimitError
	if !errors.As(err, &rateLimit) {
		t.Fatalf("expected RateLimitError, got %T: %v", err, err)
	}
	if rateLimit.RetryAfter != 60 {
		t.Errorf("expected RetryAfter=60, got %d", rateLimit.RetryAfter)
	}
}

func TestGenericError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"detail":"Internal server error."}`))
	}))
	defer srv.Close()

	c := citedhealth.New(citedhealth.WithBaseURL(srv.URL))
	_, err := c.SearchIngredients(context.Background(), "test", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var chErr *citedhealth.CitedHealthError
	if !errors.As(err, &chErr) {
		t.Fatalf("expected CitedHealthError, got %T: %v", err, err)
	}
}
