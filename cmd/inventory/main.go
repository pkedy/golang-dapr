package main

import (
	"context"
	"errors"
	"flag"
	"net"
	"os"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	dapr_server_grpc "github.com/dapr/go-sdk/service/grpc"
	dapr_server_http "github.com/dapr/go-sdk/service/http"
	"github.com/go-logr/zapr"
	"github.com/gofiber/fiber/v2"
	"github.com/oklog/run"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pkedy/golang-dapr/pkg/components/secrets"
	"github.com/pkedy/golang-dapr/pkg/components/state"
	"github.com/pkedy/golang-dapr/pkg/connect/postgres"
	"github.com/pkedy/golang-dapr/pkg/dapr"
	"github.com/pkedy/golang-dapr/pkg/errorz"
	gadgets_repo "github.com/pkedy/golang-dapr/pkg/features/gadgets/repository"
	gadgets_service "github.com/pkedy/golang-dapr/pkg/features/gadgets/service"
	products_repo "github.com/pkedy/golang-dapr/pkg/features/products/repository"
	products_service "github.com/pkedy/golang-dapr/pkg/features/products/service"
	widgets_repo "github.com/pkedy/golang-dapr/pkg/features/widgets/repository"
	widgets_service "github.com/pkedy/golang-dapr/pkg/features/widgets/service"
)

// api is an interface to embed all the components.
type api interface {
	secrets.Store
	state.Store
	Name() string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize logger
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log := zapr.NewLogger(zapLog)

	clientType := "sdk"

	flag.Parse()
	args := flag.Args()
	if len(args) > 0 {
		clientType = args[0]
	}

	////////////////////////////////////////////////////////
	// For example purposes only, this application can
	// connect to Dapr using:
	//
	//   * Custom code for HTTP
	//   * Custom code for gRPC
	//   * Using the Go SDK (protocol doesn't matter)
	//

	var daprClient api
	switch clientType {
	case "http":
		daprClient, err = dapr.NewHTTP(ctx)
	case "grpc":
		daprClient, err = dapr.NewGRPC(ctx)
	default:
		daprClient, err = dapr.NewSDK(ctx)
	}
	if err != nil {
		log.Error(err, "could not create connection to Dapr")
		os.Exit(1)
	}
	log.Info("Client initialized", "name", daprClient.Name())

	// Connect to database
	pool, err := postgres.Connect(ctx, daprClient,
		"secrets", "postgres",
		widgets_repo.AfterConnect)
	if err != nil {
		log.Error(err, "could not create connection to Postgres")
		os.Exit(1)
	}
	defer pool.Close()

	// Wire up dependencies

	// Uses Postgres database
	widgetRepo := widgets_repo.New(log, pool)
	widgetRest := widgets_service.New(log, widgetRepo)

	// Uses state store
	gadgetRepo := gadgets_repo.New(log, daprClient, "statestore")
	gadgetRest := gadgets_service.New(log, gadgetRepo)

	// Uses service invocation
	productRepo, err := products_repo.New(log)
	if err != nil {
		log.Error(err, "could not create connection to Products service")
		os.Exit(1)
	}
	defer productRepo.Close()
	productRest := products_service.New(log, productRepo)

	// Fiber app config with custom error handler
	config := fiber.Config{
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			errz := errorz.From(err)
			return c.Status(errz.Code).JSON(errz)
		},
	}

	var g run.Group
	// Public REST API operations
	{
		app := fiber.New(config)
		dapr.RegisterServices(app,
			widgetRest, gadgetRest, productRest)
		g.Add(func() error {
			return app.Listen(":3000")
		}, func(err error) {
			app.Shutdown()
		})
	}

	////////////////////////////////////////////////////////
	// It is recommended to listen on a different port
	// for private Dapr callbacks. Below are listeners for:
	//
	//   * Custom code for HTTP
	//   * Custom code for gRPC
	//   * Using the SDK for HTTP
	//   * Using the SDK for gRPC
	//

	////////////////////////////////////////////////////////
	// Each of the feature packages will add their own
	// subscriptions. The helper code in
	// /pkg/dapr/subscriptions.go merges them into a single
	// payload response to the Dapr sidecar.
	//
	// It will look like this:
	//
	// subscriptions:
	//   - pubsubname: pubsub
	//     topic: inventory
	//     routes:
	//       rules:
	//         - match: "event.type == 'widget.v1'"
	//           path: /widgets
	//         - match: "event.type == 'gadget.v1'"
	//           path: /gadgets
	//       default: /products
	//

	// Custom - HTTP events handlers
	{
		app := fiber.New(config)
		dapr.RegisterEventHandlers(app,
			widgetRest, gadgetRest, productRest)
		dapr.Subscribe(log, dapr.SubscribeHTTPHandler(log, app),
			widgetRest, gadgetRest, productRest)
		g.Add(func() error {
			return app.Listen(":3001")
		}, func(err error) {
			app.Shutdown()
		})
	}
	// Custom - gRPC event handlers
	{
		gs := grpc.NewServer()
		server := dapr.NewServer(log)
		server.RegisterTopicEventHandlers(
			widgetRest, gadgetRest, productRest)
		dapr.Subscribe(log, server.Subscribe,
			widgetRest, gadgetRest, productRest)
		pb.RegisterAppCallbackServer(gs, server)
		g.Add(func() error {
			ln, err := net.Listen("tcp", ":4001")
			if err != nil {
				return err
			}
			return gs.Serve(ln)
		}, func(err error) {
			gs.GracefulStop()
		})
	}
	// Using SDK - HTTP events handlers
	{
		var s common.Service
		g.Add(func() error {
			s = dapr_server_http.NewService(":3002")
			err = multierr.Combine(
				widgetRest.RegisterTopicEventHandlersSDK(s),
				gadgetRest.RegisterTopicEventHandlersSDK(s),
				productRest.RegisterTopicEventHandlersSDK(s))
			if err != nil {
				return err
			}
			return s.Start()
		}, func(err error) {
			if s != nil {
				s.Stop()
			}
		})
	}
	// Using SDK - gRPC events handlers
	{
		var s common.Service
		g.Add(func() (err error) {
			s, err = dapr_server_grpc.NewService(":4002")
			if err != nil {
				return err
			}
			err = multierr.Combine(
				widgetRest.RegisterTopicEventHandlersSDK(s),
				gadgetRest.RegisterTopicEventHandlersSDK(s),
				productRest.RegisterTopicEventHandlersSDK(s))
			if err != nil {
				return err
			}
			return s.Start()
		}, func(err error) {
			if s != nil {
				s.Stop()
			}
		})
	}
	// Termination signals
	{
		g.Add(run.SignalHandler(ctx, os.Interrupt, os.Kill))
	}

	var se run.SignalError
	if err := g.Run(); err != nil && !errors.As(err, &se) {
		log.Error(err, "goroutine error")
		os.Exit(1)
	}
}
