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

type HistoryRepository interface {
	Get(ctx context.Context, date Date) ([]History, error)
}

type Cache interface {
	Set(key int, value []History, expireAt time.Duration) []History
	Get(key int) ([]History, bool)
}

type service struct {
	logger          logging.Logger
	cache           Cache
	cacheExpiration time.Duration
	history         HistoryRepository
}

func New(history HistoryRepository, cache Cache, expiration time.Duration, logger logging.Logger) *service {
	return &service{
		history:         history,
		cache:           cache,
		cacheExpiration: expiration,
		logger:          logger,
	}
}

func (s *service) GetUsersHistory(ctx context.Context, date Date) ([]History, error) {
	if err := date.Validate(); err != nil {
		s.logger.Errorf("invalid date : %s", err.Error())
		return nil, err
	}
	s.logger.Debugf("try to get users history for %d-%d", date.Year, date.Month)

	cacheKey := createCacheKey(date.Year, date.Month)
	if histories, inCache := s.cache.Get(cacheKey); inCache {
		return histories, nil
	}

	histories, err := s.history.Get(ctx, date)
	if err != nil {
		s.logger.Errorf("error in getting histories from the repository, %w", err)
		return nil, err
	}
	s.cache.Set(cacheKey, histories, s.cacheExpiration)
	return histories, nil
}

func createCacheKey(year int, month int) int {
	return year*100 + month
}
