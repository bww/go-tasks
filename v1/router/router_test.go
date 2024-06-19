package router

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/bww/go-tasks/v1"

	"github.com/bww/go-router/v1/path"
	"github.com/stretchr/testify/assert"
)

func TestUTDs(t *testing.T) {
	tests := []struct {
		utd                string
		scheme, host, path string
	}{
		{
			"foo://bar/zip/zap",
			"foo", "bar", "/zip/zap",
		},
		{
			"",
			"", "", "",
		},
		{
			"foo",
			"foo", "", "",
		},
		{
			"foo://",
			"foo", "", "",
		},
		{
			"foo://bar",
			"foo", "bar", "",
		},
		{
			"foo:/bar",
			"foo", "", "/bar",
		},
		{
			"foo:///",
			"foo", "", "/",
		},
		{
			"foo:bar",
			"foo", "", "/bar",
		},
		{
			"foo:bar/car",
			"foo", "", "/bar/car",
		},
		{
			"foo://bar/zip/{a}",
			"foo", "bar", "/zip/{a}",
		},
		{
			"foo://{zap}/zip/{a}",
			"foo", "{zap}", "/zip/{a}",
		},
	}
	for _, e := range tests {
		fmt.Println("-->", e.utd)
		s, h, p := parseUTD(e.utd)
		assert.Equal(t, e.scheme, s, e.utd)
		assert.Equal(t, e.host, h, e.utd)
		assert.Equal(t, e.path, p, e.utd)
	}
}

var errTestOk = errors.New("test OK")

func testRunTask(_ context.Context, req *tasks.Request, params tasks.Params) (tasks.Result, error) {
	return tasks.Result{}, errTestOk
}

func TestRoutes(t *testing.T) {
	rr := New()

	r1 := rr.Add("foo://bar/zip", tasks.TaskFunc(testRunTask))
	r2 := rr.Add("foo://bar/zip/{m}", tasks.TaskFunc(testRunTask))
	r3 := rr.Add("foo:/zip", tasks.TaskFunc(testRunTask))
	r4 := rr.Add("foo://{bop}/fop", tasks.TaskFunc(testRunTask))
	r5 := rr.Add("foo://{bop}/zip/{m}", tasks.TaskFunc(testRunTask))

	_ = rr.Add("foo://zzz/*", tasks.TaskFunc(testRunTask))
	_ = rr.Add("xxx://bar/*", tasks.TaskFunc(testRunTask))
	_ = rr.Add("yyy://*", tasks.TaskFunc(testRunTask))
	_ = rr.Add("zzz:*", tasks.TaskFunc(testRunTask))

	for _, e := range rr.Routes() {
		fmt.Println(">>>", e)
	}

	tests := []struct {
		utd   string
		route *Route
		vars  path.Vars
		err   error
	}{
		{
			"foo://bar/zip",
			r1, nil, nil,
		},
		{
			"foo://bar/zip/",
			r1, nil, nil,
		},
		{
			"foo://bar/zip/zap",
			r2, path.Vars{"m": "zap"}, nil,
		},
		{
			"foo://bar/zip/zap/zop",
			nil, nil, nil,
		},
		{
			"foo:///zip",
			r3, nil, nil,
		},
		{
			"foo:/zip",
			r3, nil, nil,
		},
		{
			"foo://zim/fop",
			r4, path.Vars{"bop": "zim"}, nil,
		},
		{
			"foo://zap/zip/zop",
			r5, path.Vars{"bop": "zap", "m": "zop"}, nil,
		},
		{
			"foo://anything-is-fine/zip/zop",
			r5, path.Vars{"bop": "anything-is-fine", "m": "zop"}, nil,
		},
	}

	for _, e := range tests {
		d, err := url.Parse(e.utd)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			x, v, err := rr.Find(d)
			if err != nil {
				assert.Equal(t, e.err, err, e.utd)
			} else {
				assert.Equal(t, e.route, x, e.utd)
				assert.Equal(t, e.vars, v, e.utd)
				if e.route != nil && assert.NotNil(t, x, e.utd) {
					_, err := x.Exec(context.Background(), tasks.NewRequest(d), tasks.Params{Vars: v})
					assert.Equal(t, errTestOk, err, e.utd)
				}
			}
		}
	}
}

func TestWildcards(t *testing.T) {
	rr := New()
	r1 := rr.Add("foo://bar/*", tasks.TaskFunc(testRunTask))
	r2 := rr.Add("foo://car/*", tasks.TaskFunc(testRunTask))
	r3 := rr.Add("bar:*", tasks.TaskFunc(testRunTask))

	for _, e := range rr.Routes() {
		fmt.Println(">>>", e)
	}

	tests := []struct {
		utd   string
		route *Route
		vars  path.Vars
		err   error
	}{
		{
			"foo://bar/zip",
			r1, nil, nil,
		},
		{
			"foo://car/zip/bar/jerkle",
			r2, nil, nil,
		},
		{
			"foo://bar/anything/will/match/this/wildcard/route/____",
			r1, nil, nil,
		},
		{
			"foo://car/____",
			r2, nil, nil,
		},
		{
			"zip://this/one/matches/anything/not/matched/by/a/preceding/route",
			nil, nil, nil,
		},
		{
			"bar://this/one/matches/anything/not/matched/by/a/preceding/route",
			r3, nil, nil,
		},
	}

	for _, e := range tests {
		d, err := url.Parse(e.utd)
		if assert.Nil(t, err, fmt.Sprint(err)) {
			x, v, err := rr.Find(d)
			if err != nil {
				assert.Equal(t, e.err, err, e.utd)
			} else {
				assert.Equal(t, e.route, x, e.utd)
				assert.Equal(t, e.vars, v, e.utd)
				if e.route != nil && assert.NotNil(t, x, e.utd) {
					_, err := x.Exec(context.Background(), tasks.NewRequest(d), tasks.Params{Vars: v})
					assert.Equal(t, errTestOk, err, e.utd)
				}
			}
		}
	}
}
