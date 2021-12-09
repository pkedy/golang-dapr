package dapr

import (
	"context"
	"encoding/json"

	dapr "github.com/dapr/go-sdk/client"

	"github.com/pkedy/golang-dapr/pkg/components/secrets"
	"github.com/pkedy/golang-dapr/pkg/components/state"
	"github.com/pkedy/golang-dapr/pkg/errorz"
)

type Client struct {
	client dapr.Client
}

var (
	_ = state.Store((*Client)(nil))
	_ = secrets.Store((*Client)(nil))
)

func NewSDK(ctx context.Context) (*Client, error) {
	client, err := dapr.NewClient()
	if err != nil {
		return nil, err
	}
	return &Client{
		client: client,
	}, nil
}

func (c *Client) Name() string {
	return "Go SDK"
}

func (c *Client) SetState(ctx context.Context, store string, items ...state.Item) error {
	stateItems := make([]*dapr.SetStateItem, len(items))
	for i := range items {
		item := items[i]
		data, err := json.Marshal(item.Value)
		if err != nil {
			return errorz.Internal(err, "could not serialize value for key %q", item.Key)
		}
		stateItems[i] = &dapr.SetStateItem{
			Key:   item.Key,
			Etag:  etag(item.ETag),
			Value: data,
		}
	}
	if err := c.client.SaveBulkState(ctx, store, stateItems...); err != nil {
		return errorz.Internal(err, "could not save state in store %q", store)
	}

	return nil
}

func (c *Client) GetState(ctx context.Context, store string, key string, target interface{}) error {
	state, err := c.client.GetState(ctx, store, key)
	if err != nil {
		return errorz.Internal(err, "could not load state %q", key)
	}
	if state.Value == nil {
		return errorz.NotFound("key %q not found", key)
	}
	if err = json.Unmarshal(state.Value, target); err != nil {
		return errorz.Internal(err, "could decode state %q", key)
	}
	return nil
}

func (c *Client) GetSecret(ctx context.Context, store string, name string, target interface{}) error {
	secret, err := c.client.GetSecret(ctx, store, name, nil)
	if err != nil {
		return errorz.Internal(err, "could not load secret %q", name)
	}
	if secret == nil {
		return errorz.NotFound("secret %q not found", name)
	}
	secretBytes, err := json.Marshal(secret)
	if err != nil {
		return errorz.Internal(err, "could decode secret %q", name)
	}
	err = json.Unmarshal(secretBytes, target)
	if err != nil {
		return errorz.Internal(err, "could decode secret %q", name)
	}
	return nil
}

func etag(value string) *dapr.ETag {
	if value == "" {
		return nil
	}
	return &dapr.ETag{
		Value: value,
	}
}
