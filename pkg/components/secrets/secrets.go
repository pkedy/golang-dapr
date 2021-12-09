package secrets

import (
	"context"
)

type Store interface {
	GetSecret(ctx context.Context, store string, name string, target interface{}) error
}
