package worklog

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

var errInvalidState = fmt.Errorf("Invalid state")

type State string

const (
	Pending  = State("pending")
	Running  = State("running")
	Complete = State("complete")
	Canceled = State("canceled")
	Failed   = State("failed")
	Unknown  = State("unknown")
)

var ordinals = map[State]int{
	Pending:  0,
	Running:  1,
	Complete: 2,
	Canceled: 3,
	Failed:   4,
	Unknown:  -1,
}

var states = map[State]string{
	Pending:  "Pending",
	Running:  "Running",
	Complete: "Complete",
	Canceled: "Canceled",
	Failed:   "Failed",
	Unknown:  "Unknown",
}

func ParseState(s string) (State, error) {
	c := State(s)
	_, ok := states[c]
	if ok {
		return c, nil
	} else {
		return "", errInvalidState
	}
}

func (s State) Before(another State) bool {
	so, ok := ordinals[s]
	if !ok {
		so = -1
	}
	ao, ok := ordinals[another]
	if !ok {
		ao = -1
	}
	return so < ao
}

func (s State) After(another State) bool {
	so, ok := ordinals[s]
	if !ok {
		so = -1
	}
	ao, ok := ordinals[another]
	if !ok {
		ao = -1
	}
	return so > ao
}

func (s State) Resolved() bool {
	switch s {
	case Complete, Canceled, Failed:
		return true
	default:
		return false
	}
}

func (s State) Failure() bool {
	switch s {
	case Canceled, Failed:
		return true
	default:
		return false
	}
}

func (c State) String() string {
	return string(c)
}

func (c State) Name() string {
	n, ok := states[c]
	if ok {
		return n
	} else {
		return "Invalid"
	}
}

func (c State) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *State) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	v, err := ParseState(s)
	if err != nil {
		return err
	}
	*c = v
	return nil
}

func (c State) MarshalText() ([]byte, error) {
	return []byte(c.String()), nil
}

func (c *State) UnmarshalText(data []byte) error {
	v, err := ParseState(string(data))
	if err != nil {
		return err
	}
	*c = v
	return nil
}

func (c State) Value() (driver.Value, error) {
	return c.String(), nil
}

func (c *State) Scan(src interface{}) error {
	var err error
	var v State
	switch c := src.(type) {
	case []byte:
		v, err = ParseState(string(c))
	case string:
		v, err = ParseState(c)
	default:
		err = fmt.Errorf("Unsupported type: %T", src)
	}
	if err != nil {
		return err
	}
	*c = v
	return nil
}
