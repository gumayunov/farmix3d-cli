package parser

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