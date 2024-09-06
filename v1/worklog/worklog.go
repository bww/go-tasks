package worklog

import (
	"context"
	"errors"
	"time"

	"github.com/bww/go-ident/v1"
	siter "github.com/bww/go-iterator/v1"
)

var (
	ErrNotFound = errors.New("Not found")
	ErrConflict = errors.New("Sequence conflict")
)

type Criteria struct {
	Expired     bool      // Expired...
	Resolved    bool      // ... and Resolved are logically mutually exclusive
	IdleSince   time.Time // Excludes entries that HAVE BEEN updated after this time
	ActiveSince time.Time // Excludes entries that HAVE NOT BEEN updated since this time
	States      []State   // Only include results in these states; mutually exclusive with Expired and Resolved
}

type Worklog interface {
	CreateEntry(context.Context, *Entry) error
	StoreEntry(context.Context, *Entry) error
	RenewEntry(context.Context, *Entry, time.Time) (*Entry, error)

	FetchEntry(context.Context, ident.Ident, int64) (*Entry, error)
	FetchLatestEntryForTask(context.Context, ident.Ident) (*Entry, error)
	IterLatestEntryForEveryTask(context.Context, Criteria, time.Time) (siter.Iterator[*Entry], error)

	DeleteTask(context.Context, ident.Ident) error
}
