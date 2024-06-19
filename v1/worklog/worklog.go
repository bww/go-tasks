package worklog

import (
	"context"
	"errors"
	"time"

	"github.com/bww/go-ident/v1"
)

var ErrNotFound = errors.New("Not found")

type Worklog interface {
	CreateEntry(context.Context, *Entry) error
	StoreEntry(context.Context, *Entry) error
	RenewEntry(context.Context, *Entry, time.Time) (*Entry, error)
	FetchEntry(context.Context, ident.Ident, int64) (*Entry, error)
	FetchLatestEntryForTask(context.Context, ident.Ident) (*Entry, error)
}
