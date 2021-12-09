package widgets

import (
	"context"
)

type (
	Store interface {
		Load(ctx context.Context, id string) (*Widget, error)
		Save(ctx context.Context, widget *Widget) error
	}

	Widget struct {
		ID          string  `json:"id"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
	}
)
