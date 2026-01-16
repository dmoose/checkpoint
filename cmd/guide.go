package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dmoose/checkpoint/internal/guides"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(guideCmd)
}

var guideCmd = &cobra.Command{
	Use:   "guide [topic]",
	Short: "Show detailed guides and documentation",
	Long: `Displays built-in guide documents.
Topics: first-time-user, llm-workflow, best-practices`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		topic := ""
		if len(args) > 0 {
			topic = args[0]
		}
		Guide(topic)
	},
}

// Guide displays guide documents from embedded content
func Guide(topic string) {
	// If no topic specified, list available guides
	if topic == "" {
		listGuides()
		return
	}

	// Show specific guide
	showGuide(topic)
}

// listGuides shows available guide topics
func listGuides() {
	fmt.Println("\nCHECKPOINT GUIDES")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println("\nAvailable guides:")
	fmt.Println()

	for name, guide := range guides.EmbeddedGuides {
		fmt.Printf("  %-20s %s\n", name, guide.Description)
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  checkpoint guide [topic]")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  checkpoint guide first-time-user")
	fmt.Println("  checkpoint guide llm-workflow")
	fmt.Println("  checkpoint guide best-practices")
	fmt.Println()
}

// showGuide displays a specific guide
func showGuide(topic string) {
	guide, ok := guides.GetGuide(topic)
	if !ok {
		fmt.Fprintf(os.Stderr, "Guide '%s' not found\n", topic)
		fmt.Fprintf(os.Stderr, "Run 'checkpoint guide' to see available guides\n")
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("GUIDE: %s\n", strings.ToUpper(strings.ReplaceAll(topic, "-", " ")))
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()
	fmt.Println(guide.Content)
}
