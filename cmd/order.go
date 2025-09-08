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

var orderCmd = &cobra.Command{
	Use:   "order [file]",
	Short: "Generate PDF order report for a 3MF file",
	Long: `Generate a detailed PDF order report with object information, materials, and layouts.
	
This command creates a professional PDF report suitable for order processing and documentation.
The report includes object counts, material information, and plate layouts in a structured format.`,
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

		// Генерируем имя выходного файла для заказа
		baseName := strings.TrimSuffix(filepath.Base(filePath), ".3mf")
		outputPath := baseName + "_order.pdf"

		// Создаем PDF отчет
		if err := formatter.FormatAsPDF(data, outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to create PDF order report: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("PDF order report created: %s\n", outputPath)
	},
}

func init() {
	rootCmd.AddCommand(orderCmd)
}