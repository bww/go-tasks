package worklog

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Triggers maps a state to UTDs that should be executed when the
// controlling task reaches that state
type Triggers map[State][]string

func (t Triggers) Value() (driver.Value, error) {
	if len(t) < 1 {
		return nil, nil
	} else {
		return json.Marshal(t)
	}
}

func (t *Triggers) Scan(src interface{}) error {
	switch c := src.(type) {
	case nil:
		return nil
	case []byte:
		return json.Unmarshal(c, t)
	case string:
		return json.Unmarshal([]byte(c), t)
	default:
		return fmt.Errorf("Unsupported type: %T", src)
	}
}
