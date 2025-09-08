package stl

import (
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/hschendel/stl"
)

// CalculateVolume вычисляет объем STL файла
func CalculateVolume(filePath string, config VolumeConfig) (*VolumeResult, error) {
	// Проверка существования файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("STL file not found: %s", filePath)
	}

	// Проверка расширения файла
	if !strings.HasSuffix(strings.ToLower(filePath), ".stl") {
		return nil, fmt.Errorf("file must have .stl extension: %s", filePath)
	}

	// Чтение STL файла
	solid, err := stl.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read STL file: %w", err)
	}

	// Конвертация в наш формат треугольников
	triangles := convertToTriangles(solid)
	if len(triangles) == 0 {
		return nil, fmt.Errorf("no triangles found in STL file")
	}

	// Вычисление объема с помощью алгоритма signed tetrahedron volumes
	volume := calculateMeshVolume(triangles)

	// Валидация - объем должен быть положительным
	isValid := volume > 0
	if !isValid {
		// Если объем отрицательный, вероятно нормали повернуты неправильно
		volume = math.Abs(volume)
	}

	// Конвертация единиц измерения
	convertedVolume, volumeUnit := convertVolumeUnits(volume, config.Units)

	// Определение плотности материала
	density := config.Density
	material := config.Material
	if density == 0 && material != "" {
		if d, exists := MaterialDensity[strings.ToUpper(material)]; exists {
			density = d
		} else {
			material = "Unknown"
		}
	}

	// Вычисление веса (если известна плотность)
	var weight float64
	if density > 0 {
		// Конвертируем объем в см³ для расчета веса
		volumeCm3 := convertToCm3(volume)
		weight = volumeCm3 * density
	}

	result := &VolumeResult{
		Volume:        convertedVolume,
		VolumeUnit:    volumeUnit,
		Triangles:     len(triangles),
		Weight:        weight,
		Material:      material,
		Density:       density,
		FilePath:      filePath,
		IsValid:       isValid,
	}

	return result, nil
}

// convertToTriangles конвертирует STL solid в наш формат треугольников
func convertToTriangles(solid *stl.Solid) []Triangle {
	triangles := make([]Triangle, len(solid.Triangles))
	
	for i, t := range solid.Triangles {
		triangles[i] = Triangle{
			V0: Vector3D{X: float64(t.Vertices[0][0]), Y: float64(t.Vertices[0][1]), Z: float64(t.Vertices[0][2])},
			V1: Vector3D{X: float64(t.Vertices[1][0]), Y: float64(t.Vertices[1][1]), Z: float64(t.Vertices[1][2])},
			V2: Vector3D{X: float64(t.Vertices[2][0]), Y: float64(t.Vertices[2][1]), Z: float64(t.Vertices[2][2])},
		}
	}
	
	return triangles
}

// calculateMeshVolume вычисляет объем mesh используя алгоритм signed tetrahedron volumes
func calculateMeshVolume(triangles []Triangle) float64 {
	volume := 0.0
	
	// Для каждого треугольника создаем тетраэдр от начала координат
	// и суммируем их signed volumes
	for _, triangle := range triangles {
		v0 := triangle.V0
		v1 := triangle.V1  
		v2 := triangle.V2
		
		// Signed volume of tetrahedron formed by origin and triangle
		// V = (1/6) * |det([v1-v0, v2-v0, -v0])|
		// Simplified: V = (1/6) * dot(v0, cross(v1, v2))
		cross := crossProduct(v1, v2)
		signedVolume := dotProduct(v0, cross) / 6.0
		
		volume += signedVolume
	}
	
	return math.Abs(volume)
}

// crossProduct вычисляет векторное произведение двух векторов
func crossProduct(a, b Vector3D) Vector3D {
	return Vector3D{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}

// dotProduct вычисляет скалярное произведение двух векторов
func dotProduct(a, b Vector3D) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

// convertVolumeUnits конвертирует объем в нужные единицы измерения
func convertVolumeUnits(volume float64, targetUnit string) (float64, string) {
	// Предполагаем, что исходные единицы - миллиметры (стандарт для STL)
	switch strings.ToLower(targetUnit) {
	case "mm3", "":
		return volume, "mm³"
	case "cm3":
		return volume / 1000.0, "cm³"
	case "in3":
		return volume / 16387.064, "in³"
	case "m3":
		return volume / 1000000000.0, "m³"
	default:
		return volume, "mm³"
	}
}

// convertToCm3 конвертирует объем из мм³ в см³
func convertToCm3(volumeMM3 float64) float64 {
	return volumeMM3 / 1000.0
}

// GetBoundingBox возвращает ограничивающий параллелепипед для STL файла
func GetBoundingBox(filePath string) (*BoundingBox, error) {
	solid, err := stl.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read STL file: %w", err)
	}

	if len(solid.Triangles) == 0 {
		return nil, fmt.Errorf("no triangles found in STL file")
	}

	// Инициализируем bbox первой вершиной
	firstVertex := solid.Triangles[0].Vertices[0]
	bbox := &BoundingBox{
		Min: Vector3D{X: float64(firstVertex[0]), Y: float64(firstVertex[1]), Z: float64(firstVertex[2])},
		Max: Vector3D{X: float64(firstVertex[0]), Y: float64(firstVertex[1]), Z: float64(firstVertex[2])},
	}

	// Проходим по всем вершинам
	for _, triangle := range solid.Triangles {
		for _, vertex := range triangle.Vertices {
			x, y, z := float64(vertex[0]), float64(vertex[1]), float64(vertex[2])
			
			if x < bbox.Min.X { bbox.Min.X = x }
			if y < bbox.Min.Y { bbox.Min.Y = y }
			if z < bbox.Min.Z { bbox.Min.Z = z }
			
			if x > bbox.Max.X { bbox.Max.X = x }
			if y > bbox.Max.Y { bbox.Max.Y = y }
			if z > bbox.Max.Z { bbox.Max.Z = z }
		}
	}

	return bbox, nil
}

// GetDimensions возвращает размеры модели в миллиметрах
func (bbox *BoundingBox) GetDimensions() (width, depth, height float64) {
	return bbox.Max.X - bbox.Min.X,
		   bbox.Max.Y - bbox.Min.Y,
		   bbox.Max.Z - bbox.Min.Z
}