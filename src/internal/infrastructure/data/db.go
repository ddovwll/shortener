package data

import (
	"time"

	"github.com/wb-go/wbf/dbpg"
)

type PostgresConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func InitDb(cfg PostgresConfig) (*dbpg.DB, error) {
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
