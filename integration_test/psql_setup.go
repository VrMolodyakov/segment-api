package integrationtest

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PSQLContainer struct {
	testcontainers.Container
	MappedPort string
	Host       string
}

func (c PSQLContainer) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "postgres", "postgres", c.Host, c.MappedPort, "segment_test")
}

func NewPSQLContainer(ctx context.Context) (*PSQLContainer, error) {
	req := testcontainers.ContainerRequest{
		Env: map[string]string{
			"POSTGRES_USER":      "postgres",
			"POSTGRES_PASSWORD":  "postgres",
			"POSTGRES_DB":        "segment_test",
			"POSTGRES_HOST":      "postgres",
			"POSTGRES_PORT":      "5432",
			"POSTGRES_POOL_SIZE": "100",
			"POSTGRES_SSL_MODE":  "disable",
		},
		ExposedPorts: []string{"5432/tcp"},
		Image:        "postgres:14",
		WaitingFor: wait.ForExec([]string{"pg_isready", "-d", "segment_test", "-U", "postgres"}).
			WithPollInterval(1 * time.Second).
			WithExitCodeMatcher(func(exitCode int) bool {
				return exitCode == 0
			}),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	return &PSQLContainer{
		Container:  container,
		MappedPort: mappedPort.Port(),
		Host:       host,
	}, nil
}
