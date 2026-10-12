package main

import "sort"

// maxListed is how many lines of a ranked list the diagnostics print.
const maxListed = 40

// rankedKeys returns the keys of hist ordered by descending count; ties are
// broken alphabetically so the output is stable.
func rankedKeys(hist map[string]int) []string {
	keys := make([]string, 0, len(hist))
	for k := range hist {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if hist[keys[i]] != hist[keys[j]] {
			return hist[keys[i]] > hist[keys[j]]
		}
		return keys[i] < keys[j]
	})
	return keys
}
