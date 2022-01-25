package repository

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/pkedy/golang-dapr/pkg/errorz"
	"github.com/pkedy/golang-dapr/pkg/features/widgets"
)

const (
	nameUpsert = "upsert"
	sqlUpsert  = `INSERT INTO widgets (id, description, price)
	VALUES ($1, $2, $3)
	ON CONFLICT ON CONSTRAINT widgets_pkey
	DO UPDATE SET description = $2, price = $3;`

	nameSelect = "select"
	sqlSelect  = `SELECT description, price FROM widgets WHERE id = $1`
)

type Repository struct {
	log  logr.Logger
	pool *pgxpool.Pool
}

func New(log logr.Logger, pool *pgxpool.Pool) *Repository {
	return &Repository{
		log:  log,
		pool: pool,
	}
}

func AfterConnect(ctx context.Context, conn *pgx.Conn) (err error) {
	if _, err = conn.Prepare(ctx, nameUpsert, sqlUpsert); err != nil {
		return err
	}
	if _, err = conn.Prepare(ctx, nameSelect, sqlSelect); err != nil {
		return err
	}
	return nil
}

func (r *Repository) Save(ctx context.Context, widget *widgets.Widget) error {
	r.log.Info("Saving widget to DB", "widget", widget)
	if _, err := r.pool.Exec(ctx, nameUpsert, widget.ID, widget.Description, widget.Price); err != nil {
		r.log.Error(err, "error saving widget", "widget", widget)
		return errorz.Internal(err, "could not save widget %q", widget.ID)
	}
	return nil
}

func (r *Repository) Load(ctx context.Context, id string) (*widgets.Widget, error) {
	r.log.Info("Loading widget from DB", "id", id)
	row := r.pool.QueryRow(ctx, nameSelect, id)
	var description string
	var price float64
	if err := row.Scan(&description, &price); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorz.NotFound("widget with id %q was not found", id)
		}
		r.log.Info("error loading widget", "id", id)
		return nil, errorz.Internal(err, "could not load widget %q", id)
	}
	return &widgets.Widget{
		ID:          id,
		Description: description,
		Price:       price,
	}, nil
}
