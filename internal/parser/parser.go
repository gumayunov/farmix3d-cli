package parser

import (
	"fmt"
	"strings"
)

func Parse3MF(filePath string) (*Parser3MF, error) {
	extractDir, err := ExtractArchive(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract 3MF archive: %w", err)
	}
	defer CleanupTemp(extractDir)

	model, err := ParseModel3D(extractDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse main model: %w", err)
	}

	settings, err := ParseModelSettings(extractDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse model settings: %w", err)
	}

	result := &Parser3MF{}

	plateMap := make(map[int]*PlateInfo)
	for _, plate := range settings.Plates {
		plateMap[plate.PlaterID] = &PlateInfo{
			PlateID:   plate.PlaterID,
			PlateName: plate.PlaterName,
			Objects:   []PlateObject{},
		}
	}

	objectNameMap := buildObjectMetadataMap(settings.Objects)

	partNameMap := make(map[int]string)
	partFileMap := make(map[int]string)
	for _, part := range settings.Parts {
		partNameMap[part.ID] = part.Name
		partFileMap[part.ID] = part.SourceFile
	}

	instanceToPlateMap := make(map[int]int)
	for _, instance := range settings.Instances {
		instanceToPlateMap[instance.ObjectID] = instance.InstanceID
	}

	modelObjectMap := make(map[int]*ModelObject)
	for i := range model.Resources {
		obj := &model.Resources[i]
		modelObjectMap[obj.ID] = obj
	}

	for _, buildItem := range model.Build {
		modelObj := modelObjectMap[buildItem.ObjectID]
		if modelObj == nil {
			continue
		}

		printable := true
		if buildItem.Printable != nil {
			printable = *buildItem.Printable
		}

		plateObject := PlateObject{
			ID:        buildItem.ObjectID,
			Name:      getObjectName(buildItem.ObjectID, objectNameMap, partNameMap, modelObj),
			Position:  ParseTransform(buildItem.Transform),
			Printable: printable,
		}

		if modelObj.Mesh != nil {
			plateObject.Type = "mesh"
		} else if modelObj.Components != nil {
			plateObject.Type = "assembly"
			components, err := processAssemblyComponents(extractDir, modelObj.Components, partNameMap, partFileMap)
			if err != nil {
				return nil, fmt.Errorf("failed to process assembly components: %w", err)
			}
			plateObject.Components = components
		}

		plateID := findPlateForObject(buildItem.ObjectID, instanceToPlateMap, plateMap)
		if plate, exists := plateMap[plateID]; exists {
			plate.Objects = append(plate.Objects, plateObject)
		}
	}

	for _, plate := range plateMap {
		result.Plates = append(result.Plates, *plate)
	}

	return result, nil
}

func getObjectName(objectID int, objectNameMap, partNameMap map[int]string, modelObj *ModelObject) string {
	if name, exists := objectNameMap[objectID]; exists && name != "" {
		return name
	}
	if name, exists := partNameMap[objectID]; exists && name != "" {
		return name
	}
	if modelObj.Name != "" {
		return modelObj.Name
	}
	return fmt.Sprintf("Object_%d", objectID)
}

func findPlateForObject(objectID int, instanceToPlateMap map[int]int, plateMap map[int]*PlateInfo) int {
	if plateID, exists := instanceToPlateMap[objectID]; exists {
		if _, plateExists := plateMap[plateID]; plateExists {
			return plateID
		}
	}
	
	for plateID := range plateMap {
		return plateID
	}
	return 0
}

func processAssemblyComponents(extractDir string, components *ComponentsCollection, partNameMap, partFileMap map[int]string) ([]ComponentInfo, error) {
	var result []ComponentInfo

	for _, comp := range components.Components {
		compInfo := ComponentInfo{
			ID:        comp.ObjectID,
			Transform: ParseTransform(comp.Transform),
		}

		if name, exists := partNameMap[comp.ObjectID]; exists {
			compInfo.Name = name
		} else {
			compInfo.Name = fmt.Sprintf("Component_%d", comp.ObjectID)
		}

		if comp.Path != "" {
			compInfo.SourceFile = comp.Path
			
			assemblyModel, err := ParseAssemblyModel(extractDir, comp.Path)
			if err == nil && len(assemblyModel.Resources) > 0 {
				for _, res := range assemblyModel.Resources {
					if res.ID == comp.ObjectID && res.Name != "" {
						compInfo.Name = res.Name
						break
					}
				}
			}
		} else if sourceFile, exists := partFileMap[comp.ObjectID]; exists {
			compInfo.SourceFile = sourceFile
		}

		if !isEmptyAssembly(extractDir, compInfo.SourceFile) {
			result = append(result, compInfo)
		}
	}

	return result, nil
}

func isEmptyAssembly(extractDir, assemblyPath string) bool {
	if assemblyPath == "" {
		return true
	}

	if !strings.HasSuffix(assemblyPath, ".model") {
		return false
	}

	assemblyModel, err := ParseAssemblyModel(extractDir, assemblyPath)
	if err != nil {
		return true
	}

	return len(assemblyModel.Resources) == 0
}