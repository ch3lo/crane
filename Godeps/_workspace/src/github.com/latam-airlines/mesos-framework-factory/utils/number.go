package utils

/* Returns a Natural Number or -1 if not in the map */
func ExtractNaturalNumber(params map[string]interface{}, key string) int {
	number, ok := params[key].(int)

	if ok {
		return number
	} else {
		return -1
	}
}
