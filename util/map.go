package util

func FilterMapKey(data map[string]interface{}, canInclude []string) {
	for key := range data {
		valid := false
		for _, v := range canInclude {
			if key == v {
				valid = true
				break
			}
		}
		if !valid {
			delete(data, key)
		}
	}
}
