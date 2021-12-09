package service

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v2"

	"github.com/pkedy/golang-dapr/pkg/dapr"
	"github.com/pkedy/golang-dapr/pkg/features/widgets"
)

type (
	Service struct {
		log   logr.Logger
		store widgets.Store
	}
)

func New(log logr.Logger, store widgets.Store) *Service {
	return &Service{
		log:   log,
		store: store,
	}
}

// SERVICE OPERATIONS

func (s *Service) RegisterService(app *fiber.App) {
	app.Get("/v1/widgets/:id", func(c *fiber.Ctx) error {
		widget, err := s.store.Load(c.Context(), c.Params("id"))
		return response(c, widget, err)
	})
}

func response(c *fiber.Ctx, val interface{}, err error) error {
	if err != nil {
		return err
	}
	//return c.Format(val)
	return c.JSON(val)
}

// EVENT HANDLERS

func (s *Service) Subscriptions() []dapr.Subscription {
	return []dapr.Subscription{
		{
			PubsubName: "pubsub",
			Topic:      "inventory",
			Routes: dapr.Routes{
				Rules: []dapr.Rule{
					{
						Match: `event.type == "widget.v1"`,
						Path:  "/widgets.v1",
					},
				},
			},
		},
	}
}

// HTTP

func (s *Service) RegisterEventHandlers(app *fiber.App) {
	app.Post("/widgets.v1", s.SaveHTTP)
}

func (s *Service) SaveHTTP(c *fiber.Ctx) error {
	var widget widgets.Widget
	if err := dapr.DecodeCloudEvent(c, nil, &widget); err != nil {
		return err
	}
	if err := s.store.Save(c.Context(), &widget); err != nil {
		return err
	}
	return c.SendString("OK")
}

func (s *Service) RegisterTopicEventHandlers(register dapr.RegisterEventHandler) {
	register("/widgets.v1", s.SaveGRPC)
}

// gRPC

func (s *Service) SaveGRPC(ctx context.Context, in *pb.TopicEventRequest) (*pb.TopicEventResponse, error) {
	var widget widgets.Widget
	if err := json.Unmarshal(in.Data, &widget); err != nil {
		return nil, err
	}
	if err := s.store.Save(ctx, &widget); err != nil {
		return nil, err
	}

	return &pb.TopicEventResponse{
		Status: pb.TopicEventResponse_SUCCESS,
	}, nil
}

// SDK

func (s *Service) RegisterTopicEventHandlersSDK(service common.Service) error {
	return service.AddTopicEventHandler(&common.Subscription{
		PubsubName: "pubsub",
		Topic:      "inventory",
		Match:      `event.type == "widget.v1"`,
		Route:      "/widgets.v1",
		Priority:   1,
	}, s.SaveSDK)
}

func (s *Service) SaveSDK(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	var widget widgets.Widget
	if err := e.Struct(&widget); err != nil {
		return false, err
	}
	if err := s.store.Save(ctx, &widget); err != nil {
		return false, err
	}
	return false, nil
}
