package dapr

import (
	"encoding/json"

	"github.com/go-logr/logr"
	"github.com/gofiber/fiber/v2"
)

type (
	Service interface {
		RegisterService(app *fiber.App)
	}

	Events interface {
		RegisterEventHandlers(app *fiber.App)
	}
)

func RegisterServices(app *fiber.App, services ...Service) {
	for _, s := range services {
		s.RegisterService(app)
	}
}

func RegisterEventHandlers(app *fiber.App, events ...Events) {
	for _, e := range events {
		e.RegisterEventHandlers(app)
	}
}

func SubscribeHTTPHandler(log logr.Logger, app *fiber.App) func(subscriptions []*Subscription) {
	return func(subscriptions []*Subscription) {
		app.Get("/dapr/subscribe", func(c *fiber.Ctx) error {
			log.Info("subscribe called", "subscriptions", subscriptions)
			return c.JSON(subscriptions)
		})
	}
}

func DecodeCloudEvent(c *fiber.Ctx, ce *CloudEvent, target interface{}) error {
	var event CloudEvent
	if err := c.BodyParser(&event); err != nil {
		return err
	}
	if ce != nil {
		*ce = event
	}
	return json.Unmarshal(event.Data, target)
}
