package parser

import (
	"fmt"
	"strconv"
)

func extractMetadataValue(metadata []MetadataEntry, key string) string {
	for _, entry := range metadata {
		if entry.Key == key {
			return entry.Value
		}
	}
	return ""
}

func buildObjectMetadataMap(objects []ObjectMeta) map[int]string {
	metadataMap := make(map[int]string)
	for _, obj := range objects {
		if name := extractMetadataValue(obj.Metadata, "name"); name != "" {
			metadataMap[obj.ID] = name
		}
	}
	return metadataMap
}

func isAssemblyObject(obj ObjectMeta) bool {
	objectName := extractMetadataValue(obj.Metadata, "name")
	return objectName == "Assembly" && len(obj.Parts) > 1
}

func getObjectNameFromParts(obj ObjectMeta) string {
	if len(obj.Parts) == 1 {
		if name := extractMetadataValue(obj.Parts[0].Metadata, "name"); name != "" {
			return name
		}
	}
	
	if name := extractMetadataValue(obj.Metadata, "name"); name != "" {
		return name
	}
	
	return fmt.Sprintf("Object_%d", obj.ID)
}

func getPartComponents(obj ObjectMeta) []ComponentInfo {
	var components []ComponentInfo
	
	for _, part := range obj.Parts {
		comp := ComponentInfo{
			ID:   part.ID,
			Name: extractMetadataValue(part.Metadata, "name"),
		}
		
		if comp.Name == "" {
			comp.Name = fmt.Sprintf("Part_%d", part.ID)
		}
		
		if sourceFile := extractMetadataValue(part.Metadata, "source_file"); sourceFile != "" {
			comp.SourceFile = sourceFile
		}
		
		components = append(components, comp)
	}
	
	return components
}

func parseModelInstances(instances []ModelInstance) map[int]int {
	objectToPlateMap := make(map[int]int)
	
	for _, instance := range instances {
		objectID := instance.ObjectID
		instanceID := instance.InstanceID
		
		// Если есть metadata, используем его для получения точных значений
		if len(instance.Metadata) > 0 {
			if objectIDStr := extractMetadataValue(instance.Metadata, "object_id"); objectIDStr != "" {
				if id, err := strconv.Atoi(objectIDStr); err == nil {
					objectID = id
				}
			}
			if instanceIDStr := extractMetadataValue(instance.Metadata, "instance_id"); instanceIDStr != "" {
				if id, err := strconv.Atoi(instanceIDStr); err == nil {
					instanceID = id
				}
			}
		}
		
		objectToPlateMap[objectID] = instanceID
	}
	
	return objectToPlateMap
}

func parsePlates(plates []Plate) (map[int]*PlateInfo, map[int]int) {
	plateMap := make(map[int]*PlateInfo)
	objectToPlateMap := make(map[int]int)
	
	for _, plate := range plates {
		plateID := plate.PlaterID
		plateName := plate.PlaterName
		
		// Извлекаем plater_id и plater_name из metadata
		if plateIDStr := extractMetadataValue(plate.Metadata, "plater_id"); plateIDStr != "" {
			if id, err := strconv.Atoi(plateIDStr); err == nil {
				plateID = id
			}
		}
		if nameStr := extractMetadataValue(plate.Metadata, "plater_name"); nameStr != "" {
			plateName = nameStr
		}
		
		plateMap[plateID] = &PlateInfo{
			PlateID:   plateID,
			PlateName: plateName,
			Objects:   []PlateObject{},
		}
		
		// Обрабатываем model_instance элементы для этого plate
		for _, instance := range plate.Instances {
			objectID := instance.ObjectID
			
			// Извлекаем object_id из metadata если есть
			if objectIDStr := extractMetadataValue(instance.Metadata, "object_id"); objectIDStr != "" {
				if id, err := strconv.Atoi(objectIDStr); err == nil {
					objectID = id
				}
			}
			
			objectToPlateMap[objectID] = plateID
		}
	}
	
	return plateMap, objectToPlateMap
}