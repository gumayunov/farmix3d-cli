package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"3mfanalyzer/internal/formatter"
	"3mfanalyzer/internal/parser"

	"github.com/spf13/cobra"
)

var (
	orderFormat string
)

var orderCmd = &cobra.Command{
	Use:   "order [file]",
	Short: "Generate order report for a 3MF file in PDF or Excel format",
	Long: `Generate a detailed order report with object information, materials, and layouts.
	
This command creates professional reports suitable for order processing and documentation.
The report includes object counts, material information, and plate layouts in a structured format.

Supported formats:
  pdf   - Professional PDF report with tables and formatting
  xlsx  - Microsoft Excel spreadsheet with multiple sheets`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]

		// Проверяем расширение файла
		if !strings.HasSuffix(strings.ToLower(filePath), ".3mf") {
			fmt.Fprintf(os.Stderr, "Error: File must have .3mf extension: %s\n", filePath)
			os.Exit(1)
		}

		// Проверяем существование файла
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: File does not exist: %s\n", filePath)
			os.Exit(1)
		}

		// Парсим 3MF файл
		data, err := parser.Parse3MF(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to parse 3MF file: %v\n", err)
			os.Exit(1)
		}

		// Определяем формат и создаем отчет
		baseName := strings.TrimSuffix(filepath.Base(filePath), ".3mf")
		
		switch strings.ToLower(orderFormat) {
		case "pdf", "":
			outputPath := baseName + "_order.pdf"
			if err := formatter.FormatAsPDF(data, outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to create PDF order report: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("PDF order report created: %s\n", outputPath)
			
		case "xlsx", "excel":
			outputPath := baseName + "_order.xlsx"
			if err := formatter.FormatAsExcel(data, outputPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to create Excel order report: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Excel order report created: %s\n", outputPath)
			
		default:
			fmt.Fprintf(os.Stderr, "Error: Unsupported format: %s. Supported formats: pdf, xlsx\n", orderFormat)
			os.Exit(1)
		}
	},
}

func init() {
	orderCmd.Flags().StringVarP(&orderFormat, "format", "f", "pdf", "Output format (pdf, xlsx)")
	rootCmd.AddCommand(orderCmd)
}