# citedhealth-go

[![Go Reference](https://pkg.go.dev/badge/github.com/citedhealth/citedhealth-go.svg)](https://pkg.go.dev/github.com/citedhealth/citedhealth-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Zero Dependencies](https://img.shields.io/badge/dependencies-0-brightgreen)](https://pkg.go.dev/github.com/citedhealth/citedhealth-go)

Go client for the [CITED Health](https://citedhealth.com) evidence-based supplement API. Query 74 ingredients, 30 conditions, 152 evidence links, and 2,881 PubMed papers — zero dependencies, stdlib only (`net/http` + `encoding/json`).

CITED Health indexes PubMed research and calculates evidence grades from A (strong: multiple RCTs/meta-analyses) to F (negative: most studies show no effect). The API is free, no authentication required, and returns JSON with CORS enabled.

> **Explore the evidence at [citedhealth.com](https://citedhealth.com)** — [Ingredients](https://citedhealth.com/api/ingredients/) · [Evidence](https://citedhealth.com/api/evidence/) · [Papers](https://citedhealth.com/api/papers/) · [Developer Docs](https://citedhealth.com/developers/)

## Table of Contents

- [Install](#install)
- [Quick Start](#quick-start)
- [What You Can Do](#what-you-can-do)
  - [Search Supplement Ingredients](#search-supplement-ingredients)
  - [Check Evidence Grades](#check-evidence-grades)
  - [Search PubMed Papers](#search-pubmed-papers)
- [Error Handling](#error-handling)
- [Evidence Grades](#evidence-grades)
- [API Reference](#api-reference)
- [Learn More About Evidence-Based Supplements](#learn-more-about-evidence-based-supplements)
- [Also Available](#also-available)
- [License](#license)

## Install

```bash
go get github.com/citedhealth/citedhealth-go
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	citedhealth "github.com/citedhealth/citedhealth-go"
)

func main() {
	client := citedhealth.New()
	ctx := context.Background()

	// Search ingredients
	ingredients, err := client.SearchIngredients(ctx, "biotin", "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ingredients[0].Name) // "Biotin"

	// Get evidence grade for ingredient-condition pair
	evidence, err := client.GetEvidence(ctx, "biotin", "hair-loss")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Grade: %s — %s\n", evidence.Grade, evidence.GradeLabel)
	// Grade: A — Strong Evidence

	// Search PubMed papers
	papers, err := client.SearchPapers(ctx, "biotin hair loss", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d papers found\n", len(papers))
}
```

## What You Can Do

### Search Supplement Ingredients

Find ingredients by name or filter by category. Each ingredient includes mechanism of action, recommended dosage by population, available forms, and evidence linkage.

| Category | Examples |
|----------|---------|
| vitamins | Biotin, Vitamin D, Vitamin C |
| minerals | Magnesium, Zinc, Iron |
| amino-acids | L-Theanine, Tryptophan, Glycine |
| herbs | Ashwagandha, Valerian, Melatonin |

```go
client := citedhealth.New()
ctx := context.Background()

// Search by keyword — returns matching ingredients
results, _ := client.SearchIngredients(ctx, "melatonin", "")
fmt.Println(results[0].Mechanism) // "Regulates circadian rhythm..."

// Filter by category
minerals, _ := client.SearchIngredients(ctx, "", "minerals")

// Get a specific ingredient by slug
biotin, _ := client.GetIngredient(ctx, "biotin")
fmt.Println(biotin.RecommendedDosage) // map[general:2.5-5mg deficiency:10-30mg]
```

Learn more: [Browse Ingredients](https://citedhealth.com/) · [Evidence Database](https://citedhealth.com/evidence/) · [Developer Docs](https://citedhealth.com/developers/)

### Check Evidence Grades

Every ingredient-condition pair has an evidence grade calculated from peer-reviewed PubMed studies. Grades reflect the strength, consistency, and quantity of evidence.

```go
client := citedhealth.New()
ctx := context.Background()

// Get evidence for a specific ingredient-condition pair
evidence, _ := client.GetEvidence(ctx, "biotin", "hair-loss")
fmt.Printf("Grade %s: %d studies\n", evidence.Grade, evidence.TotalStudies)
// Grade A: 12 studies

// Evidence includes direction of effect
fmt.Println(evidence.Direction) // "positive" | "negative" | "neutral" | "mixed"
fmt.Println(evidence.Summary)   // Human-readable summary

// Fetch by ID if you already know it
ev, _ := client.GetEvidenceByID(ctx, 1)
```

Learn more: [Evidence Reviews](https://citedhealth.com/evidence/) · [Grading Methodology](https://citedhealth.com/editorial-policy/) · [Hair Health](https://haircited.com) · [Sleep Health](https://sleepcited.com)

### Search PubMed Papers

All 2,881 papers are indexed from PubMed and enriched with citation data from Semantic Scholar. Filter by keyword or publication year.

```go
client := citedhealth.New()
ctx := context.Background()

// Search papers by title/abstract keyword
papers, _ := client.SearchPapers(ctx, "melatonin sleep quality", nil)
for _, p := range papers {
	// Each paper includes PMID, journal, citation count, open access status
	fmt.Printf("[PMID %s] %s (%v)\n", p.PMID, p.Title, p.PublicationYear)
	fmt.Printf("  %d citations — %s\n", p.CitationCount, p.PubMedLink)
}

// Filter by publication year
year := 2023
recent, _ := client.SearchPapers(ctx, "biotin", &year)

// Fetch a specific paper by PubMed ID
paper, _ := client.GetPaper(ctx, "12345678")
```

Learn more: [Browse Papers](https://citedhealth.com/papers/) · [OpenAPI Spec](https://citedhealth.com/api/openapi.json) · [REST API Docs](https://citedhealth.com/developers/)

## Error Handling

The client returns typed errors for common failure cases:

```go
import (
	"errors"

	citedhealth "github.com/citedhealth/citedhealth-go"
)

evidence, err := client.GetEvidence(ctx, "biotin", "nonexistent-condition")
if err != nil {
	var notFound *citedhealth.NotFoundError
	var rateLimit *citedhealth.RateLimitError
	var apiErr *citedhealth.CitedHealthError

	switch {
	case errors.As(err, &notFound):
		// 404 response or empty result for GetEvidence
		fmt.Printf("Not found: %s — %s\n", notFound.Resource, notFound.Identifier)
	case errors.As(err, &rateLimit):
		// 429 response (rate limit exceeded)
		fmt.Printf("Rate limited. Retry after %d seconds\n", rateLimit.RetryAfter)
	case errors.As(err, &apiErr):
		// Other API errors (5xx, network issues)
		fmt.Printf("API error: %s\n", apiErr.Message)
	default:
		fmt.Printf("Unexpected error: %v\n", err)
	}
}
```

## Evidence Grades

| Grade | Label | Criteria |
|-------|-------|----------|
| A | Strong Evidence | Multiple RCTs/meta-analyses, consistent positive results |
| B | Good Evidence | At least one RCT, mostly consistent |
| C | Some Evidence | Small studies, some positive signals |
| D | Very Early Research | In vitro, case reports, pilot studies |
| F | Evidence Against | <30% of studies show positive effects |

## API Reference

| Method | Description | Returns |
|--------|-------------|---------|
| `SearchIngredients(ctx, query, category)` | Search ingredients by name or category | `([]Ingredient, error)` |
| `GetIngredient(ctx, slug)` | Get ingredient by slug | `(*Ingredient, error)` |
| `GetEvidence(ctx, ingredientSlug, conditionSlug)` | Get evidence for ingredient-condition pair | `(*EvidenceLink, error)` |
| `GetEvidenceByID(ctx, id)` | Get evidence link by numeric ID | `(*EvidenceLink, error)` |
| `SearchPapers(ctx, query, year)` | Search PubMed papers | `([]Paper, error)` |
| `GetPaper(ctx, pmid)` | Get paper by PubMed ID | `(*Paper, error)` |

### Constructor Options

```go
// Default client
client := citedhealth.New()

// Custom base URL and timeout
client := citedhealth.New(
	citedhealth.WithBaseURL("https://citedhealth.com"),
	citedhealth.WithTimeout(10 * time.Second),
)
```

Full API documentation: [citedhealth.com/developers/](https://citedhealth.com/developers/)
OpenAPI 3.1.0 spec: [citedhealth.com/api/openapi.json](https://citedhealth.com/api/openapi.json)

## Learn More About Evidence-Based Supplements

- **Tools**: [Evidence Checker](https://citedhealth.com/evidence/) · [Ingredient Browser](https://citedhealth.com/) · [Paper Search](https://citedhealth.com/papers/)
- **Browse**: [Hair Health](https://haircited.com) · [Sleep Health](https://sleepcited.com) · [All Ingredients](https://citedhealth.com/api/ingredients/)
- **Guides**: [Grading Methodology](https://citedhealth.com/editorial-policy/) · [Medical Disclaimer](https://citedhealth.com/medical-disclaimer/)
- **API**: [REST API Docs](https://citedhealth.com/developers/) · [OpenAPI Spec](https://citedhealth.com/api/openapi.json)
- **Python**: [citedhealth on PyPI](https://pypi.org/project/citedhealth/)
- **TypeScript**: [citedhealth on npm](https://www.npmjs.com/package/citedhealth)

## Also Available

| Platform | Install | Link |
|----------|---------|------|
| **PyPI** | `pip install citedhealth` | [PyPI](https://pypi.org/project/citedhealth/) |
| **npm** | `npm install citedhealth` | [npm](https://www.npmjs.com/package/citedhealth) |
| **MCP** | `uvx citedhealth-mcp` | [PyPI](https://pypi.org/project/citedhealth-mcp/) |

## License

MIT — see [LICENSE](LICENSE).
