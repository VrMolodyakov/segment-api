package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/config"
	"github.com/VrMolodyakov/segment-api/internal/repository/user"
	"github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	pgCfg := postgresql.NewPgConfig(
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.PoolSize,
		cfg.Postgres.SSLMode,
	)

	ctx := context.Background()
	client, err := postgresql.NewClient(ctx, 5, 10*time.Second, pgCfg)
	if err != nil {
		log.Fatal(err)
	}

	repo := user.New(client)
	id, err := repo.Get(ctx, 1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(id)

}
