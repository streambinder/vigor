package util

import "slices"

func Dedupe(s []string) []string {
	slices.Sort(s)
	return slices.Compact(s)
}
