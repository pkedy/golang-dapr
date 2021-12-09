package service

import (
	"context"
	"encoding/json"

	pb "github.com/dapr/dapr/pkg/proto/runtime/v1"
	"github.com/dapr/go-sdk/service/common"
	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v2"

	"github.com/pkedy/golang-dapr/pkg/dapr"
	"github.com/pkedy/golang-dapr/pkg/features/products"
)

type (
	Service struct {
		log   logr.Logger
		store products.Store
	}
)

func New(log logr.Logger, store products.Store) *Service {
	return &Service{
		log:   log,
		store: store,
	}
}

// SERVICE OPERATIONS

func (s *Service) RegisterService(app *fiber.App) {
	app.Get("/v1/products/:id", func(c *fiber.Ctx) error {
		product, err := s.store.Load(c.Context(), c.Params("id"))
		return response(c, product, err)
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
				Default: "/products.v1",
			},
		},
	}
}

// HTTP

func (s *Service) RegisterEventHandlers(app *fiber.App) {
	app.Post("/products.v1", s.SaveHTTP)
}

func (s *Service) SaveHTTP(c *fiber.Ctx) error {
	var product products.Product
	if err := dapr.DecodeCloudEvent(c, nil, &product); err != nil {
		return err
	}
	if err := s.store.Save(c.Context(), &product); err != nil {
		return err
	}
	return c.SendString("OK")
}

// gRPC

func (s *Service) RegisterTopicEventHandlers(register dapr.RegisterEventHandler) {
	register("/products.v1", s.SaveGRPC)
}

func (s *Service) SaveGRPC(ctx context.Context, in *pb.TopicEventRequest) (*pb.TopicEventResponse, error) {
	var gadget products.Product
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
	// Default route
	return service.AddTopicEventHandler(&common.Subscription{
		PubsubName: "pubsub",
		Topic:      "inventory",
		Route:      "/products.v1",
	}, s.SaveSDK)
}

func (s *Service) SaveSDK(ctx context.Context, e *common.TopicEvent) (retry bool, err error) {
	var product products.Product
	if err := e.Struct(&product); err != nil {
		return false, err
	}
	// if err := dapr.DecodeTopicEvent(e, &product); err != nil {
	// 	return false, err
	// }
	if err := s.store.Save(ctx, &product); err != nil {
		return false, err
	}
	return false, nil
}
