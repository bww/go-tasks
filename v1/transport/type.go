package transport

import (
	"errors"
	"fmt"
	"strings"
)

var errInvalidType = errors.New("Invalid type")

type Type string

const (
	Managed = Type("managed") // task is managed; we log work, suppress duplicates, retry failures, etc
	Oneshot = Type("oneshot") // task is executed on a best-effort basis; work is not logged
	Cronjob = Type("cronjob") // a managed task which isn't fully realized because it is initialized by a cron service
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
	return "", fmt.Errorf("%w: %s", errInvalidType, t)
}

func (t Type) MarshalText() ([]byte, error) {
	// zero value marshals to one-shot, as a special case
	if t == "" {
		return []byte(string(Oneshot)), nil
	} else {
		return []byte(t.String()), nil
	}
}

func (t *Type) UnmarshalText(text []byte) error {
	// zero value unmarhsals to one-shot, as a special case
	if len(text) == 0 {
		*t = Oneshot
		return nil
	}
	// otherwise, unmarshal the input as normal
	v, err := ParseType(string(text))
	if err != nil {
		return err
	}
	*t = v
	return nil
}

func (t Type) String() string {
	return string(t)
}
