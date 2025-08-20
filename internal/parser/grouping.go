package parser

func GroupObjectsByName(objects []PlateObject) map[string]GroupedObject {
	groups := make(map[string]GroupedObject)

	for _, obj := range objects {
		key := obj.Name + "|" + obj.Type // Группируем по имени и типу
		
		if existing, exists := groups[key]; exists {
			// Увеличиваем счетчик и добавляем ID
			existing.Count++
			existing.ObjectIDs = append(existing.ObjectIDs, obj.ID)
			groups[key] = existing
		} else {
			// Создаем новую группу
			groups[key] = GroupedObject{
				Name:       obj.Name,
				Type:       obj.Type,
				Count:      1,
				Components: obj.Components,
				ObjectIDs:  []int{obj.ID},
			}
		}
	}

	return groups
}