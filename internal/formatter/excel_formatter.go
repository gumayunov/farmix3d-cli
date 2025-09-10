package formatter

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"3mfanalyzer/internal/parser"

	"github.com/xuri/excelize/v2"
)

// ExcelColors определяет цветовую схему для Excel таблиц
type ExcelColors struct {
	HeaderBg     string // Фон заголовков
	HeaderText   string // Текст заголовков
	AltRowBg     string // Фон четных строк
	BorderColor  string // Цвет границ
	SummaryBg    string // Фон итоговых ячеек
}

// DefaultExcelColors возвращает стандартную цветовую схему
func DefaultExcelColors() ExcelColors {
	return ExcelColors{
		HeaderBg:    "#4472C4",  // Синий
		HeaderText:  "#FFFFFF",  // Белый
		AltRowBg:    "#F2F2F2",  // Светло-серый
		BorderColor: "#D9D9D9",  // Серый
		SummaryBg:   "#E7E6E6",  // Светло-серый для итогов
	}
}

// FormatAsExcel создает Excel отчет из данных 3MF файла
func FormatAsExcel(data *parser.Parser3MF, outputPath string) error {
	// Создаем новый Excel файл
	f := excelize.NewFile()
	colors := DefaultExcelColors()
	
	// Удаляем стандартный лист
	f.DeleteSheet("Sheet1")
	
	// Создаем листы
	if err := createSummarySheet(f, data, colors, outputPath); err != nil {
		return fmt.Errorf("failed to create summary sheet: %w", err)
	}
	
	if err := createPlatesSheet(f, data, colors); err != nil {
		return fmt.Errorf("failed to create plates sheet: %w", err)
	}
	
	if err := createObjectsSheet(f, data, colors); err != nil {
		return fmt.Errorf("failed to create objects sheet: %w", err)
	}
	
	// Устанавливаем активный лист
	summaryIndex, _ := f.GetSheetIndex("Summary")
	f.SetActiveSheet(summaryIndex)
	
	// Сохраняем файл
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save Excel file: %w", err)
	}
	
	return nil
}

// createSummarySheet создает лист с общей информацией
func createSummarySheet(f *excelize.File, data *parser.Parser3MF, colors ExcelColors, outputPath string) error {
	sheetName := "Summary"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	
	// Заголовок документа
	filename := filepath.Base(strings.TrimSuffix(outputPath, "_order.xlsx"))
	f.SetCellValue(sheetName, "A1", "3MF Order Report")
	f.SetCellValue(sheetName, "A2", filename)
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	
	// Стиль заголовка
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
	})
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	
	// Статистика по столам
	row := 5
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Plates Summary:")
	
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
	
	row++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Plate ID")
	f.SetCellValue(sheetName, "B"+strconv.Itoa(row), "Plate Name")
	f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "Objects Count")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "C"+strconv.Itoa(row), headerStyle)
	
	// Стиль данных
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	for _, plate := range data.Plates {
		row++
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), plate.PlateID)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), plate.PlateName)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), len(plate.Objects))
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "C"+strconv.Itoa(row), dataStyle)
	}
	
	// Материалы
	row += 3
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Materials Used:")
	
	materialsSet := make(map[string]bool)
	for _, plate := range data.Plates {
		for _, obj := range plate.Objects {
			if obj.Material != "" {
				cleanMaterial := cleanMaterialName(obj.Material)
				materialsSet[cleanMaterial] = true
			}
		}
	}
	
	var materials []string
	for material := range materialsSet {
		materials = append(materials, material)
	}
	sort.Strings(materials)
	
	row++
	f.SetCellValue(sheetName, "A"+strconv.Itoa(row), "Material")
	f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "A"+strconv.Itoa(row), headerStyle)
	
	for _, material := range materials {
		row++
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), material)
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "A"+strconv.Itoa(row), dataStyle)
	}
	
	// Настройка ширины колонок
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 30)
	f.SetColWidth(sheetName, "C", "C", 15)
	
	_ = index
	return nil
}

// createPlatesSheet создает лист с информацией по столам
func createPlatesSheet(f *excelize.File, data *parser.Parser3MF, colors ExcelColors) error {
	sheetName := "Plates"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	
	// Заголовки
	headers := []string{"Plate ID", "Plate Name", "Object Name", "Object Type", "Material", "Count"}
	for i, header := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"1", header)
	}
	
	// Стиль заголовка
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
	
	f.SetCellStyle(sheetName, "A1", "F1", headerStyle)
	
	// Стили для данных
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	altRowStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{colors.AltRowBg},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	row := 2
	for _, plate := range data.Plates {
		groups := parser.GroupObjectsByName(plate.Objects)
		for _, group := range groups {
			cleanMaterial := cleanMaterialName(group.Material)
			objectType := group.Type
			if objectType == "assembly" {
				objectType = "assembly"
			}
			
			// Основной объект
			style := dataStyle
			if row%2 == 0 {
				style = altRowStyle
			}
			
			f.SetCellValue(sheetName, "A"+strconv.Itoa(row), plate.PlateID)
			f.SetCellValue(sheetName, "B"+strconv.Itoa(row), plate.PlateName)
			f.SetCellValue(sheetName, "C"+strconv.Itoa(row), group.Name)
			f.SetCellValue(sheetName, "D"+strconv.Itoa(row), objectType)
			f.SetCellValue(sheetName, "E"+strconv.Itoa(row), cleanMaterial)
			f.SetCellValue(sheetName, "F"+strconv.Itoa(row), group.Count)
			f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "F"+strconv.Itoa(row), style)
			row++
			
			// Компоненты сборки
			if group.Type == "assembly" && len(group.Components) > 0 {
				for _, comp := range group.Components {
					style := dataStyle
					if row%2 == 0 {
						style = altRowStyle
					}
					
					f.SetCellValue(sheetName, "A"+strconv.Itoa(row), plate.PlateID)
					f.SetCellValue(sheetName, "B"+strconv.Itoa(row), plate.PlateName)
					f.SetCellValue(sheetName, "C"+strconv.Itoa(row), "  ├─ "+comp.Name)
					f.SetCellValue(sheetName, "D"+strconv.Itoa(row), "component")
					f.SetCellValue(sheetName, "E"+strconv.Itoa(row), "")
					f.SetCellValue(sheetName, "F"+strconv.Itoa(row), "")
					f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "F"+strconv.Itoa(row), style)
					row++
				}
			}
		}
	}
	
	// Настройка ширины колонок
	f.SetColWidth(sheetName, "A", "A", 10)
	f.SetColWidth(sheetName, "B", "B", 20)
	f.SetColWidth(sheetName, "C", "C", 40)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 20)
	f.SetColWidth(sheetName, "F", "F", 10)
	
	return nil
}

// createObjectsSheet создает лист с полной информацией по объектам
func createObjectsSheet(f *excelize.File, data *parser.Parser3MF, colors ExcelColors) error {
	sheetName := "Objects"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	
	// Заголовки
	headers := []string{"Object Name", "Type", "Material", "Total Count", "Plates"}
	for i, header := range headers {
		col := string(rune('A' + i))
		f.SetCellValue(sheetName, col+"1", header)
	}
	
	// Стиль заголовка
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
	
	f.SetCellStyle(sheetName, "A1", "E1", headerStyle)
	
	// Собираем статистику по объектам
	objectStats := make(map[string]*ObjectStat)
	
	for _, plate := range data.Plates {
		groups := parser.GroupObjectsByName(plate.Objects)
		for _, group := range groups {
			cleanMaterial := cleanMaterialName(group.Material)
			key := group.Name + "|" + cleanMaterial
			
			if stat, exists := objectStats[key]; exists {
				stat.Count += group.Count
				stat.Plates = append(stat.Plates, plate.PlateID)
			} else {
				objectStats[key] = &ObjectStat{
					Name:     group.Name,
					Type:     group.Type,
					Material: cleanMaterial,
					Count:    group.Count,
					Plates:   []int{plate.PlateID},
				}
			}
		}
	}
	
	// Стили для данных
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	altRowStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{colors.AltRowBg},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "left", Color: colors.BorderColor, Style: 1},
			{Type: "top", Color: colors.BorderColor, Style: 1},
			{Type: "bottom", Color: colors.BorderColor, Style: 1},
			{Type: "right", Color: colors.BorderColor, Style: 1},
		},
	})
	
	// Сортируем объекты по имени
	var sortedObjects []*ObjectStat
	for _, stat := range objectStats {
		sortedObjects = append(sortedObjects, stat)
	}
	sort.Slice(sortedObjects, func(i, j int) bool {
		return sortedObjects[i].Name < sortedObjects[j].Name
	})
	
	row := 2
	for _, stat := range sortedObjects {
		style := dataStyle
		if row%2 == 0 {
			style = altRowStyle
		}
		
		// Формируем список столов
		var plateStrs []string
		for _, plateID := range stat.Plates {
			plateStrs = append(plateStrs, strconv.Itoa(plateID))
		}
		platesText := strings.Join(plateStrs, ", ")
		
		f.SetCellValue(sheetName, "A"+strconv.Itoa(row), stat.Name)
		f.SetCellValue(sheetName, "B"+strconv.Itoa(row), stat.Type)
		f.SetCellValue(sheetName, "C"+strconv.Itoa(row), stat.Material)
		f.SetCellValue(sheetName, "D"+strconv.Itoa(row), stat.Count)
		f.SetCellValue(sheetName, "E"+strconv.Itoa(row), platesText)
		f.SetCellStyle(sheetName, "A"+strconv.Itoa(row), "E"+strconv.Itoa(row), style)
		row++
	}
	
	// Настройка ширины колонок
	f.SetColWidth(sheetName, "A", "A", 40)
	f.SetColWidth(sheetName, "B", "B", 15)
	f.SetColWidth(sheetName, "C", "C", 20)
	f.SetColWidth(sheetName, "D", "D", 12)
	f.SetColWidth(sheetName, "E", "E", 15)
	
	return nil
}

// ObjectStat содержит статистику по объекту
type ObjectStat struct {
	Name     string
	Type     string
	Material string
	Count    int
	Plates   []int
}