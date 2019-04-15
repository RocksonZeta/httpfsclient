package kv

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// var (
// ErrRedisClosed an error with message 'Redis is already closed'
// ErrRedisClosed = errors.New("Redis is already closed")
// ErrKeyNotFound an error with message 'Key $thekey doesn't found'
// ErrKeyNotFound = errors.New("Key '%s' doesn't found")
// )
type RedisConfig struct {
	Url string
	// MaxIdle 0 no limit
	MaxIdle int
	// MaxActive 0 no limit
	MaxActive int
	// IdleTimeout  time.Duration(5) * time.Minute
	IdleTimeout time.Duration
	// Prefix "myprefix-for-this-website". Default ""
	// Prefix string
}

type ServiceFactory struct {
	Config *RedisConfig
	Pool   *redis.Pool
}

func NewFactory(Config *RedisConfig) *ServiceFactory {
	r := &ServiceFactory{Config: Config}
	r.initPool()
	return r
}

func (s *ServiceFactory) Get() *Service {
	return &Service{Redis: s.Pool.Get()}
}

// Connect connects to the redis, called only once
func (s *ServiceFactory) initPool() {
	c := s.Config

	if c.IdleTimeout <= 0 {
		c.IdleTimeout = time.Duration(5) * time.Minute
	}

	Pool := &redis.Pool{IdleTimeout: time.Duration(5) * time.Minute, MaxIdle: c.MaxIdle, MaxActive: c.MaxActive}
	Pool.TestOnBorrow = func(c redis.Conn, t time.Time) error {
		_, err := c.Do("PING")
		return err
	}

	Pool.Dial = func() (redis.Conn, error) {
		return redis.DialURL(c.Url)
	}
	s.Pool = Pool
}
