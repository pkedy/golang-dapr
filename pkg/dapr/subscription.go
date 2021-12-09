package dapr

import (
	"errors"

	"github.com/go-logr/logr"
)

// Dapr subscription response
type (
	Subscription struct {
		PubsubName string            `json:"pubsubname"`
		Topic      string            `json:"topic"`
		Metadata   map[string]string `json:"metadata,omitempty"`
		Routes     Routes            `json:"routes"`
	}

	Routes struct {
		Rules   []Rule `json:"rules,omitempty"`
		Default string `json:"default,omitempty"`
	}

	Rule struct {
		Match string `json:"match"`
		Path  string `json:"path"`
	}

	Subscriber interface {
		Subscriptions() []Subscription
	}
)

var ErrDuplicateDefaultRoute = errors.New("duplicate default route")

// Subscribe will gather all the subscriptions from `subscribers`,
// merge them, and pass them to `register`.
func Subscribe(log logr.Logger, register func(subscriptions []*Subscription), subscribers ...Subscriber) {
	subscriptions := make([]*Subscription, 0, 10)
	subscriptionMap := make(map[string]*Subscription)

	// Merge subscriptions from all subscribers.
	for _, subscriber := range subscribers {
		subs := subscriber.Subscriptions()
		for _, s := range subs {
			key := s.PubsubName + ":" + s.Topic
			sub, ok := subscriptionMap[key]
			if !ok {
				sub = &Subscription{
					PubsubName: s.PubsubName,
					Topic:      s.Topic,
				}
				subscriptionMap[key] = sub
				subscriptions = append(subscriptions, sub)
			}
			sub.Routes.Rules = append(sub.Routes.Rules, s.Routes.Rules...)
			if s.Routes.Default != "" {
				if sub.Routes.Default != "" {
					log.Error(ErrDuplicateDefaultRoute,
						"Default route already exists for subscription",
						"pubsub", sub.PubsubName, "topic", sub.Topic)
				}
				sub.Routes.Default = s.Routes.Default
			}
		}
	}

	register(subscriptions)
}
