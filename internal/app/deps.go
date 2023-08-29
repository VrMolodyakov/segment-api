package app

import (
	"context"
	"net/http"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/bufferpool"
	"github.com/VrMolodyakov/segment-api/internal/cache"
	"github.com/VrMolodyakov/segment-api/internal/config"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver"
	"github.com/VrMolodyakov/segment-api/internal/domain/cleaner"
	historyDomain "github.com/VrMolodyakov/segment-api/internal/domain/history"
	membershipDomain "github.com/VrMolodyakov/segment-api/internal/domain/membership"
	segmentDomain "github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/internal/repository/history"
	"github.com/VrMolodyakov/segment-api/internal/repository/membership"
	"github.com/VrMolodyakov/segment-api/internal/repository/segment"
	"github.com/VrMolodyakov/segment-api/pkg/client/postgresql"
	"github.com/VrMolodyakov/segment-api/pkg/clock"
	"github.com/VrMolodyakov/segment-api/pkg/csv"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/VrMolodyakov/segment-api/pkg/random"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maximumPercentage int           = 100
	cleanUpInterval   time.Duration = 5 * time.Minute
)

type Cleaner interface {
	Start(ctx context.Context, interval time.Duration)
}

type Deps struct {
	server   *http.Server
	psqlPool *pgxpool.Pool
	cleaner  Cleaner
}

func (d *Deps) Setup(ctx context.Context, cfg *config.Config, logger logging.Logger) error {
	pgCfg := postgresql.NewPgConfig(
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Database,
		cfg.Postgres.PoolSize,
		cfg.Postgres.SSLMode,
	)

	client, err := postgresql.NewClient(ctx, 5, 10*time.Second, pgCfg)
	if err != nil {
		logger.Errorf("couldn't create psql client %w", err)
		return err
	}
	d.psqlPool = client

	generator := random.NewRandomGenerator(maximumPercentage)
	clock := clock.New()

	historyRepo := history.New(d.psqlPool)
	membershipRepo := membership.New(d.psqlPool, clock)
	segmentRepo := segment.New(d.psqlPool)

	dataCache := cache.New[int64, []membershipDomain.MembershipInfo](cleanUpInterval)
	historyCache := cache.New[int, []historyDomain.History](cleanUpInterval)

	segmentService := segmentDomain.New(segmentRepo, logger)

	membershipService := membershipDomain.New(
		membershipRepo,
		dataCache,
		time.Duration(cfg.Cachce.SegmentExpiration)*time.Second,
		generator,
		logger,
	)

	historyService := historyDomain.New(
		historyRepo,
		historyCache,
		time.Duration(cfg.Cachce.CSVExpiration)*time.Second,
		logger,
	)

	d.cleaner = cleaner.New(membershipRepo, logger)

	pool := bufferpool.New()
	f := csv.Write[historyDomain.History]
	writer := csv.NewCSVWriter[historyDomain.History](f)

	d.cleaner = cleaner.New(membershipRepo, logger)
	d.server = apiserver.New(cfg.HTTP, segmentService, historyService, membershipService, pool, &writer)

	return nil
}

func (d *Deps) Close(ctx context.Context, logger logging.Logger) {
	if d.server != nil {
		if err := d.server.Shutdown(ctx); err != nil {
			logger.Errorf("couldn't close server %s", err.Error())
		}
	}

	if d.psqlPool != nil {
		d.psqlPool.Close()
	}
}
