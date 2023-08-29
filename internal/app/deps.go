package app

// import (
// 	"context"
// 	"net/http"
// 	"time"

// 	"github.com/VrMolodyakov/segment-api/internal/config"
// 	"github.com/VrMolodyakov/segment-api/internal/repository/history"
// 	"github.com/VrMolodyakov/segment-api/internal/repository/membership"
// 	"github.com/VrMolodyakov/segment-api/internal/repository/segment"
// 	"github.com/VrMolodyakov/segment-api/internal/repository/user"
// 	historyDomain"github.com/VrMolodyakov/segment-api/internal/domain/history"
// 	membershipDomain"github.com/VrMolodyakov/segment-api/internal/domain/membership"
// 	segmentDomain"github.com/VrMolodyakov/segment-api/internal/domain/segment"
// 	userDomain"github.com/VrMolodyakov/segment-api/internal/domain/user"
// 	"github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
// 	"github.com/VrMolodyakov/segment-api/pkg/clock"
// 	"github.com/VrMolodyakov/segment-api/pkg/logging"
// 	"github.com/VrMolodyakov/segment-api/pkg/random"
// 	"github.com/jackc/pgx/v5/pgxpool"
// )

// const (
// 	maximumPercentage int = 100
// )

// type Deps struct {
// 	segmentServer *http.Server
// 	postgresPool  *pgxpool.Pool
// }

// func (d *Deps) Setup(ctx context.Context, cfg *config.Config, logger logging.Logger) error {
// 	pgCfg := postgresql.NewPgConfig(
// 		cfg.Postgres.User,
// 		cfg.Postgres.Password,
// 		cfg.Postgres.Host,
// 		cfg.Postgres.Port,
// 		cfg.Postgres.Database,
// 		cfg.Postgres.PoolSize,
// 		cfg.Postgres.SSLMode,
// 	)

// 	client, err := postgresql.NewClient(ctx, 5, 10*time.Second, pgCfg)
// 	if err != nil {
// 		logger.Errorf("couldn't create psql client %w",err)
// 		return err
// 	}

// 	generator := random.NewRandomGenerator(maximumPercentage)
// 	clock := clock.New()

// 	userRepo := user.New(client)
// 	historyRepo := history.New(client)
// 	membershipRepo := membership.New(client, clock, generator)
// 	segmentRepo := segment.New(client)

// 	segmentService := segmentDomain.New(segmentRepo,logger)
// 	userService := userDomain.New(userRepo,logger)
// 	membershipDomain.

// 	return nil
// }
