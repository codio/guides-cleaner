package utils

func ContainedInArray(array []string, value string) bool {
	for _, n := range array {
		if value == n {
			return true
		}
	}
	return false
}
