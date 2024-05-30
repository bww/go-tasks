package router

import (
	"strings"
)

func parseUTD(s string) (string, string, string) {
	const (
		authsep = "//"
		pathsep = "/"
	)
	var scheme, host, path string
	if x := strings.Index(s, ":"); x < 0 {
		return s, "", ""
	} else {
		scheme, s = s[:x], s[x+1:]
	}
	if s == wildcard {
		host = wildcard
	} else if strings.HasPrefix(s, authsep) {
		s = s[len(authsep):]
		if s != "" {
			if x := strings.Index(s, pathsep); x < 0 {
				host, s = s, ""
			} else {
				host, s = s[:x], s[x:]
			}
		}
	}
	if s != "" && !strings.HasPrefix(s, pathsep) {
		path = pathsep + s
	} else {
		path = s
	}
	return scheme, host, path
}
