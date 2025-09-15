package formatter

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"farmix-cli/internal/parser"

	"github.com/go-pdf/fpdf"
)

// PDFFormatter представляет генератор PDF отчетов
type PDFFormatter struct {
	pdf      *fpdf.Fpdf
	template PDFTemplate
	widths   TableColumnWidths
}

// NewPDFFormatter создает новый генератор PDF с настройками по умолчанию
func NewPDFFormatter() *PDFFormatter {
	template := DefaultPDFTemplate()
	pdf := fpdf.New(template.Orientation, "mm", template.PageSize, "")
	
	return &PDFFormatter{
		pdf:      pdf,
		template: template,
		widths:   DefaultTableWidths(),
	}
}

// FormatAsPDF создает PDF отчет из данных 3MF файла
func FormatAsPDF(data *parser.Parser3MF, outputPath string) error {
	formatter := NewPDFFormatter()
	return formatter.Generate(data, outputPath)
}

// Generate создает PDF документ и сохраняет его в файл
func (f *PDFFormatter) Generate(data *parser.Parser3MF, outputPath string) error {
	// Настройка PDF
	f.setupPDF()
	
	// Добавляем первую страницу
	f.pdf.AddPage()
	
	// Заголовок документа
	filename := filepath.Base(strings.TrimSuffix(outputPath, "_analysis.pdf"))
	f.addDocumentHeader(filename)
	
	// Анализ по печатным столам
	if len(data.Plates) == 0 {
		f.addText("No plates found in the file.", f.template.FontSize)
	} else {
		for i, plate := range data.Plates {
			if i > 0 {
				f.addVerticalSpace(f.template.SectionSpacing)
			}
			f.addPlateSection(plate)
		}
	}
	
	// Сводка материалов
	f.addMaterialsSummary(data)
	
	// Сохранение файла
	return f.pdf.OutputFileAndClose(outputPath)
}

// setupPDF настраивает основные параметры PDF документа
func (f *PDFFormatter) setupPDF() {
	// Добавляем Unicode шрифты DejaVu Sans
	f.pdf.AddUTF8Font("DejaVuSans", "", "assets/fonts/DejaVuSans.ttf")
	f.pdf.AddUTF8Font("DejaVuSans", "B", "assets/fonts/DejaVuSans-Bold.ttf")
	
	// Установка автоматических разрывов страниц
	f.pdf.SetAutoPageBreak(true, f.template.MarginY)
	
	// Установка шрифта по умолчанию
	f.pdf.SetFont(f.template.FontFamily, "", f.template.FontSize)
	
	// Установка цвета текста по умолчанию
	f.setTextColor(f.template.Colors.Text)
}

// addDocumentHeader добавляет заголовок документа
func (f *PDFFormatter) addDocumentHeader(filename string) {
	// Заголовок
	f.setTextColor(f.template.Colors.Title)
	f.pdf.SetFont(f.template.FontFamily, "B", f.template.TitleFontSize)
	f.pdf.CellFormat(0, f.template.HeaderHeight*0.6, "3MF File Analysis", "", 1, "C", false, 0, "")
	
	// Имя файла
	f.pdf.SetFont(f.template.FontFamily, "", f.template.HeaderFontSize)
	f.pdf.CellFormat(0, f.template.HeaderHeight*0.4, filename, "", 1, "C", false, 0, "")
	
	// Дата создания отчета
	f.setTextColor(f.template.Colors.Text)
	f.pdf.SetFont(f.template.FontFamily, "", f.template.FontSize)
	dateStr := fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05"))
	f.pdf.CellFormat(0, f.template.TableRowHeight, dateStr, "", 1, "C", false, 0, "")
	
	f.addVerticalSpace(f.template.SectionSpacing)
}

// addPlateSection добавляет секцию с информацией о печатном столе
func (f *PDFFormatter) addPlateSection(plate parser.PlateInfo) {
	// Заголовок секции
	f.addSectionHeader(fmt.Sprintf("Plate %d: %s", plate.PlateID, plate.PlateName))
	
	if len(plate.Objects) == 0 {
		f.addText("No objects on this plate.", f.template.FontSize)
		return
	}
	
	// Подготовка данных таблицы
	groups := parser.GroupObjectsByName(plate.Objects)
	headers := []string{"Object Name", "Count", "Type", "Material"}
	var rows [][]string
	
	for _, group := range groups {
		cleanMaterial := cleanMaterialName(group.Material)
		objectType := group.Type
		if objectType == "assembly" {
			objectType = "assembly"
		}
		
		row := []string{
			group.Name,
			strconv.Itoa(group.Count),
			objectType,
			cleanMaterial,
		}
		rows = append(rows, row)
		
		// Добавляем компоненты сборки как вложенные строки
		if group.Type == "assembly" && len(group.Components) > 0 {
			for _, comp := range group.Components {
				componentRow := []string{
					"  ├─ " + comp.Name,
					"",
					"component",
					"",
				}
				rows = append(rows, componentRow)
			}
		}
	}
	
	// Отрисовка таблицы
	f.addTable(headers, rows)
}

// addTable создает и отрисовывает таблицу
func (f *PDFFormatter) addTable(headers []string, rows [][]string) {
	widths := []float64{f.widths.ObjectName, f.widths.Count, f.widths.Type, f.widths.Material}
	
	// Заголовок таблицы
	f.setFillColor(f.template.Colors.TableHead)
	f.setTextColor(f.template.Colors.Text)
	f.pdf.SetFont(f.template.FontFamily, "B", f.template.FontSize)
	
	for i, header := range headers {
		f.pdf.CellFormat(widths[i], f.template.TableRowHeight, header, "1", 0, "C", true, 0, "")
	}
	f.pdf.Ln(-1)
	
	// Строки таблицы с автоматическим переносом текста
	f.pdf.SetFont(f.template.FontFamily, "", f.template.FontSize)
	for i, row := range rows {
		f.addTableRowWithWrapping(row, widths, i)
	}
	
	f.addVerticalSpace(f.template.SectionSpacing)
}

// addTableRowWithWrapping добавляет строку таблицы с переносом длинного текста
func (f *PDFFormatter) addTableRowWithWrapping(row []string, widths []float64, rowIndex int) {
	// Чередующиеся цвета строк
	if rowIndex%2 == 0 {
		f.setFillColor(f.template.Colors.TableRow1)
	} else {
		f.setFillColor(f.template.Colors.TableRow2)
	}
	
	// Проверяем, нужен ли перенос страницы
	_, pageH := f.pdf.GetPageSize()
	_, _, _, bottomMargin := f.pdf.GetMargins()
	_, y := f.pdf.GetXY()
	
	// Вычисляем необходимую высоту строки для первой колонки (название объекта)
	rowHeight := f.calculateRowHeight(row[0], widths[0])
	
	if y+rowHeight > pageH-bottomMargin {
		f.pdf.AddPage()
		// Повторяем заголовок таблицы на новой странице
		f.addTableHeader([]string{"Object Name", "Count", "Type", "Material"}, widths)
		
		if rowIndex%2 == 0 {
			f.setFillColor(f.template.Colors.TableRow1)
		} else {
			f.setFillColor(f.template.Colors.TableRow2)
		}
	}
	
	// Сохраняем начальную позицию
	startX, startY := f.pdf.GetXY()
	
	// Используем единый подход для всех ячеек в строке
	f.addUniformTableRow(row, widths, startX, startY, rowHeight)
	
	// Переходим на следующую строку
	f.pdf.SetXY(startX, startY+rowHeight)
}

// addUniformTableRow рисует все ячейки в строке с одинаковой высотой
func (f *PDFFormatter) addUniformTableRow(row []string, widths []float64, startX, startY, height float64) {
	currentX := startX
	
	for j, cell := range row {
		f.pdf.SetXY(currentX, startY)
		
		// Подготавливаем текст
		text := cell
		if j == 0 && len(cell) > 50 {
			text = f.wrapText(cell)
		}
		
		// Определяем выравнивание
		alignment := "L"
		if j == 1 { // Колонка Count выравнивается по центру
			alignment = "C"
		}
		
		// Все ячейки рисуем через MultiCell с фиксированной высотой
		f.addFixedHeightMultiCell(widths[j], height, text, "1", alignment, true)
		
		currentX += widths[j]
	}
}

// addFixedHeightMultiCell рисует MultiCell с точно заданной высотой
func (f *PDFFormatter) addFixedHeightMultiCell(width, height float64, text, border, align string, fill bool) {
	// Сохраняем текущую позицию
	x, y := f.pdf.GetXY()
	
	// Рисуем фон ячейки если нужно
	if fill {
		f.pdf.Rect(x, y, width, height, "F")
	}
	
	// Рисуем границы
	f.setBorderColor(f.template.Colors.Border)
	f.pdf.Rect(x, y, width, height, "D")
	
	// Добавляем текст как MultiCell без границ и фона
	// Используем компактную высоту строки для многострочного текста
	lineHeight := f.template.TableRowHeight * 0.7 // Более компактная высота строки
	f.pdf.MultiCell(width, lineHeight, text, "", align, false)
	
	// Возвращаемся на правильную позицию для следующей ячейки
	f.pdf.SetXY(x+width, y)
}

// calculateRowHeight вычисляет необходимую высоту строки для текста
func (f *PDFFormatter) calculateRowHeight(text string, width float64) float64 {
	if len(text) <= 50 {
		return f.template.TableRowHeight
	}
	
	// Приблизительно вычисляем количество строк
	lines := f.estimateLines(text, width)
	
	// Используем компактную высоту: базовая высота + дополнительные строки с меньшим интервалом
	baseHeight := f.template.TableRowHeight
	additionalHeight := float64(lines-1) * (f.template.TableRowHeight * 0.6) // 60% от стандартной высоты для дополнительных строк
	
	return baseHeight + additionalHeight
}

// estimateLines приблизительно оценивает количество строк для переноса текста
func (f *PDFFormatter) estimateLines(text string, width float64) int {
	// Средняя ширина символа при текущем размере шрифта (приблизительно)
	charWidth := f.template.FontSize * 0.6
	charsPerLine := int(width / charWidth)
	
	if charsPerLine <= 0 {
		return 1
	}
	
	lines := (len(text) + charsPerLine - 1) / charsPerLine
	if lines < 1 {
		return 1
	}
	return lines
}

// wrapText добавляет переносы в длинный текст
func (f *PDFFormatter) wrapText(text string) string {
	if len(text) <= 50 {
		return text
	}
	
	// Попытка разбить по разделителям
	for _, sep := range []string{"_", "-", ".", " "} {
		if strings.Contains(text, sep) {
			parts := strings.Split(text, sep)
			var result strings.Builder
			currentLine := ""
			
			for i, part := range parts {
				testLine := currentLine + part
				if i < len(parts)-1 {
					testLine += sep
				}
				
				if len(testLine) <= 50 {
					currentLine = testLine
				} else {
					if currentLine != "" {
						result.WriteString(currentLine + "\n")
						currentLine = part
						if i < len(parts)-1 {
							currentLine += sep
						}
					} else {
						currentLine = testLine
					}
				}
			}
			if currentLine != "" {
				result.WriteString(currentLine)
			}
			return result.String()
		}
	}
	
	// Принудительный перенос каждые 50 символов
	var result strings.Builder
	for i := 0; i < len(text); i += 50 {
		end := i + 50
		if end > len(text) {
			end = len(text)
		}
		result.WriteString(text[i:end])
		if end < len(text) {
			result.WriteString("\n")
		}
	}
	return result.String()
}

// addMultiCell добавляет ячейку с многострочным текстом
func (f *PDFFormatter) addMultiCell(width, height float64, text, border, align string, fill bool) {
	f.pdf.MultiCell(width, f.template.TableRowHeight, text, border, align, fill)
}

// addControlledMultiCell добавляет MultiCell с фиксированной высотой
func (f *PDFFormatter) addControlledMultiCell(width, height float64, text, border, align string, fill bool) {
	// Создаем MultiCell с фиксированной высотой строки
	f.pdf.MultiCell(width, f.template.TableRowHeight, text, border, align, fill)
}

// addTableHeader добавляет заголовок таблицы
func (f *PDFFormatter) addTableHeader(headers []string, widths []float64) {
	f.setFillColor(f.template.Colors.TableHead)
	f.setTextColor(f.template.Colors.Text)
	f.pdf.SetFont(f.template.FontFamily, "B", f.template.FontSize)
	
	for i, header := range headers {
		f.pdf.CellFormat(widths[i], f.template.TableRowHeight, header, "1", 0, "C", true, 0, "")
	}
	f.pdf.Ln(-1)
	f.pdf.SetFont(f.template.FontFamily, "", f.template.FontSize)
}

// addMaterialsSummary добавляет сводку использованных материалов
func (f *PDFFormatter) addMaterialsSummary(data *parser.Parser3MF) {
	// Собираем уникальные материалы
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
		return
	}
	
	// Преобразуем в отсортированный список
	var materials []string
	for material := range materialsSet {
		materials = append(materials, material)
	}
	sort.Strings(materials)
	
	// Добавляем секцию материалов
	f.addSectionHeader("Materials Used")
	
	// Конфигурация для списка
	config := DefaultMaterialsConfig()
	
	f.pdf.SetFont(f.template.FontFamily, "", f.template.FontSize)
	f.setTextColor(f.template.Colors.Text)
	
	// Выводим материалы в две колонки
	x, y := f.pdf.GetXY()
	startX := x
	
	for i, material := range materials {
		// Проверяем переход на новую страницу
		_, pageH := f.pdf.GetPageSize()
		_, _, _, bottomMargin := f.pdf.GetMargins()
		if y+config.LineSpacing > pageH-bottomMargin {
			f.pdf.AddPage()
			x, y = f.pdf.GetXY()
			startX = x
		}
		
		// Переход на вторую колонку
		if i > 0 && i%10 == 0 {
			if x == startX {
				x += config.ColumnWidth
			} else {
				x = startX
				y += config.LineSpacing
			}
		}
		
		f.pdf.SetXY(x, y)
		
		// Маркер списка
		f.pdf.CellFormat(config.IndentSize, config.LineSpacing, "•", "", 0, "L", false, 0, "")
		
		// Текст материала
		f.pdf.CellFormat(config.ColumnWidth-config.IndentSize, config.LineSpacing, material, "", 0, "L", false, 0, "")
		
		y += config.LineSpacing
	}
}

// addSectionHeader добавляет заголовок секции
func (f *PDFFormatter) addSectionHeader(title string) {
	f.setTextColor(f.template.Colors.Header)
	f.pdf.SetFont(f.template.FontFamily, "B", f.template.HeaderFontSize)
	f.pdf.CellFormat(0, f.template.HeaderHeight*0.6, title, "", 1, "L", false, 0, "")
	
	// Горизонтальная линия под заголовком
	f.setBorderColor(f.template.Colors.Header)
	x, y := f.pdf.GetXY()
	f.pdf.Line(x, y-2, x+170, y-2)
	
	f.addVerticalSpace(5)
	f.setTextColor(f.template.Colors.Text)
	f.pdf.SetFont(f.template.FontFamily, "", f.template.FontSize)
}

// addText добавляет обычный текст
func (f *PDFFormatter) addText(text string, fontSize float64) {
	f.pdf.SetFont(f.template.FontFamily, "", fontSize)
	f.pdf.CellFormat(0, f.template.TableRowHeight, text, "", 1, "L", false, 0, "")
}

// addVerticalSpace добавляет вертикальный отступ
func (f *PDFFormatter) addVerticalSpace(height float64) {
	f.pdf.Ln(height)
}

// setTextColor устанавливает цвет текста
func (f *PDFFormatter) setTextColor(color [3]int) {
	f.pdf.SetTextColor(color[0], color[1], color[2])
}

// setFillColor устанавливает цвет заливки
func (f *PDFFormatter) setFillColor(color [3]int) {
	f.pdf.SetFillColor(color[0], color[1], color[2])
}

// setBorderColor устанавливает цвет границ
func (f *PDFFormatter) setBorderColor(color [3]int) {
	f.pdf.SetDrawColor(color[0], color[1], color[2])
}