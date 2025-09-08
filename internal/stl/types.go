package stl

// VolumeResult содержит результат вычисления объема STL модели
type VolumeResult struct {
	Volume        float64 `json:"volume"`         // Объем в кубических единицах файла
	VolumeUnit    string  `json:"volume_unit"`    // Единица измерения объема
	Triangles     int     `json:"triangles"`      // Количество треугольников
	Weight        float64 `json:"weight"`         // Вес в граммах (если указана плотность)
	Material      string  `json:"material"`       // Тип материала
	Density       float64 `json:"density"`        // Плотность материала г/см³
	FilePath      string  `json:"file_path"`      // Путь к исходному файлу
	IsValid       bool    `json:"is_valid"`       // Валидность модели (замкнутая поверхность)
}

// MaterialDensity содержит плотности популярных 3D материалов (г/см³)
var MaterialDensity = map[string]float64{
	"PLA":      1.24,
	"ABS":      1.04,
	"PETG":     1.27,
	"TPU":      1.20,
	"WOOD":     1.28,
	"BRONZE":   1.50,
	"STEEL":    1.37,
	"COPPER":   1.50,
	"ASA":      1.05,
	"PC":       1.20,
	"NYLON":    1.15,
	"PP":       0.90,
	"PVA":      1.19,
	"HIPS":     1.04,
}

// Vector3D представляет трехмерный вектор
type Vector3D struct {
	X, Y, Z float64
}

// Triangle представляет треугольник с тремя вершинами
type Triangle struct {
	V0, V1, V2 Vector3D
}

// BoundingBox представляет ограничивающий параллелепипед
type BoundingBox struct {
	Min, Max Vector3D
}

// VolumeConfig содержит конфигурацию для расчета объема
type VolumeConfig struct {
	Units    string  // mm3, cm3, in3, m3
	Material string  // название материала
	Density  float64 // плотность в г/см³ (переопределяет материал)
}