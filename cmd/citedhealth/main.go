package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	citedhealth "github.com/citedhealth/citedhealth-go"
	"github.com/spf13/cobra"
)

const version = "0.4.0"

var jsonCompact bool

func main() {
	rootCmd := &cobra.Command{
		Use:     "citedhealth",
		Short:   "CLI for the CITED Health evidence-based supplement API",
		Version: version,
	}

	rootCmd.PersistentFlags().BoolVar(&jsonCompact, "json", false, "Output compact JSON instead of pretty-printed")

	rootCmd.AddCommand(
		ingredientsCmd(),
		ingredientCmd(),
		evidenceCmd(),
		papersCmd(),
		paperCmd(),
		conditionsCmd(),
		conditionCmd(),
		glossaryCmd(),
		glossaryTermCmd(),
		guidesCmd(),
		guideCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func printJSON(v any) {
	var data []byte
	var err error
	if jsonCompact {
		data, err = json.Marshal(v)
	} else {
		data, err = json.MarshalIndent(v, "", "  ")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

func exitError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func ingredientsCmd() *cobra.Command {
	var category string
	cmd := &cobra.Command{
		Use:   "ingredients [query]",
		Short: "Search supplement ingredients",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			client := citedhealth.New()
			results, err := client.SearchIngredients(context.Background(), query, category)
			if err != nil {
				exitError(err)
			}
			printJSON(results)
		},
	}
	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category (e.g. vitamins, minerals, herbs)")
	return cmd
}

func ingredientCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ingredient <slug>",
		Short: "Get a single ingredient by slug",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := citedhealth.New()
			result, err := client.GetIngredient(context.Background(), args[0])
			if err != nil {
				exitError(err)
			}
			printJSON(result)
		},
	}
}

func evidenceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "evidence <ingredient> <condition>",
		Short: "Get evidence for an ingredient-condition pair",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			client := citedhealth.New()
			result, err := client.GetEvidence(context.Background(), args[0], args[1])
			if err != nil {
				exitError(err)
			}
			printJSON(result)
		},
	}
}

func papersCmd() *cobra.Command {
	var year int
	cmd := &cobra.Command{
		Use:   "papers [query]",
		Short: "Search PubMed papers",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			var yearPtr *int
			if cmd.Flags().Changed("year") {
				yearPtr = &year
			}
			client := citedhealth.New()
			results, err := client.SearchPapers(context.Background(), query, yearPtr)
			if err != nil {
				exitError(err)
			}
			printJSON(results)
		},
	}
	cmd.Flags().IntVarP(&year, "year", "y", 0, "Filter by publication year")
	return cmd
}

func paperCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "paper <pmid>",
		Short: "Get a paper by PubMed ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Validate PMID is numeric
			if _, err := strconv.Atoi(args[0]); err != nil {
				exitError(fmt.Errorf("invalid PMID %q: must be a number", args[0]))
			}
			client := citedhealth.New()
			result, err := client.GetPaper(context.Background(), args[0])
			if err != nil {
				exitError(err)
			}
			printJSON(result)
		},
	}
}

func conditionsCmd() *cobra.Command {
	var featured bool
	cmd := &cobra.Command{
		Use:   "conditions",
		Short: "List health conditions",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var featuredPtr *bool
			if cmd.Flags().Changed("featured") {
				featuredPtr = &featured
			}
			client := citedhealth.New()
			results, err := client.ListConditions(context.Background(), featuredPtr)
			if err != nil {
				exitError(err)
			}
			printJSON(results)
		},
	}
	cmd.Flags().BoolVar(&featured, "featured", false, "Filter by featured status")
	return cmd
}

func conditionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "condition <slug>",
		Short: "Get a single condition by slug",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := citedhealth.New()
			result, err := client.GetCondition(context.Background(), args[0])
			if err != nil {
				exitError(err)
			}
			printJSON(result)
		},
	}
}

func glossaryCmd() *cobra.Command {
	var category string
	cmd := &cobra.Command{
		Use:   "glossary",
		Short: "List glossary terms",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client := citedhealth.New()
			results, err := client.ListGlossary(context.Background(), category)
			if err != nil {
				exitError(err)
			}
			printJSON(results)
		},
	}
	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category")
	return cmd
}

func glossaryTermCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "glossary-term <slug>",
		Short: "Get a single glossary term by slug",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := citedhealth.New()
			result, err := client.GetGlossaryTerm(context.Background(), args[0])
			if err != nil {
				exitError(err)
			}
			printJSON(result)
		},
	}
}

func guidesCmd() *cobra.Command {
	var category string
	cmd := &cobra.Command{
		Use:   "guides",
		Short: "List educational guides",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			client := citedhealth.New()
			results, err := client.ListGuides(context.Background(), category)
			if err != nil {
				exitError(err)
			}
			printJSON(results)
		},
	}
	cmd.Flags().StringVarP(&category, "category", "c", "", "Filter by category")
	return cmd
}

func guideCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "guide <slug>",
		Short: "Get a single guide by slug",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			client := citedhealth.New()
			result, err := client.GetGuide(context.Background(), args[0])
			if err != nil {
				exitError(err)
			}
			printJSON(result)
		},
	}
}
