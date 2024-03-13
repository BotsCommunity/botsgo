package converter

import "fmt"

func BooleanToInteger(property bool) int {
	if value := 1; property {
		return value
	}

	return 0
}

func SliceToString[element any](array []element) string {
	str, count := "", len(array)

	for index, field := range array {
		if str += fmt.Sprintf("%v", field); index != count-1 {
			str += ","
		}
	}

	return str
}
