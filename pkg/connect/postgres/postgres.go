package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/pkedy/golang-dapr/pkg/components/secrets"
)

type DBCreds struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func Connect(ctx context.Context, store secrets.Store,
	storeName, secretName string,
	afterConnect ...func(context.Context, *pgx.Conn) error) (*pgxpool.Pool, error) {
	var creds DBCreds
	if err := store.GetSecret(ctx, storeName, secretName, &creds); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("postgres://%s:%s@%s/%s",
		creds.Username, creds.Password, creds.Host, creds.Database)

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	if len(afterConnect) > 0 {
		config.AfterConnect = afterConnect[0]
	}

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, pool.Ping(ctx)
}
