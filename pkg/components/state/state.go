package state

import (
	"context"
)

type Item struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
	ETag  string      `json:"etag,omitempty"`
}

type Store interface {
	SetState(ctx context.Context, store string, items ...Item) error
	GetState(ctx context.Context, store string, key string, target interface{}) error
}
