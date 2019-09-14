package maps

import "sort"

func GetKeys(mapObj map[string]string) []string {
	keys := make([]string, 0, len(mapObj))
	for k, _ := range mapObj {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func DeepCopy(oldMap map[string]string) map[string]string {
	newMap := make(map[string]string, len(oldMap))

	for k, v := range oldMap {
		newMap[k] = v
	}

	return newMap
}
