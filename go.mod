module github.com/pkedy/golang-dapr

go 1.17

require (
	github.com/dapr/dapr v1.5.1
	github.com/dapr/go-sdk v1.3.0
	github.com/go-logr/logr v1.2.0
	github.com/go-logr/zapr v1.2.0
	github.com/gofiber/fiber/v2 v2.22.0
	github.com/golang/protobuf v1.5.2
	github.com/jackc/pgx/v4 v4.14.0
	github.com/oklog/run v1.1.0
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.19.1
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
)

require (
	github.com/andybalholm/brotli v1.0.2 // indirect
	github.com/gofrs/uuid v4.1.0+incompatible // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.10.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.2.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.9.0 // indirect
	github.com/jackc/puddle v1.2.0 // indirect
	github.com/klauspost/compress v1.13.4 // indirect
	github.com/lib/pq v1.10.4 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shopspring/decimal v1.3.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.31.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	go.uber.org/atomic v1.8.0 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/genproto v0.0.0-20210524171403-669157292da3 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

// Will remove once https://github.com/dapr/go-sdk/pull/227 is merged
replace github.com/dapr/go-sdk => github.com/pkedy/go-sdk v1.2.1-0.20211209131922-5fd24998e2ee
