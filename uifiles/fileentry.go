package uifiles

import "strings"

type FileEntryList []string

func fileEntriesWithType(l FileEntryList, t string) FileEntryList {
	ret := make(FileEntryList, 0, len(l))
	for _, e := range l {
		if strings.HasPrefix(e, t+":") {
			ret = append(ret, e)
		}
	}
	return ret
}
