package service

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v2"

	"github.com/pkedy/golang-dapr/pkg/dapr"
	"github.com/pkedy/golang-dapr/pkg/features/gadgets"
)

type (
	Service struct {
		log   logr.Logger
		store gadgets.Store
	}
)

func New(log logr.Logger, store gadgets.Store) *Service {
	return &Service{
		log:   log,
		store: store,
	}
}

// SERVICE OPERATIONS

func (s *Service) RegisterService(app *fiber.App) {
	app.Get("/v1/gadgets/:id", func(c *fiber.Ctx) error {
		gadget, err := s.store.Load(c.Context(), c.Params("id"))
		return response(c, gadget, err)
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
						Match: `event.type == "gadget.v1"`,
						Path:  "/gadgets.v1",
					},
				},
			},
		},
	}
}

// HTTP

func (s *Service) RegisterEventHandlers(app *fiber.App) {
	app.Post("/gadgets.v1", s.SaveHTTP)
}

func (s *Service) SaveHTTP(c *fiber.Ctx) error {
	var gadget gadgets.Gadget
	if err := dapr.DecodeCloudEvent(c, nil, &gadget); err != nil {
		return err
	}
	if err := s.store.Save(c.Context(), &gadget); err != nil {
		return err
	}
	return c.SendString("OK")
}

// gRPC

func (s *Service) RegisterTopicEventHandlers(register dapr.RegisterEventHandler) {
	register("/gadgets.v1", s.SaveGRPC)
}

func (s *Service) SaveGRPC(ctx context.Context, in *pb.TopicEventRequest) (*pb.TopicEventResponse, error) {
	var gadget gadgets.Gadget
	if err := json.Unmarshal(in.Data, &gadget); err != nil {
		return nil, err
	}
	if err := s.store.Save(ctx, &gadget); err != nil {
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
		Match:      `event.type == "gadget.v1"`,
		Route:      "/gadgets.v1",
		Priority:   2,
	}, s.SaveSDK)
}

func (s *Service) SaveSDK(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	var gadget gadgets.Gadget
	if err := e.Struct(&gadget); err != nil {
		return false, err
	}
	// if err := dapr.DecodeTopicEvent(e, &gadget); err != nil {
	// 	return false, err
	// }
	if err := s.store.Save(ctx, &gadget); err != nil {
		return false, err
	}
	return false, nil
}
