package data

import (
	"shortener/src/internal/application/config"

	"github.com/wb-go/wbf/dbpg"
)

func InitDb(cfg config.PostgresConfig) (*dbpg.DB, error) {
	opts := &dbpg.Options{
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	}
	db, err := dbpg.New(cfg.DSN, []string{}, opts)
	if err != nil {
		return nil, err
	}

	return db, nil
}
