package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func ParseModel3D(extractDir string) (*Model3D, error) {
	modelPath := filepath.Join(extractDir, "3D", "3dmodel.model")
	
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("main model file not found: %s", modelPath)
	}

	data, err := os.ReadFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read model file: %w", err)
	}

	var model Model3D
	if err := xml.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse model XML: %w", err)
	}

	return &model, nil
}

func ParseModelSettings(extractDir string) (*ModelSettings, error) {
	settingsPath := filepath.Join(extractDir, "Metadata", "model_settings.config")
	
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("model settings file not found: %s", settingsPath)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings file: %w", err)
	}

	var settings ModelSettings
	if err := xml.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings XML: %w", err)
	}

	return &settings, nil
}

func ParseAssemblyModel(extractDir, assemblyPath string) (*Model3D, error) {
	fullPath := filepath.Join(extractDir, assemblyPath)
	
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("assembly file not found: %s", fullPath)
	}

	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read assembly file: %w", err)
	}

	var model Model3D
	if err := xml.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse assembly XML: %w", err)
	}

	return &model, nil
}

func ParseTransform(transformStr string) Transform3D {
	var transform Transform3D
	
	if transformStr == "" {
		transform.Matrix = [12]float64{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}
		return transform
	}

	parts := strings.Fields(transformStr)
	if len(parts) != 12 {
		transform.Matrix = [12]float64{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}
		return transform
	}

	for i, part := range parts {
		if val, err := strconv.ParseFloat(part, 64); err == nil {
			transform.Matrix[i] = val
		}
	}

	return transform
}