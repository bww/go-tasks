package transport

import (
	"errors"
	"strings"
)

var errInvalidType = errors.New("Invalid type")

type Type string

const (
	Managed = Type("managed") // task is managed; we log work, suppress duplicates, retry failures, etc
	Oneshot = Type("oneshot") // task is executed on a best-effort basis; work is not logged
	Invalid = Type("")
)

var types = []Type{
	Managed,
	Oneshot,
}

func ParseType(t string) (Type, error) {
	for _, e := range types {
		if strings.EqualFold(t, string(e)) {
			return e, nil
		}
	}
	return Invalid, errInvalidType
}

func (t Type) String() string {
	return string(t)
}
