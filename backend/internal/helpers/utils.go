package helpers

// Any returns true if any element in the slice satisfies the predicate.
func Any(slice []string, test func(string) bool) bool {
	for _, v := range slice {
		if test(v) {
			return true
		}
	}
	return false
}

// All returns true if all elements in the slice satisfy the predicate.
func All(slice []string, test func(string) bool) bool {
	for _, v := range slice {
		if !test(v) {
			return false
		}
	}
	return true
}

// Filter returns a new slice containing only the elements that satisfy the predicate.
func Filter(slice []string, test func(string) bool) []string {
	var result []string
	for _, v := range slice {
		if test(v) {
			result = append(result, v)
		}
	}
	return result
}
