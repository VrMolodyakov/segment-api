package main

import (
	_ "github.com/VrMolodyakov/segment-api/internal/domain/segment"
	_ "github.com/VrMolodyakov/segment-api/internal/domain/user"
	_ "github.com/VrMolodyakov/segment-api/internal/repository/history"
	_ "github.com/VrMolodyakov/segment-api/internal/repository/user"
	_ "github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	_ "github.com/VrMolodyakov/segment-api/pkg/clock"
)

func main() {
	// cfg, err := config.GetConfig()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// pgCfg := postgresql.NewPgConfig(
	// 	cfg.Postgres.User,
	// 	cfg.Postgres.Password,
	// 	cfg.Postgres.Host,
	// 	cfg.Postgres.Port,
	// 	cfg.Postgres.Database,
	// 	cfg.Postgres.PoolSize,
	// 	cfg.Postgres.SSLMode,
	// )

	// ctx := context.Background()
	// client, err := postgresql.NewClient(ctx, 5, 10*time.Second, pgCfg)
	// if err != nil {
	// 	log.Fatal(err)
	// }
}
