package formatter

// PDFTemplate определяет конфигурацию макета для PDF документа
type PDFTemplate struct {
	PageSize        string
	Orientation     string
	FontFamily      string
	FontSize        float64
	HeaderFontSize  float64
	TitleFontSize   float64
	MarginX         float64
	MarginY         float64
	HeaderHeight    float64
	TableRowHeight  float64
	SectionSpacing  float64
	Colors          PDFColors
}

// PDFColors определяет цветовую схему PDF документа
type PDFColors struct {
	Header    [3]int // RGB для заголовков
	Title     [3]int // RGB для заголовка документа
	TableHead [3]int // RGB для заголовка таблицы
	TableRow1 [3]int // RGB для нечетных строк таблицы
	TableRow2 [3]int // RGB для четных строк таблицы
	Text      [3]int // RGB для обычного текста
	Border    [3]int // RGB для границ таблицы
}

// DefaultPDFTemplate возвращает конфигурацию PDF по умолчанию
func DefaultPDFTemplate() PDFTemplate {
	return PDFTemplate{
		PageSize:        "A4",
		Orientation:     "P", // Portrait
		FontFamily:      "DejaVuSans",
		FontSize:        10,
		HeaderFontSize:  12,
		TitleFontSize:   16,
		MarginX:         20,
		MarginY:         20,
		HeaderHeight:    25,
		TableRowHeight:  8,
		SectionSpacing:  10,
		Colors: PDFColors{
			Header:    [3]int{52, 73, 94},   // Темно-синий
			Title:     [3]int{44, 62, 80},   // Еще темнее синий
			TableHead: [3]int{149, 165, 166}, // Серый
			TableRow1: [3]int{255, 255, 255}, // Белый
			TableRow2: [3]int{236, 240, 241}, // Светло-серый
			Text:      [3]int{52, 73, 94},   // Темно-синий для текста
			Border:    [3]int{189, 195, 199}, // Светло-серый для границ
		},
	}
}

// TableColumnWidths определяет ширину колонок для различных таблиц
type TableColumnWidths struct {
	ObjectName float64
	Count      float64
	Type       float64
	Material   float64
}

// DefaultTableWidths возвращает стандартные ширины колонок для A4 портрет
func DefaultTableWidths() TableColumnWidths {
	return TableColumnWidths{
		ObjectName: 85,  // Название объекта (увеличено для длинных названий)
		Count:      18,  // Количество (уменьшено)
		Type:       22,  // Тип (mesh/assembly) (уменьшено)
		Material:   45,  // Материал (уменьшено)
	}
}

// MaterialsListConfig определяет конфигурацию для списка материалов
type MaterialsListConfig struct {
	BulletSize   float64
	LineSpacing  float64
	IndentSize   float64
	ColumnWidth  float64
	ColumnsCount int
}

// DefaultMaterialsConfig возвращает стандартную конфигурацию для списка материалов
func DefaultMaterialsConfig() MaterialsListConfig {
	return MaterialsListConfig{
		BulletSize:   3,
		LineSpacing:  5,
		IndentSize:   5,
		ColumnWidth:  85,
		ColumnsCount: 2,
	}
}