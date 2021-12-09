package dapr

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/multierr"

	"github.com/pkedy/golang-dapr/pkg/components/secrets"
	"github.com/pkedy/golang-dapr/pkg/components/state"
	"github.com/pkedy/golang-dapr/pkg/errorz"
)

type HTTP struct{}

var (
	APIURL = fmt.Sprintf("http://127.0.0.1:%s/", os.Getenv("DAPR_HTTP_PORT"))

	_ = state.Store((*HTTP)(nil))
	_ = secrets.Store((*HTTP)(nil))
)

func NewHTTP(ctx context.Context) (*HTTP, error) {
	return &HTTP{}, nil
}

func (c *HTTP) Name() string {
	return "Custom HTTP (using Fiber client)"
}

func (c *HTTP) SetState(ctx context.Context, store string, items ...state.Item) error {
	url := APIURL + path.Join("v1.0/state", store)
	a := fiber.Post(url)
	defer fiber.ReleaseAgent(a)
	code, _, errs := a.JSON(items).Bytes()
	if code/100 != 2 {
		return errorz.Internal(fmt.Errorf("received %d status", code), "could not save state")
	}
	if len(errs) > 0 {
		return errorz.Internal(multierr.Combine(errs...), "could not save state")
	}

	return nil
}

func (c *HTTP) GetState(ctx context.Context, store string, key string, target interface{}) error {
	url := APIURL + path.Join("v1.0/state", store, key)
	a := fiber.Get(url)
	defer fiber.ReleaseAgent(a)
	code, _, errs := a.Struct(target)
	if code == 204 || code == 404 {
		return errorz.NotFound("key %q not found", key)
	}
	if len(errs) > 0 {
		return errorz.Internal(multierr.Combine(errs...), "could not load key %q", key)
	}
	return nil
}

func (c *HTTP) GetSecret(ctx context.Context, store string, name string, target interface{}) error {
	url := APIURL + path.Join("v1.0/secrets", store, name)
	a := fiber.Get(url)
	defer fiber.ReleaseAgent(a)
	code, _, errs := a.Struct(target)
	if code == 204 || code == 404 {
		return errorz.NotFound("secret %q not found", name)
	}
	if len(errs) > 0 {
		return errorz.Internal(multierr.Combine(errs...), "could not load secret %q", name)
	}
	return nil
}
