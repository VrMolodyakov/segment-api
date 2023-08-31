package integrationtest

import (
	"context"
	"database/sql"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/VrMolodyakov/segment-api/internal/bufferpool"
	"github.com/VrMolodyakov/segment-api/internal/cache"
	"github.com/VrMolodyakov/segment-api/internal/config"
	"github.com/VrMolodyakov/segment-api/internal/controller/http/v1/apiserver"
	historyDomain "github.com/VrMolodyakov/segment-api/internal/domain/history"
	membershipDomain "github.com/VrMolodyakov/segment-api/internal/domain/membership"
	segmentDomain "github.com/VrMolodyakov/segment-api/internal/domain/segment"
	"github.com/VrMolodyakov/segment-api/internal/repository/history"
	"github.com/VrMolodyakov/segment-api/internal/repository/membership"
	"github.com/VrMolodyakov/segment-api/internal/repository/segment"
	"github.com/VrMolodyakov/segment-api/pkg/clock"
	"github.com/VrMolodyakov/segment-api/pkg/csv"
	"github.com/VrMolodyakov/segment-api/pkg/logging"
	"github.com/VrMolodyakov/segment-api/pkg/random"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	psqlContainer *PSQLContainer
	logger        logging.Logger
	client        *pgxpool.Pool
	loader        *FixtureLoader
	server        *httptest.Server
}

const (
	cleanUpInterval   time.Duration = 5 * time.Minute
	segmentExpiration int           = 50
	host              string        = "localhost"
	port              int           = 8081
	CSVExpiration     int           = 50
)

func (s *TestSuite) SetupSuite() {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer ctxCancel()

	psqlContainer, err := NewPSQLContainer(ctx)
	s.Require().NoError(err)

	s.psqlContainer = psqlContainer

	m, err := migrate.New("file://../migrations", psqlContainer.GetDSN())
	s.Require().NoError(err)
	if err := m.Up(); err != nil {
		s.Require().NoError(err)
	}

	s.client, err = pgxpool.New(ctx, psqlContainer.GetDSN())
	s.Require().NoError(err)
	err = s.client.Ping(ctx)
	s.Require().NoError(err)
	s.logger, err = logging.MockLogger()
	s.Require().NoError(err)

	generator := random.NewRandomGenerator(100)
	clock := clock.New()

	historyRepo := history.New(s.client)
	membershipRepo := membership.New(s.client, clock)
	segmentRepo := segment.New(s.client)

	dataCache := cache.New[int64, []membershipDomain.MembershipInfo](cleanUpInterval)
	historyCache := cache.New[int, []historyDomain.History](cleanUpInterval)

	segmentService := segmentDomain.New(segmentRepo, s.logger)

	membershipService := membershipDomain.New(
		membershipRepo,
		dataCache,
		time.Duration(segmentExpiration)*time.Second,
		generator,
		s.logger,
	)

	historyService := historyDomain.New(
		historyRepo,
		historyCache,
		time.Duration(CSVExpiration)*time.Second,
		s.logger,
	)

	pool := bufferpool.New()
	f := csv.Write[historyDomain.History]
	writer := csv.NewCSVWriter[historyDomain.History](f)
	cfgHTTP := config.HTTP{
		Host:         host,
		Port:         port,
		ReadTimeout:  5,
		WriteTimeout: 5,
	}
	cfgDownload := config.Download{
		Host: host,
		Port: port,
	}
	server := apiserver.New(cfgHTTP, cfgDownload, segmentService, historyService, membershipService, pool, &writer)
	s.server = httptest.NewServer(server.Handler)

	s.loader = NewFixtureLoader(s.T(), Fixtures)

}

func (s *TestSuite) SetupTest() {
	db, err := sql.Open("postgres", s.psqlContainer.GetDSN())
	s.Require().NoError(err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgresql"),
		testfixtures.Directory("./fixtures/storage"),
	)
	s.Require().NoError(err)
	s.Require().NoError(fixtures.Load())
}

func TestSuite_Run(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
