package lib

// Contains returns a boolean indicating whether a string is present in a passed in set of strings
func Contains(set map[string]struct{}, item string) bool {
	_, ok := set[item]
	return ok
}

// Unique returns a new copy of the passed in slice with all duplicates removed
func Unique(slice []string) (res []string) {
	set := make(map[string]struct{})

	for _, s := range slice {
		if Contains(set, s) {
			continue
		}
		res = append(res, s)
		set[s] = struct{}{}
	}

	return res
}
