package service

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/bww/go-acl/v1"
	"github.com/bww/go-auth/v1/jwt"
	"github.com/bww/go-auth/v1/middle"
	"github.com/bww/go-rest/v2"
	resterrs "github.com/bww/go-rest/v2/errors"
	"github.com/bww/go-rest/v2/httputil"
	"github.com/bww/go-rest/v2/response"
	"github.com/bww/go-router/v2"
	"github.com/bww/go-tasks/v1"
	"github.com/bww/go-tasks/v1/exec"
	"github.com/bww/go-tasks/v1/transport"
	"github.com/bww/go-util/v1/urls"
	"github.com/bww/go-validate/v1"
	"github.com/dustin/go-humanize"
)

const (
	ControlRealm = "control"
	DataRealm    = "data"

	ExecResource  = "exec"
	QueueResource = "queue"
)

func scope(r string, a ...acl.Action) acl.Scopes {
	return acl.Scopes{acl.NewScope(r, a...)}
}

type Service struct {
	*rest.Service
	addr  string
	queue *tasks.Queue
	exec  *exec.Executor
	log   *slog.Logger
}

func NewWithConfig(conf Config) (*Service, error) {
	r, err := rest.New(
		rest.WithVerbose(conf.Verbose),
		rest.WithDebug(conf.Debug),
		rest.WithMetrics(conf.Metrics),
	)
	if err != nil {
		return nil, err
	}

	jwtacl := jwt.New(conf.Secret)
	dataRealm := acl.Realm{{Type: DataRealm}}

	s := &Service{
		Service: r,
		addr:    conf.Addr,
		queue:   conf.Queue,
		exec:    conf.Exec,
		log:     slog.With("service", "tasks"),
	}

	// Obtain the status of this service; this can be used as a health check
	r.Add(urls.Join(conf.Prefix, "/status"), s.handleStatus).Methods("GET")
	// Submit a task to the queue so it can be scheduled for normal execution; this is the way work is normally submitted to the service
	r.Add(urls.Join(conf.Prefix, "/v1/queue"), s.handleWriteQueue).Methods("POST").Use(middle.ACL(jwtacl, dataRealm, scope(QueueResource, acl.Write)))
	// Submit a task DIRECTLY to the local executor and wait for it to finish SYNCHRONOUSLY; this is really only intended for testing scenarios
	r.Add(urls.Join(conf.Prefix, "/v1/tasks"), s.handleExecTask).Methods("POST").Use(middle.ACL(jwtacl, dataRealm, scope(ExecResource, acl.Write)))

	return s, nil
}

func (s *Service) Run(cxt context.Context) error {
	server := &http.Server{
		Addr:           s.addr,
		Handler:        s,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		<-cxt.Done()                          // wait for the context to end, if it ever does...
		server.Shutdown(context.Background()) // ...and shutdown the service
	}()

	return server.ListenAndServe()
}

func (s *Service) handleStatus(req *router.Request, cxt router.Context) (*router.Response, error) {
	return response.JSON(struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	}), nil
}

func (s *Service) handleWriteQueue(req *router.Request, cxt router.Context) (*router.Response, error) {
	if s.queue == nil {
		return nil, resterrs.Errorf(http.StatusServiceUnavailable, "Task queue is not available")
	}

	conf, err := tasks.PublishConfigFromParams(req.URL.Query())
	if err != nil {
		return nil, resterrs.Errorf(http.StatusBadRequest, "Invalid parameters").SetCause(err)
	}

	var msg *transport.Message
	err = httputil.Unmarshal(req, &msg)
	if err != nil {
		return nil, resterrs.Errorf(http.StatusBadRequest, "Could not unmarshal entity").SetCause(err)
	}
	errs := validate.New().Validate(msg)
	if len(errs) > 0 {
		return nil, resterrs.Errorf(http.StatusBadRequest, "Invalid entity").SetFieldErrors(errs)
	}

	s.log.With("utd", msg.UTD, "size", humanize.Bytes(uint64(len(msg.Data)))).Info("Publish task")
	err = s.queue.Publish(req.Context(), msg, tasks.UseConfig(conf))
	if err != nil {
		return nil, resterrs.Errorf(http.StatusBadGateway, "Could not publish task").SetCause(err)
	}

	return response.JSON(msg), nil
}

func (s *Service) handleExecTask(req *router.Request, cxt router.Context) (*router.Response, error) {
	if s.exec == nil {
		return nil, resterrs.Errorf(http.StatusServiceUnavailable, "Task executor is not available")
	}

	var msg *transport.Message
	err := httputil.Unmarshal(req, &msg)
	if err != nil {
		return nil, resterrs.Errorf(http.StatusBadRequest, "Could not unmarshal entity").SetCause(err)
	}
	errs := validate.New().Validate(msg)
	if len(errs) > 0 {
		return nil, resterrs.Errorf(http.StatusBadRequest, "Invalid entity").SetFieldErrors(errs)
	}

	s.log.With("utd", msg.UTD, "size", humanize.Bytes(uint64(len(msg.Data)))).Info("Execute task (synchronous)")
	res, err := s.exec.Proc(req.Context(), msg, nil)
	if err != nil {
		return nil, resterrs.Errorf(http.StatusBadGateway, "%s", err.Error()).SetCause(err)
	}

	return response.JSON(res), nil
}
