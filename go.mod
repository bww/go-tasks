module github.com/bww/go-tasks

go 1.22.3

replace (
	github.com/bww/go-alert => ../go-alert
	github.com/bww/go-ident => ../go-ident
	github.com/bww/go-metrics => ../go-metrics
	github.com/bww/go-queue => ../go-queue
	github.com/bww/go-router => ../go-router
	github.com/bww/go-util => ../go-util
)

require (
	github.com/bww/go-alert v0.1.0
	github.com/bww/go-ident v0.1.0
	github.com/bww/go-metrics v0.1.0
	github.com/bww/go-queue v1.0.0
	github.com/bww/go-router v1.9.0
	github.com/bww/go-util v1.34.0
	github.com/dustin/go-humanize v1.0.1
	github.com/orcaman/concurrent-map/v2 v2.0.1
	github.com/stretchr/testify v1.9.0
	google.golang.org/grpc v1.63.2
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bww/go-acl v0.2.4 // indirect
	github.com/bww/go-auth v0.1.2 // indirect
	github.com/bww/go-xid v0.1.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/getsentry/sentry-go v0.28.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.5.0 // indirect
	github.com/prometheus/common v0.48.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240429193739-8cf5692501f6 // indirect
	google.golang.org/protobuf v1.34.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
