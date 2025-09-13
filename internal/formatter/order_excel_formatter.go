package formatter

import (
	"fmt"
	"strconv"
	"time"

	"3mfanalyzer/internal/bitrix"
	"3mfanalyzer/internal/parser"

	"github.com/xuri/excelize/v2"
)

// FormatAsOrderExcel creates the main order report Excel file
func FormatAsOrderExcel(data *parser.Parser3MF, deal *bitrix.Deal, user *bitrix.User, customerName string, client *bitrix.Client, outputPath string) error {
	// Create new Excel file
	f := excelize.NewFile()
	colors := DefaultExcelColors()
	
	// Delete default sheet
	f.DeleteSheet("Sheet1")
	
	// Create the order sheet
	sheetName := "Наряд-заказ"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create order sheet: %w", err)
	}
	
	// Set active sheet
	f.SetActiveSheet(0)
	
	// Create order content
	if err := createOrderContent(f, sheetName, data, deal, user, customerName, client, colors); err != nil {
		return fmt.Errorf("failed to create order content: %w", err)
	}
	
	// Save file
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save order Excel file: %w", err)
	}
	
	return nil
}

// FormatAsAssignmentExcel creates the assignment report Excel file
func FormatAsAssignmentExcel(data *parser.Parser3MF, deal *bitrix.Deal, user *bitrix.User, customerName string, client *bitrix.Client, outputPath string) error {
	// Create new Excel file
	f := excelize.NewFile()
	colors := DefaultExcelColors()
	
	// Delete default sheet
	f.DeleteSheet("Sheet1")
	
	// Create the assignment sheet
	sheetName := "Сменное задание"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create assignment sheet: %w", err)
	}
	
	// Set active sheet
	f.SetActiveSheet(0)
	
	// Create assignment content
	if err := createAssignmentContent(f, sheetName, data, deal, user, customerName, colors); err != nil {
		return fmt.Errorf("failed to create assignment content: %w", err)
	}
	
	// Save file
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save assignment Excel file: %w", err)
	}
	
	return nil
}

// createOrderContent creates the detailed order report content
func createOrderContent(f *excelize.File, sheetName string, data *parser.Parser3MF, deal *bitrix.Deal, user *bitrix.User, customerName string, client *bitrix.Client, colors ExcelColors) error {
	row := 1
	
	// Title
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "НАРЯД-ЗАКАЗ")
	
	// Title style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 18,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "H"+strconv.Itoa(row), titleStyle)
	f.MergeCell(sheetName, "A"+strconv.Itoa(row), "H"+strconv.Itoa(row))
	
	row += 2
	
	// Deal information block
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Ответственный:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), user.FullName)
	row++
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Заказчик:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), customerName)
	row++
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Сделка:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), deal.ID)
	row++
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Ссылка:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), client.GetDealURL(deal.ID))
	row++
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Дата:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), time.Now().Format("02.01.2006"))
	row += 2
	
	// Process each plate
	for _, plate := range data.Plates {
		row = createPlateSection(f, sheetName, plate, row, colors)
		row += 2 // Add space between plates
	}
	
	// Materials summary
	row = createMaterialsSection(f, sheetName, data, row, colors)
	row += 2
	
	// Hours section
	row = createHoursSection(f, sheetName, row, colors)
	
	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 20)
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 15)
	f.SetColWidth(sheetName, "G", "G", 15)
	f.SetColWidth(sheetName, "H", "H", 15)
	
	return nil
}

// createAssignmentContent creates the assignment report content (simplified version)
func createAssignmentContent(f *excelize.File, sheetName string, data *parser.Parser3MF, deal *bitrix.Deal, user *bitrix.User, customerName string, colors ExcelColors) error {
	row := 1
	
	// Title
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "СМЕННОЕ ЗАДАНИЕ")
	
	// Title style
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 18,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "F"+strconv.Itoa(row), titleStyle)
	f.MergeCell(sheetName, "A"+strconv.Itoa(row), "F"+strconv.Itoa(row))
	
	row += 2
	
	// Basic info
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Заказчик:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), customerName)
	row++
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Сделка:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), deal.ID+" - "+deal.Title)
	row++
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Дата:")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), time.Now().Format("02.01.2006"))
	row += 2
	
	// Simplified plate information
	for _, plate := range data.Plates {
		groups := parser.GroupObjectsByName(plate.Objects)
		if len(groups) == 0 {
			continue
		}
		
		// Get first material from groups
		var firstMaterial string
		for _, group := range groups {
			firstMaterial = cleanMaterialName(group.Material)
			break
		}
		
		// Plate header
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Стол")
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), plate.PlateID)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "Материал")
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), firstMaterial)
		row++
		
		// Objects table header
		headerStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Bold: true,
			},
			Fill: excelize.Fill{
				Type:    "pattern",
				Color:   []string{colors.HeaderBg},
				Pattern: 1,
			},
		})
		
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Название детали")
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), "Количество")
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "B"+strconv.Itoa(row), headerStyle)
		row++
		
		// Objects data
		for _, group := range groups {
			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), group.Name)
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), group.Count)
			row++
		}
		
		row++ // Space between plates
	}
	
	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 40)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 15)
	f.SetColWidth(sheetName, "D", "D", 25)
	f.SetColWidth(sheetName, "E", "E", 15)
	f.SetColWidth(sheetName, "F", "F", 15)
	
	return nil
}

// createPlateSection creates a section for one plate in the order report
func createPlateSection(f *excelize.File, sheetName string, plate parser.PlateInfo, startRow int, colors ExcelColors) int {
	row := startRow
	groups := parser.GroupObjectsByName(plate.Objects)
	
	if len(groups) == 0 {
		return row
	}
	
	// Get material from first group
	var firstMaterial string
	for _, group := range groups {
		firstMaterial = cleanMaterialName(group.Material)
		break
	}
	
	// Plate header row
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Стол")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), plate.PlateID)
	f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "Повторений")
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), 1)
	f.SetCellValue(sheetName, "E"+strconv.Itoa(row), "Материал")
	f.SetCellValue(sheetName, "F"+strconv.Itoa(row), firstMaterial)
	row++
	
	// Weight and time row (empty values as requested) - убрали "Вес модели"
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Общий вес, г")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), "")
	f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "Вес поддержек, г")
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "")
	f.SetCellValue(sheetName, "E"+strconv.Itoa(row), "Время печати, ч")
	f.SetCellValue(sheetName, "F"+strconv.Itoa(row), "")
	row++
	
	// Parts table header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: colors.HeaderText,
			Bold:  true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{colors.HeaderBg},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Название детали")
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "Количество на столе")
	f.SetCellValue(sheetName, "E"+strconv.Itoa(row), "Примерный вес")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "E"+strconv.Itoa(row), headerStyle)
	// Объединяем ячейки A, B, C для названия детали
	f.MergeCell(sheetName, "A"+strconv.Itoa(row), "C"+strconv.Itoa(row))
	row++
	
	// Parts data
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	for _, group := range groups {
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), group.Name)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), group.Count)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), "") // Empty as requested
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "E"+strconv.Itoa(row), dataStyle)
		// Объединяем ячейки A, B, C для названия детали
		f.MergeCell(sheetName, "A"+strconv.Itoa(row), "C"+strconv.Itoa(row))
		row++
	}
	
	return row
}

// createMaterialsSection creates the materials summary section
func createMaterialsSection(f *excelize.File, sheetName string, data *parser.Parser3MF, startRow int, colors ExcelColors) int {
	row := startRow
	
	// Collect unique materials
	materialsSet := make(map[string]bool)
	for _, plate := range data.Plates {
		for _, obj := range plate.Objects {
			if obj.Material != "" {
				cleanMaterial := cleanMaterialName(obj.Material)
				materialsSet[cleanMaterial] = true
			}
		}
	}
	
	if len(materialsSet) == 0 {
		return row
	}
	
	// Materials table header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: colors.HeaderText,
			Bold:  true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{colors.HeaderBg},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Название")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), "Вес")
	f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "Стоимость за кг")
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "Стоимость")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "D"+strconv.Itoa(row), headerStyle)
	row++
	
	// Materials data
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	for material := range materialsSet {
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), material)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), "")
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "")
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "")
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "D"+strconv.Itoa(row), dataStyle)
		row++
	}
	
	return row
}

// createHoursSection creates the hours summary section
func createHoursSection(f *excelize.File, sheetName string, startRow int, colors ExcelColors) int {
	row := startRow
	
	// Hours table header
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color: colors.HeaderText,
			Bold:  true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{colors.HeaderBg},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Тип работ")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), "Часы")
	f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "Ставка")
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "Стоимость")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "D"+strconv.Itoa(row), headerStyle)
	row++
	
	// Hours data
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	// Machine hours row
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Машино-часы")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), "")
	f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "")
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "D"+strconv.Itoa(row), dataStyle)
	row++
	
	// Operator hours row
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Работа оператора")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), "")
	f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "")
	f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "D"+strconv.Itoa(row), dataStyle)
	row++
	
	return row
}