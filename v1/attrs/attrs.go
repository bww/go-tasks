package attrs

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var ErrNotFound = errors.New("Not found")

type Attributes map[string]string

func (a Attributes) Int(k string) (int, error) {
	if len(a) == 0 {
		return 0, ErrNotFound
	}
	if v, ok := a[k]; ok {
		return strconv.Atoi(v)
	} else {
		return 0, ErrNotFound
	}
}

func (a Attributes) SetInt(k string, v int) {
	a[k] = strconv.Itoa(v)
}

func (a Attributes) Bool(k string) (bool, error) {
	if len(a) == 0 {
		return false, ErrNotFound
	}
	if v, ok := a[k]; ok {
		return strconv.ParseBool(v)
	} else {
		return false, ErrNotFound
	}
}

func (a Attributes) SetBool(k string, v bool) {
	a[k] = strconv.FormatBool(v)
}

func (a Attributes) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	} else {
		return json.Marshal(a)
	}
}

func (a *Attributes) Scan(src interface{}) error {
	var err error
	switch c := src.(type) {
	case nil:
		// nothing to do
	case []byte:
		err = json.Unmarshal(c, a)
	case string:
		err = json.Unmarshal([]byte(c), a)
	default:
		err = fmt.Errorf("Unsupported type: %T", src)
	}
	return err
}
