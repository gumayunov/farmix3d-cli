package cmd

import (
	"fmt"
	"os"
	"strings"

	"farmix-cli/internal/stl"

	"github.com/spf13/cobra"
)

var (
	volumeUnits    string
	volumeFormat   string
	volumeMaterial string
	volumeDensity  float64
	showBounds     bool
)

var volumeCmd = &cobra.Command{
	Use:   "volume [STL файл]",
	Short: "Вычисление объема и веса 3D модели из STL файла",
	Long: `Вычисляет объем и приблизительный вес 3D модели из STL файла.
Команда использует алгоритмы расчета объема mesh для точного вычисления
объемных характеристик и может оценить вес материала на основе плотности.

Расчет использует метод signed tetrahedron volume, который работает корректно
для замкнутых manifold mesh объектов. Незамкнутые или поврежденные модели
могут давать неточные результаты.

Примеры использования:
  farmix-cli volume модель.stl
  farmix-cli volume --units cm3 --material PLA модель.stl
  farmix-cli volume --format json --density 1.04 модель.stl
  farmix-cli volume --show-bounds модель.stl`,
	Args: cobra.ExactArgs(1),
	Run:  runVolumeCommand,
}

func runVolumeCommand(cmd *cobra.Command, args []string) {
	stlFile := args[0]

	// Валидация параметров
	if err := validateVolumeParams(stlFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Создание конфигурации
	config := stl.VolumeConfig{
		Units:    volumeUnits,
		Material: volumeMaterial,
		Density:  volumeDensity,
	}

	// Вычисление объема
	fmt.Printf("Вычисление объема для %s...\n", stlFile)
	result, err := stl.CalculateVolume(stlFile, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка расчета объема: %v\n", err)
		os.Exit(1)
	}

	// Получение размеров если требуется
	var bbox *stl.BoundingBox
	if showBounds {
		bbox, err = stl.GetBoundingBox(stlFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Предупреждение: Не удалось получить габариты: %v\n", err)
		}
	}

	// Вывод результатов
	if err := outputVolumeResult(result, bbox); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка форматирования вывода: %v\n", err)
		os.Exit(1)
	}
}

func validateVolumeParams(stlFile string) error {
	// Проверка STL файла
	if !strings.HasSuffix(strings.ToLower(stlFile), ".stl") {
		return fmt.Errorf("файл должен иметь расширение .stl: %s", stlFile)
	}

	if _, err := os.Stat(stlFile); os.IsNotExist(err) {
		return fmt.Errorf("STL файл не найден: %s", stlFile)
	}

	// Проверка единиц измерения
	validUnits := map[string]bool{
		"mm3": true, "cm3": true, "in3": true, "m3": true, "": true,
	}
	if !validUnits[strings.ToLower(volumeUnits)] {
		return fmt.Errorf("неверные единицы измерения: %s. Допустимые единицы: mm3, cm3, in3, m3", volumeUnits)
	}

	// Проверка формата вывода
	validFormats := map[string]bool{
		"text": true, "csv": true, "json": true, "": true,
	}
	if !validFormats[strings.ToLower(volumeFormat)] {
		return fmt.Errorf("неверный формат: %s. Допустимые форматы: text, csv, json", volumeFormat)
	}

	// Проверка плотности
	if volumeDensity < 0 {
		return fmt.Errorf("плотность не может быть отрицательной: %f", volumeDensity)
	}

	return nil
}

func outputVolumeResult(result *stl.VolumeResult, bbox *stl.BoundingBox) error {
	switch strings.ToLower(volumeFormat) {
	case "json":
		return outputVolumeJSON(result, bbox)
	case "csv":
		return outputVolumeCSV(result, bbox)
	case "text", "":
		return outputVolumeText(result, bbox)
	default:
		return fmt.Errorf("неподдерживаемый формат вывода: %s", volumeFormat)
	}
}

func outputVolumeText(result *stl.VolumeResult, bbox *stl.BoundingBox) error {
	fmt.Println("=== Volume Analysis Results ===")
	
	fmt.Printf("File: %s\n", result.FilePath)
	fmt.Printf("Valid Mesh: %v\n", result.IsValid)
	if !result.IsValid {
		fmt.Println("Warning: Mesh may have incorrect winding order")
	}
	
	fmt.Printf("Volume: %.4f %s\n", result.Volume, result.VolumeUnit)
	fmt.Printf("Triangles: %d\n", result.Triangles)
	
	if result.Weight > 0 {
		fmt.Printf("Material: %s\n", result.Material)
		fmt.Printf("Density: %.2f g/cm³\n", result.Density)
		fmt.Printf("Estimated Weight: %.2f grams\n", result.Weight)
	}
	
	if bbox != nil {
		width, depth, height := bbox.GetDimensions()
		fmt.Printf("\nBounding Box:\n")
		fmt.Printf("  Width:  %.2f mm\n", width)
		fmt.Printf("  Depth:  %.2f mm\n", depth) 
		fmt.Printf("  Height: %.2f mm\n", height)
		fmt.Printf("  Min: (%.2f, %.2f, %.2f)\n", bbox.Min.X, bbox.Min.Y, bbox.Min.Z)
		fmt.Printf("  Max: (%.2f, %.2f, %.2f)\n", bbox.Max.X, bbox.Max.Y, bbox.Max.Z)
	}
	
	return nil
}

func outputVolumeCSV(result *stl.VolumeResult, bbox *stl.BoundingBox) error {
	// CSV заголовок
	header := "file,volume,volume_unit,triangles,weight,material,density,is_valid"
	if bbox != nil {
		header += ",width,depth,height,min_x,min_y,min_z,max_x,max_y,max_z"
	}
	fmt.Println(header)
	
	// CSV данные
	csvLine := fmt.Sprintf("%s,%.4f,%s,%d,%.2f,%s,%.2f,%v",
		result.FilePath, result.Volume, result.VolumeUnit, 
		result.Triangles, result.Weight, result.Material, 
		result.Density, result.IsValid)
	
	if bbox != nil {
		width, depth, height := bbox.GetDimensions()
		csvLine += fmt.Sprintf(",%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f,%.2f",
			width, depth, height,
			bbox.Min.X, bbox.Min.Y, bbox.Min.Z,
			bbox.Max.X, bbox.Max.Y, bbox.Max.Z)
	}
	
	fmt.Println(csvLine)
	return nil
}

func outputVolumeJSON(result *stl.VolumeResult, bbox *stl.BoundingBox) error {
	// Простой JSON вывод
	fmt.Printf(`{
  "file": "%s",
  "volume": %.4f,
  "volume_unit": "%s",
  "triangles": %d,
  "weight": %.2f,
  "material": "%s",
  "density": %.2f,
  "is_valid": %t`,
		result.FilePath, result.Volume, result.VolumeUnit,
		result.Triangles, result.Weight, result.Material,
		result.Density, result.IsValid)

	if bbox != nil {
		width, depth, height := bbox.GetDimensions()
		fmt.Printf(`,
  "bounding_box": {
    "width": %.2f,
    "depth": %.2f,
    "height": %.2f,
    "min": {"x": %.2f, "y": %.2f, "z": %.2f},
    "max": {"x": %.2f, "y": %.2f, "z": %.2f}
  }`,
			width, depth, height,
			bbox.Min.X, bbox.Min.Y, bbox.Min.Z,
			bbox.Max.X, bbox.Max.Y, bbox.Max.Z)
	}

	fmt.Println("\n}")
	return nil
}

func init() {
	volumeCmd.Flags().StringVarP(&volumeUnits, "units", "u", "mm3", "Единицы объема (mm3, cm3, in3, m3)")
	volumeCmd.Flags().StringVarP(&volumeFormat, "format", "f", "text", "Формат вывода (text, csv, json)")
	volumeCmd.Flags().StringVarP(&volumeMaterial, "material", "m", "", "Тип материала (PLA, ABS, PETG и т.д.)")
	volumeCmd.Flags().Float64VarP(&volumeDensity, "density", "d", 0, "Плотность материала в г/см³ (переопределяет материал)")
	volumeCmd.Flags().BoolVar(&showBounds, "show-bounds", false, "Включить размеры габаритного параллелепипеда")
	
	rootCmd.AddCommand(volumeCmd)
}