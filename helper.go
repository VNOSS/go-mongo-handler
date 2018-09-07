package db

func cloneStringMap(source map[string]interface{}) map[string]interface{} {
	resultMap := make(map[string]interface{})
	for key, value := range source {
		resultMap[key] = value
	}
	return resultMap
}
