package history

import (
	"context"
	"fmt"
	"time"

	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

var (
	ErrIncorrectDate = fmt.Errorf("history for dates before %d is not available", AvitoLaunchYear)
)

const (
	AvitoLaunchYear int = 2007
)

type HistoryRepository interface {
	Get(ctx context.Context, date Date) ([]History, error)
}

type Cache interface {
	Set(key int64, value []byte, expireAt time.Duration) []byte
	Get(key int64) ([]byte, bool)
}

type service struct {
	logger          logging.Logger
	cache           Cache
	cacheExpiration time.Duration
	history         HistoryRepository
}

func New(participation HistoryRepository, cache Cache, expiration time.Duration, logger logging.Logger) *service {
	return &service{
		history:         participation,
		cache:           cache,
		cacheExpiration: expiration,
		logger:          logger,
	}
}

func (s *service) GetUsersHistory(ctx context.Context, date Date) {

}
