package slicer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SliceSTL выполняет слайсинг STL файла с помощью OrcaSlicer
func SliceSTL(config SliceConfig) (*SliceResult, error) {
	// Валидация входных параметров
	if err := validateConfig(config); err != nil {
		return nil, err
	}

	// Создаем временную директорию для вывода
	if config.OutputDir == "" {
		tempDir, err := os.MkdirTemp("", "3mfanalyzer_slice_*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp directory: %w", err)
		}
		config.OutputDir = tempDir
		defer os.RemoveAll(tempDir)
	}

	// Определяем имя выходного файла (OrcaSlicer создает файлы с автоматическими именами)
	baseName := strings.TrimSuffix(filepath.Base(config.STLFile), ".stl")
	outputFile := filepath.Join(config.OutputDir, baseName+".gcode")

	// Выполняем слайсинг
	if err := executeOrcaSlicer(config, outputFile); err != nil {
		return &SliceResult{
			SlicingSuccess: false,
			ErrorMessage:   err.Error(),
		}, err
	}

	// Найдем созданный G-code файл (OrcaSlicer может создать файл с другим именем)
	actualOutputFile, err := findGCodeFile(config.OutputDir, baseName)
	if err != nil {
		return &SliceResult{
			SlicingSuccess: false,
			ErrorMessage:   fmt.Sprintf("G-code file not found: %v", err),
		}, err
	}

	// Парсим результат
	stats, err := ParseGCodeFile(actualOutputFile)
	if err != nil {
		return &SliceResult{
			SlicingSuccess: false,
			ErrorMessage:   fmt.Sprintf("failed to parse G-code: %v", err),
			OutputFile:     actualOutputFile,
		}, err
	}

	// Формируем результат
	result := &SliceResult{
		FilamentUsed: FilamentUsage{
			LengthMM:     sumFloats(stats.FilamentLengthMM),
			WeightGrams:  sumFloats(stats.FilamentWeightG),
			MaterialType: strings.Join(stats.MaterialTypes, ", "),
		},
		PrintTime:      stats.PrintTime,
		LayerCount:     stats.LayerCount,
		LayerHeight:    stats.LayerHeight,
		OutputFile:     actualOutputFile,
		SlicingSuccess: true,
	}

	return result, nil
}

// validateConfig проверяет корректность конфигурации
func validateConfig(config SliceConfig) error {
	// Проверяем существование OrcaSlicer
	if config.OrcaPath == "" {
		return fmt.Errorf("OrcaSlicer path is required")
	}
	
	if _, err := os.Stat(config.OrcaPath); os.IsNotExist(err) {
		return fmt.Errorf("OrcaSlicer not found at path: %s", config.OrcaPath)
	}

	// Проверяем STL файл
	if config.STLFile == "" {
		return fmt.Errorf("STL file path is required")
	}

	if !strings.HasSuffix(strings.ToLower(config.STLFile), ".stl") {
		return fmt.Errorf("file must have .stl extension: %s", config.STLFile)
	}

	if _, err := os.Stat(config.STLFile); os.IsNotExist(err) {
		return fmt.Errorf("STL file not found: %s", config.STLFile)
	}

	return nil
}

// executeOrcaSlicer запускает OrcaSlicer с заданными параметрами
func executeOrcaSlicer(config SliceConfig, outputFile string) error {
	args := []string{
		"--slice", "0", // slice all plates
		"--outputdir", config.OutputDir,
		"--no-check", // skip validation checks
		config.STLFile,
	}

	// Добавляем профили если указаны
	if config.PrinterProfile != "" {
		args = append(args, "--load-settings", config.PrinterProfile)
	}
	if config.MaterialProfile != "" {
		args = append(args, "--load-filaments", config.MaterialProfile)
	}
	if config.PrintProfile != "" {
		args = append(args, "--load-settings", config.PrintProfile)
	}

	// Добавляем дополнительные параметры
	for key, value := range config.ExtraParams {
		args = append(args, fmt.Sprintf("--%s", key))
		if value != "" {
			args = append(args, value)
		}
	}

	// Выполняем команду
	cmd := exec.Command(config.OrcaPath, args...)
	cmd.Dir = config.OutputDir

	output, err := cmd.CombinedOutput()
	exitCode := getExitCode(err)
	
	// Если есть вывод, показываем его (для отладки)
	if len(output) > 0 {
		fmt.Printf("OrcaSlicer output (exit code %d):\n%s\n", exitCode, string(output))
	}
	
	// Не считаем ошибкой, если OrcaSlicer завершился с предупреждениями
	// но создал выходные файлы (коды выхода 205, 239 часто означают предупреждения)
	if err != nil && exitCode != 205 && exitCode != 239 {
		return &SlicerError{
			Type:    "execution_error",
			Message: fmt.Sprintf("OrcaSlicer execution failed: %v\nOutput: %s", err, string(output)),
			Code:    exitCode,
		}
	}

	return nil
}

// findGCodeFile ищет созданный G-code файл в указанной директории
func findGCodeFile(outputDir, baseName string) (string, error) {
	// Возможные варианты имен файлов, которые может создать OrcaSlicer
	possibleNames := []string{
		baseName + ".gcode",
		baseName + "_0.gcode", 
		baseName + "_plate_1.gcode",
		"plate_1.gcode",
	}

	// Ищем первый существующий файл
	for _, name := range possibleNames {
		fullPath := filepath.Join(outputDir, name)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	// Если точное совпадение не найдено, ищем любой .gcode файл в директории
	files, err := filepath.Glob(filepath.Join(outputDir, "*.gcode"))
	if err != nil {
		return "", fmt.Errorf("failed to search for G-code files: %w", err)
	}

	if len(files) > 0 {
		return files[0], nil
	}

	return "", fmt.Errorf("no G-code files found in %s", outputDir)
}

// getExitCode извлекает код выхода из ошибки exec
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}
	return 1
}

// sumFloats суммирует массив float64
func sumFloats(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum
}

// CreateDefaultConfig создает конфигурацию по умолчанию
func CreateDefaultConfig(orcaPath, stlFile string) SliceConfig {
	return SliceConfig{
		OrcaPath:    orcaPath,
		STLFile:     stlFile,
		ExtraParams: make(map[string]string),
	}
}