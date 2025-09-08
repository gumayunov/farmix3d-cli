package slicer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseGCodeFile парсит G-code файл и извлекает статистику
func ParseGCodeFile(filePath string) (*GCodeStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open G-code file: %w", err)
	}
	defer file.Close()

	stats := &GCodeStats{
		FilamentLengthMM: make([]float64, 0),
		FilamentWeightG:  make([]float64, 0),
		MaterialTypes:    make([]string, 0),
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Пропускаем пустые строки и не-комментарии
		if !strings.HasPrefix(line, ";") {
			continue
		}

		// Убираем символ комментария и лишние пробелы
		comment := strings.TrimSpace(line[1:])
		
		// Парсим различные поля
		parseFilamentLength(comment, stats)
		parseFilamentWeight(comment, stats)
		parsePrintTime(comment, stats)
		parseLayerInfo(comment, stats)
		parseMaterialType(comment, stats)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading G-code file: %w", err)
	}

	// Если не нашли данных о филаменте, возвращаем ошибку
	if len(stats.FilamentLengthMM) == 0 && len(stats.FilamentWeightG) == 0 {
		return nil, fmt.Errorf("no filament usage data found in G-code")
	}

	return stats, nil
}

// parseFilamentLength ищет информацию о длине филамента
func parseFilamentLength(comment string, stats *GCodeStats) {
	patterns := []string{
		`filament\s+used\s*\[mm\]\s*=\s*([\d.]+)`,
		`filament\s+length\s*\[mm\]\s*=\s*([\d.]+)`,
		`filament_used_mm\s*=\s*([\d.]+)`,
		`Total\s+filament\s+used\s*:\s*([\d.]+)\s*mm`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(comment); len(matches) > 1 {
			if length, err := strconv.ParseFloat(matches[1], 64); err == nil {
				stats.FilamentLengthMM = append(stats.FilamentLengthMM, length)
				return
			}
		}
	}

	// Проверяем формат с несколькими экструдерами
	if parseMultiExtruderLength(comment, stats) {
		return
	}
}

// parseFilamentWeight ищет информацию о весе филамента
func parseFilamentWeight(comment string, stats *GCodeStats) {
	patterns := []string{
		`filament\s+used\s*\[g\]\s*=\s*([\d.]+)`,
		`filament\s+weight\s*\[g\]\s*=\s*([\d.]+)`,
		`filament_used_g\s*=\s*([\d.]+)`,
		`Total\s+filament\s+weight\s*:\s*([\d.]+)\s*g`,
		`material\s+weight\s*=\s*([\d.]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(comment); len(matches) > 1 {
			if weight, err := strconv.ParseFloat(matches[1], 64); err == nil {
				stats.FilamentWeightG = append(stats.FilamentWeightG, weight)
				return
			}
		}
	}

	// Проверяем формат с несколькими экструдерами
	parseMultiExtruderWeight(comment, stats)
}

// parseMultiExtruderLength парсит длину филамента для нескольких экструдеров
func parseMultiExtruderLength(comment string, stats *GCodeStats) bool {
	// Формат: ; filament used [mm] = 100.5, 200.3, 150.1
	re := regexp.MustCompile(`(?i)filament\s+used\s*\[mm\]\s*=\s*([\d.,\s]+)`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) < 2 {
		return false
	}

	values := strings.Split(matches[1], ",")
	for _, value := range values {
		value = strings.TrimSpace(value)
		if length, err := strconv.ParseFloat(value, 64); err == nil {
			stats.FilamentLengthMM = append(stats.FilamentLengthMM, length)
		}
	}

	return len(values) > 1
}

// parseMultiExtruderWeight парсит вес филамента для нескольких экструдеров
func parseMultiExtruderWeight(comment string, stats *GCodeStats) bool {
	// Формат: ; filament used [g] = 10.5, 20.3, 15.1
	re := regexp.MustCompile(`(?i)filament\s+used\s*\[g\]\s*=\s*([\d.,\s]+)`)
	matches := re.FindStringSubmatch(comment)
	if len(matches) < 2 {
		return false
	}

	values := strings.Split(matches[1], ",")
	for _, value := range values {
		value = strings.TrimSpace(value)
		if weight, err := strconv.ParseFloat(value, 64); err == nil {
			stats.FilamentWeightG = append(stats.FilamentWeightG, weight)
		}
	}

	return len(values) > 1
}

// parsePrintTime ищет информацию о времени печати
func parsePrintTime(comment string, stats *GCodeStats) {
	patterns := []string{
		`estimated\s+printing\s+time\s*=\s*(\d+)h\s*(\d+)m\s*(\d+)s`,
		`print\s+time\s*:\s*(\d+)h\s*(\d+)m\s*(\d+)s`,
		`total\s+print\s+time\s*:\s*(\d+):(\d+):(\d+)`,
		`TIME:(\d+)`,
	}

	// Формат: часы, минуты, секунды
	for i, pattern := range patterns[:3] {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(comment); len(matches) > 3 {
			hours, _ := strconv.Atoi(matches[1])
			minutes, _ := strconv.Atoi(matches[2])
			seconds, _ := strconv.Atoi(matches[3])
			
			if i == 2 { // формат HH:MM:SS
				stats.PrintTime = time.Duration(hours)*time.Hour + 
								 time.Duration(minutes)*time.Minute + 
								 time.Duration(seconds)*time.Second
			} else {
				stats.PrintTime = time.Duration(hours)*time.Hour + 
								 time.Duration(minutes)*time.Minute + 
								 time.Duration(seconds)*time.Second
			}
			return
		}
	}

	// Формат: только секунды
	re := regexp.MustCompile(`(?i)` + patterns[3])
	if matches := re.FindStringSubmatch(comment); len(matches) > 1 {
		if seconds, err := strconv.Atoi(matches[1]); err == nil {
			stats.PrintTime = time.Duration(seconds) * time.Second
		}
	}
}

// parseLayerInfo ищет информацию о слоях
func parseLayerInfo(comment string, stats *GCodeStats) {
	// Количество слоев
	patterns := []string{
		`layer\s+count\s*:\s*(\d+)`,
		`total\s+layers\s*:\s*(\d+)`,
		`layers\s*=\s*(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(comment); len(matches) > 1 {
			if layers, err := strconv.Atoi(matches[1]); err == nil {
				stats.LayerCount = layers
				break
			}
		}
	}

	// Высота слоя
	heightPatterns := []string{
		`layer\s+height\s*:\s*([\d.]+)`,
		`layer_height\s*=\s*([\d.]+)`,
		`first\s+layer\s+height\s*:\s*([\d.]+)`,
	}

	for _, pattern := range heightPatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(comment); len(matches) > 1 {
			if height, err := strconv.ParseFloat(matches[1], 64); err == nil {
				stats.LayerHeight = height
				break
			}
		}
	}
}

// parseMaterialType ищет информацию о типе материала
func parseMaterialType(comment string, stats *GCodeStats) {
	patterns := []string{
		`filament_type\s*=\s*([A-Z]+)`,
		`material\s*:\s*([A-Z]+)`,
		`filament\s*:\s*([A-Z]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if matches := re.FindStringSubmatch(comment); len(matches) > 1 {
			material := strings.ToUpper(strings.TrimSpace(matches[1]))
			// Проверяем, что такой материал еще не добавлен
			for _, existing := range stats.MaterialTypes {
				if existing == material {
					return
				}
			}
			stats.MaterialTypes = append(stats.MaterialTypes, material)
			return
		}
	}
}