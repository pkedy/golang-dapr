package repository

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/pkedy/golang-dapr/pkg/components/state"
	"github.com/pkedy/golang-dapr/pkg/errorz"
	"github.com/pkedy/golang-dapr/pkg/features/gadgets"
)

type Repository struct {
	log         logr.Logger
	stateClient state.Store
	store       string
}

func New(log logr.Logger, stateClient state.Store, store string) *Repository {
	return &Repository{
		log:         log,
		stateClient: stateClient,
		store:       store,
	}
}

func (r *Repository) Save(ctx context.Context, gadget *gadgets.Gadget) error {
	r.log.Info("Saving gadget state", "gadget", gadget)
	if err := r.stateClient.SetState(ctx, r.store, state.Item{
		Key:   "gadget:" + gadget.ID,
		Value: &gadget,
	}); err != nil {
		return errorz.From(err).
			WithMessage("could not save gadget %q", gadget.ID)
	}
	return nil
}

func (r *Repository) Load(ctx context.Context, id string) (*gadgets.Gadget, error) {
	r.log.Info("Loading gadget state", "id", id)
	var gadget gadgets.Gadget
	if err := r.stateClient.GetState(ctx, r.store, "gadget:"+id, &gadget); err != nil {
		err := errorz.From(err)
		if err.Code == 404 {
			return nil, err.WithMessage("gadget %q not found", id)
		}
		return nil, err.WithMessage("could not load gadget %q", id)
	}

	return &gadget, nil
}
