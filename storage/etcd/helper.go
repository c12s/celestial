package etcd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	Tombstone = "!!Tombstone"
)

func firstN(text string, n int) string {
	return text[:n]
}

func lastN(text string, n int) string {
	return text[len(text)-n:]
}

func merge(text string, n int) string {
	return strings.Join([]string{firstN(text, n), "....", lastN(text, n)}, "")
}

func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

func tolabels(ls map[string]string) string {
	lbs := []string{}
	for k, v := range ls {
		if k != "ID" && k != "NAME" && k != "RETENTION" && k != "NEW_ID" {
			fmt.Println("{{PAIR}}", k, v)
			lbs = append(lbs, strings.Join([]string{k, v}, ":"))
		}
	}
	sort.Strings(lbs)
	return strings.Join(lbs, ",")
}

func index(nodes []string, userid, namespace string) []string {
	ret := []string{}
	for _, n := range nodes {
		// ret = append(ret, strings.Split(n, ".")[2])
		n = strings.ReplaceAll(n, ".", "/")
		prefix := strings.Join([]string{
			"topology", "regions", n,
		}, "/")

		sufix := strings.Join([]string{
			userid, namespace, "topology",
		}, ":")

		ret = append(ret, strings.Join([]string{
			prefix, sufix,
		}, "/"))
	}

	return ret
}

func newId(id string) string {
	return strings.Join([]string{
		"topology",
		"regions",
		id,
	}, ".")
}

func toString(n int64) string {
	return strconv.FormatInt(n, 10)
}
