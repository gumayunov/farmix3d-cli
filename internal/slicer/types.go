package slicer

import "time"

// SliceResult содержит результат слайсинга модели
type SliceResult struct {
	FilamentUsed     FilamentUsage `json:"filament_used"`
	PrintTime        time.Duration `json:"print_time"`
	LayerCount       int           `json:"layer_count"`
	LayerHeight      float64       `json:"layer_height"`
	OutputFile       string        `json:"output_file"`
	SlicingSuccess   bool          `json:"slicing_success"`
	ErrorMessage     string        `json:"error_message,omitempty"`
}

// FilamentUsage содержит информацию о расходе филамента
type FilamentUsage struct {
	LengthMM     float64 `json:"length_mm"`     // Длина филамента в миллиметрах
	WeightGrams  float64 `json:"weight_grams"`  // Вес филамента в граммах
	VolumeMM3    float64 `json:"volume_mm3"`    // Объем филамента в мм³
	MaterialType string  `json:"material_type"` // Тип материала (PLA, ABS, PETG и т.д.)
}

// SliceConfig содержит конфигурацию для слайсинга
type SliceConfig struct {
	OrcaPath       string            `json:"orca_path"`       // Путь к исполняемому файлу OrcaSlicer
	STLFile        string            `json:"stl_file"`        // Путь к STL файлу
	OutputDir      string            `json:"output_dir"`      // Директория для сохранения результата
	PrinterProfile string            `json:"printer_profile"` // Профиль принтера
	MaterialProfile string           `json:"material_profile"` // Профиль материала
	PrintProfile   string            `json:"print_profile"`   // Профиль печати
	ExtraParams    map[string]string `json:"extra_params"`    // Дополнительные параметры
}

// GCodeStats содержит статистику, извлеченную из G-code
type GCodeStats struct {
	FilamentLengthMM []float64     `json:"filament_length_mm"` // Длина для каждого экструдера
	FilamentWeightG  []float64     `json:"filament_weight_g"`  // Вес для каждого экструдера
	PrintTime        time.Duration `json:"print_time"`         // Время печати
	LayerCount       int           `json:"layer_count"`        // Количество слоев
	LayerHeight      float64       `json:"layer_height"`       // Высота слоя
	MaterialTypes    []string      `json:"material_types"`     // Типы материалов
}

// SlicerError представляет ошибку слайсера
type SlicerError struct {
	Type    string `json:"type"`    // Тип ошибки
	Message string `json:"message"` // Сообщение об ошибке
	Code    int    `json:"code"`    // Код выхода программы
}

func (e *SlicerError) Error() string {
	return e.Message
}