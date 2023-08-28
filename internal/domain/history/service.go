package history

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/VrMolodyakov/segment-api/pkg/logging"
)

var (
	ErrIncorrectYear  = fmt.Errorf("history for dates before %d year is not available", AvitoLaunchYear)
	ErrIncorrectMonth = fmt.Errorf("impossible to get information for a month that has not yet come")
	ErrExpiredData    = errors.New("data lifetime is over")
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
	return nil, ErrExpiredData
}

func (s *service) PrepareHistoryData(ctx context.Context, date Date) error {
	if err := date.Validate(); err != nil {
		s.logger.Errorf("invalid date : %s", err.Error())
		return err
	}
	s.logger.Debugf("try to prepare history for %d-%d", date.Year, date.Month)

	cacheKey := createCacheKey(date.Year, date.Month)
	if _, inCache := s.cache.Get(cacheKey); inCache {
		return nil
	}

	histories, err := s.history.Get(ctx, date)
	if err != nil {
		s.logger.Errorf("error in getting histories from the repo, %w", err)
		return err
	}
	s.cache.Set(cacheKey, histories, s.cacheExpiration)
	return nil
}

func createCacheKey(year int, month int) int {
	return year*100 + month
}
