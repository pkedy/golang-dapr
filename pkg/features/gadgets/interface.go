package gadgets

import (
	"context"
)

type (
	Store interface {
		Load(ctx context.Context, id string) (*Gadget, error)
		Save(ctx context.Context, gadget *Gadget) error
	}

	Gadget struct {
		ID          string  `json:"id"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
	}
)
