package exec

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/bww/go-tasks/v1/worklog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type jsonError struct {
	Err error
}

func (e jsonError) Error() string {
	return e.Err.Error()
}

func (e jsonError) MarshalJSON() ([]byte, error) {
	msg, err := json.Marshal(e.Err.Error())
	if err != nil {
		return nil, err
	}
	return []byte(`{"message":` + string(msg) + `}`), nil
}

func stateForError(err error) worklog.State {
	if stat, ok := status.FromError(err); ok && stat.Code() == codes.Canceled {
		return worklog.Canceled
	} else if errors.Is(err, context.Canceled) {
		return worklog.Canceled
	} else {
		return worklog.Failed
	}
}
