package attrs

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Attributes map[string]string

func (r Attributes) Value() (driver.Value, error) {
	if r == nil {
		return nil, nil
	} else {
		return json.Marshal(r)
	}
}

func (r *Attributes) Scan(src interface{}) error {
	var err error
	switch c := src.(type) {
	case nil:
		// nothing to do
	case []byte:
		err = json.Unmarshal(c, r)
	case string:
		err = json.Unmarshal([]byte(c), r)
	default:
		err = fmt.Errorf("Unsupported type: %T", src)
	}
	return err
}
