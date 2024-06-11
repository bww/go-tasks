package utds

import (
	"crypto/sha1"
	"encoding/base32"
	"net/url"
	"path"
	"strings"
)

func hash(k string) string {
	v := sha1.Sum([]byte(k))
	return base32.StdEncoding.EncodeToString(v[:])
}

// Identity strips off query params from a UTD to yield the identity portion of
// the URI.Query parameters are not considered to be part of the identify of
// a UTD.
func Identity(utd string) string {
	if x := strings.Index(utd, "?"); x >= 0 {
		utd = utd[:x]
	}
	return utd
}

// Key produces a hashed key from the identity of the provdied UTD. It is the
// equivalent of the following:
//
//	md5(Identity(utd))
func Key(utd string) string {
	if x := strings.Index(utd, "?"); x >= 0 {
		utd = utd[:x]
	}
	return hash(utd)
}

// ParsePath parses a UTD and splits its path into components, returning an
// array
func ParsePath(utd string) ([]string, error) {
	u, err := url.Parse(utd)
	if err != nil {
		return nil, err
	}
	return SplitPath(u.Path), nil
}

// SplitPath splits a UTD path into components, returning an array.
func SplitPath(s string) []string {
	var p []string

	for s != "" && s != "/" {
		var c string
		s, c = path.Split(s)
		p = append(p, c)
		if s != "" {
			s = s[:len(s)-1]
		}
	}

	l := len(p)
	for i := 0; i < l/2; i++ {
		p[i], p[l-i-1] = p[l-i-1], p[i]
	}

	return p
}
