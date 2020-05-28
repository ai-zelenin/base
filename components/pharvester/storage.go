package pharvester

import (
	"context"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	"git.pnhub.ru/core/libs/log"
)

type Storage struct {
	ctx    context.Context
	logger log.Logger
	db     *gorm.DB
	mx     sync.RWMutex
}

func NewStorage(ctx context.Context, logger log.Logger) (*Storage, error) {
	db, err := gorm.Open("sqlite3", "proxy.db")
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(new(Proxy)).Error
	if err != nil {
		return nil, err
	}
	return &Storage{
		ctx:    ctx,
		logger: logger,
		db:     db,
	}, nil
}

func (s *Storage) Save(proxy *Proxy) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.db.Save(proxy).Error
}

func (s *Storage) LoadValidationBunch(n int) ([]*Proxy, error) {
	var slice = make([]*Proxy, 0)
	q := s.db.Where("last_check < ? and status != ?", time.Now().UTC().Add(-time.Hour), ProxyStatusNeedScan).Order("last_check").Limit(n).Find(&slice)
	err := q.Error
	if err != nil {
		return nil, err
	}
	return slice, err
}
func (s *Storage) LoadBest() ([]*Proxy, error) {
	var slice = make([]*Proxy, 0)
	q := s.db.Where("score < ?", 0).Order("score DESC").Find(&slice)
	err := q.Error
	if err != nil {
		return nil, err
	}
	return slice, err
}
