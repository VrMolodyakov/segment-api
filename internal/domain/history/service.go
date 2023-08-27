package history

import (
	"bytes"
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

type BufferPool interface {
	Get() *bytes.Buffer
	Release(buf *bytes.Buffer)
}

type Cache interface {
	Set(key int, value []byte, expireAt time.Duration) []byte
	Get(key int) ([]byte, bool)
}

type service struct {
	logger          logging.Logger
	cache           Cache
	cacheExpiration time.Duration
	pool            BufferPool
	history         HistoryRepository
}

func New(participation HistoryRepository, cache Cache, pool BufferPool, expiration time.Duration, logger logging.Logger) *service {
	return &service{
		history:         participation,
		cache:           cache,
		pool:            pool,
		cacheExpiration: expiration,
		logger:          logger,
	}
}

// func (s *service) GetUsersHistory(ctx context.Context, date Date) ([]History, error) {
// 	if err := date.Validate(); err != nil {
// 		s.logger.Errorf("invalid date : %s", err.Error())
// 		return nil, err
// 	}
// 	s.logger.Debugf("try to get users history for %d-%d", date.Year, date.Month)
// 	cacheKey := createCacheKey(date.Year, date.Month)
// 	histories, err := s.history.Get(ctx, date)
// 	if err != nil {
// 		return nil, err
// 	}
// 	w := csv.NewWriter(nil)

// }

// func createCacheKey(year int, month int) int {
// 	return year*100 + month
// }
