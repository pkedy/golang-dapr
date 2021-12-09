package products

import (
	"context"
)

type (
	Store interface {
		Load(ctx context.Context, id string) (*Product, error)
		Save(ctx context.Context, product *Product) error
	}

	Product struct {
		ID          string  `json:"id"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
	}
)
