package router

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/bww/go-tasks/v1"

	"github.com/bww/go-router/v1/path"
)

const (
	wildcard      = "*"
	slashWildcard = "/*"
)

func mergeVars(a, b map[string]string) map[string]string {
	if a == nil {
		return b
	} else if b == nil {
		return a
	}
	for k, v := range b {
		a[k] = v
	}
	return a
}

type matchState struct {
	// future use
}

// An individual route
type Route struct {
	task   tasks.Task
	scheme string
	host   string
	paths  []path.Path
}

// Add paths
func (r *Route) Paths(s ...string) *Route {
	p := make([]path.Path, len(s))
	for i, e := range s {
		p[i] = path.Parse(e)
	}
	r.paths = append(r.paths, p...)
	return r
}

// Matches or not
func (r Route) Matches(utd *url.URL, state *matchState) (bool, map[string]string) {
	if !strings.EqualFold(r.scheme, utd.Scheme) {
		return false, nil
	}

	var gvars map[string]string
	if r.host == wildcard {
		return true, nil
	} else if l := len(r.host); l > 2 && r.host[0] == '{' && r.host[l-1] == '}' {
		gvars = map[string]string{strings.TrimSpace(string(r.host[1 : l-1])): utd.Host} // matches everything
	} else if !strings.EqualFold(r.host, utd.Host) {
		return false, nil
	}

	if l := len(r.paths); l == 0 {
		return true, gvars // no paths to match, we must succeed
	}
	for _, e := range r.paths {
		if e.String() == wildcard {
			return true, nil
		} else if ok, pvars := e.Matches(utd.Path); ok {
			return true, mergeVars(gvars, pvars)
		}
	}

	return false, nil
}

// Handle the request
func (r *Route) Exec(cxt context.Context, req *tasks.Request, params tasks.Params) (tasks.Result, error) {
	return r.task.Exec(cxt, req, params)
}

// Describe this route
func (r *Route) String() string {
	b := strings.Builder{}
	b.WriteString(r.scheme + ":")
	if r.host != "" {
		b.WriteString("//" + r.host)
	}
	if len(r.paths) == 0 {
		if r.host != wildcard {
			b.WriteString("/*")
		}
	} else if len(r.paths) == 1 {
		b.WriteString(r.paths[0].String())
	} else {
		b.WriteString("{")
		for i, e := range r.paths {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(e.String())
		}
		b.WriteString("}")
	}
	return b.String()
}

// Router
type Router interface {
	Add(string, tasks.Task) *Route
	Find(*url.URL) (*Route, path.Vars, error)
	Exec(context.Context, *tasks.Request) (tasks.Result, error)
	Routes() []*Route
}

type router struct {
	routes []*Route
}

func New() Router {
	return &router{}
}

// Obtain a copy of all the routes managed by this router
func (r *router) Routes() []*Route {
	routes := make([]*Route, len(r.routes))
	copy(routes, r.routes)
	return routes
}

// Add a route
func (r *router) Add(d string, t tasks.Task) *Route {
	s, h, p := parseUTD(d)

	var c []path.Path
	if p == wildcard || p == slashWildcard {
		c = []path.Path{} // special handling for '/*' case
	} else {
		c = []path.Path{path.Parse(p)}
	}

	v := &Route{t, s, h, c}
	r.routes = append(r.routes, v)
	return v
}

// Find a route for the request, if we have one
func (r router) Find(utd *url.URL) (*Route, path.Vars, error) {
	state := &matchState{}
	for _, e := range r.routes {
		m, vars := e.Matches(utd, state)
		if m {
			return e, vars, nil
		}
	}
	return nil, nil, nil
}

// Exec a task for the provided UTD
func (r router) Exec(cxt context.Context, req *tasks.Request) (tasks.Result, error) {
	var res tasks.Result
	if req.UTD == nil {
		return res, tasks.ErrInvalidRequest
	}
	match, vars, err := r.Find(req.UTD)
	if err != nil {
		return res, err
	} else if match == nil {
		return res, fmt.Errorf("%w: %v", tasks.ErrUnsupported, req.UTD)
	}
	if vars == nil {
		vars = make(path.Vars)
	}
	return match.Exec(cxt, req, tasks.Params{
		Vars: vars,
	})
}
