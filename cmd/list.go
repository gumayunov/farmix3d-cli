package cmd

import (
	"fmt"
	"os"
	"strings"

	"3mfanalyzer/internal/formatter"
	"3mfanalyzer/internal/parser"

	"github.com/spf13/cobra"
)

var (
	outputFormat string
)

var listCmd = &cobra.Command{
	Use:   "list [file]",
	Short: "List information about a 3MF file",
	Long:  `Display detailed information about objects and plates in a 3MF file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		if !strings.HasSuffix(strings.ToLower(filePath), ".3mf") {
			fmt.Fprintf(os.Stderr, "Error: File must have .3mf extension: %s\n", filePath)
			os.Exit(1)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File does not exist: %s\n", filePath)
			os.Exit(1)
		}

		data, err := parser.Parse3MF(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to parse 3MF file: %v\n", err)
			os.Exit(1)
		}

		switch strings.ToLower(outputFormat) {
		case "csv":
			if err := formatter.FormatAsCSV(data, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to format output as CSV: %v\n", err)
				os.Exit(1)
			}
		case "text", "":
			if err := formatter.FormatAsText(data, os.Stdout); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to format output as text: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Error: Unsupported output format: %s. Supported formats: text, csv\n", outputFormat)
			os.Exit(1)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&outputFormat, "format", "f", "text", "Output format (text, csv)")
	rootCmd.AddCommand(listCmd)
}