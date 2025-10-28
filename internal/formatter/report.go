package formatter

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"farmix-cli/internal/bitrix"
)

// FormatReportAsTable formats deals report as ASCII table with aligned columns
func FormatReportAsTable(deals []bitrix.DealReportRow, writer io.Writer) error {
	if len(deals) == 0 {
		fmt.Fprintf(writer, "Нет сделок для отображения\n")
		return nil
	}

	// Define column headers
	headers := []string{
		"ID",
		"Название",
		"Дата создания",
		"М/ч (₽)",
		"Ч/ч (₽)",
		"Материал (₽)",
		"Итого (₽)",
		"Оплата",
	}

	// Prepare data rows
	rows := make([][]string, len(deals))
	for i, deal := range deals {
		// Format date (take only the date part if it's a datetime)
		date := deal.DateCreate
		if len(date) > 10 {
			date = date[:10]
		}

		rows[i] = []string{
			deal.ID,
			deal.Title,
			date,
			bitrix.ParseCustomFieldValue(deal.MachineCost),
			bitrix.ParseCustomFieldValue(deal.HumanCost),
			bitrix.ParseCustomFieldValue(deal.MaterialCost),
			bitrix.ParseCustomFieldValue(deal.TotalCost),
			bitrix.ParseCustomFieldValue(deal.PaymentReceived),
		}
	}

	// Calculate column widths (considering UTF-8 characters)
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = utf8.RuneCountInString(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			cellWidth := utf8.RuneCountInString(cell)
			if cellWidth > colWidths[i] {
				colWidths[i] = cellWidth
			}
		}
	}

	// Print top border
	printBorder(writer, colWidths, "┌", "┬", "┐")

	// Print header
	printRow(writer, headers, colWidths)

	// Print header separator
	printBorder(writer, colWidths, "├", "┼", "┤")

	// Print data rows
	for _, row := range rows {
		printRow(writer, row, colWidths)
	}

	// Print bottom border
	printBorder(writer, colWidths, "└", "┴", "┘")

	// Print summary
	fmt.Fprintf(writer, "\nВсего сделок: %d\n", len(deals))

	return nil
}

// FormatReportAsCSV formats deals report as CSV
func FormatReportAsCSV(deals []bitrix.DealReportRow, writer io.Writer) error {
	csvWriter := csv.NewWriter(writer)
	defer csvWriter.Flush()

	// Write headers
	headers := []string{
		"ID",
		"Название сделки",
		"Дата создания",
		"Рассчетная стоимость м/ч",
		"Рассчетная стоимость ч/ч",
		"Рассчетная стоимость материала",
		"Итого стоимость изготовления",
		"Оплата получена",
	}
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	// Write data rows
	for _, deal := range deals {
		// Format date (take only the date part if it's a datetime)
		date := deal.DateCreate
		if len(date) > 10 {
			date = date[:10]
		}

		record := []string{
			deal.ID,
			deal.Title,
			date,
			bitrix.ParseCustomFieldValue(deal.MachineCost),
			bitrix.ParseCustomFieldValue(deal.HumanCost),
			bitrix.ParseCustomFieldValue(deal.MaterialCost),
			bitrix.ParseCustomFieldValue(deal.TotalCost),
			bitrix.ParseCustomFieldValue(deal.PaymentReceived),
		}
		if err := csvWriter.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// printBorder prints a border line for the table
func printBorder(writer io.Writer, colWidths []int, left, middle, right string) {
	fmt.Fprint(writer, left)
	for i, width := range colWidths {
		fmt.Fprint(writer, strings.Repeat("─", width+2)) // +2 for padding
		if i < len(colWidths)-1 {
			fmt.Fprint(writer, middle)
		}
	}
	fmt.Fprintln(writer, right)
}

// printRow prints a data row with proper alignment and padding
func printRow(writer io.Writer, cells []string, colWidths []int) {
	fmt.Fprint(writer, "│")
	for i, cell := range cells {
		// Calculate padding needed (considering UTF-8 characters)
		cellWidth := utf8.RuneCountInString(cell)
		padding := colWidths[i] - cellWidth

		// Right-align numbers, left-align text
		if i >= 2 && i <= 5 { // Numeric columns (М/ч, Ч/ч, Материал, Итого)
			fmt.Fprintf(writer, " %s%s ", strings.Repeat(" ", padding), cell)
		} else {
			fmt.Fprintf(writer, " %s%s ", cell, strings.Repeat(" ", padding))
		}
		fmt.Fprint(writer, "│")
	}
	fmt.Fprintln(writer)
}
