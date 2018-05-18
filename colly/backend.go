package colly

import (
	"github.com/go-redis/redis"
	"github.com/smileboywtu/FileColly/common"
	"github.com/coreos/bbolt"
)

type CacheWriter interface {
	GetCacheEntry() (map[string]string, error)
	CacheFileEntry(filepath string, timestamp string) error
	RemoveCacheEntry(filepath string) error
}

type DestWriter interface {
	IsAllow() bool
	GetDestQueueSize() int64
	SendFileContent(buffer string) error
}

type RedisWriter struct {
	Client         *redis.Client
	DestQueueName  string
	QueueSizeLimit int
}

type BoltCacher struct {
	DBFile string
}

type Backend struct {
	Sender  *DestWriter
	Cacher  *CacheWriter
	CacheDB string
}

func NewBackend(opts *redis.Options, destQName string, qLimit int) (*Backend, error) {
	rc, errs := NewRedisWriter(opts, destQName, qLimit)
	if errs != nil {
		return nil, errs
	}
	return &Backend{
		Sender: rc,
		CacheDB: &BoltCacher{
			DBFile: "fscache.db",
		}
	}, nil

}

// NewRedisWriter init a new backend for cache and exchange
func NewRedisWriter(opts *redis.Options, destQName string, qLimit int) (*RedisWriter, error) {
	client := redis.NewClient(opts)

	pong, err := client.Ping().Result()
	if err != nil || pong != "PONG" {
		return nil, &common.WriterError{Params: opts.Addr, Prob: "redis connection test fail"}
	}

	return &RedisWriter{
		Client:         client,
		DestQueueName:  destQName,
		QueueSizeLimit: qLimit,
	}, nil
}

func NewFileCacheWriter(dbfile string) (*bolt.DB, error) {

}

func (w *BoltCacher) CacheFileEntry(filepath string, timestamp string) error {

}

func (w *BoltCacher) GetCacheEntry() (map[string]string, error) {

}

// RemoveCacheEntry delete cache file in records
func (w *BoltCacher) RemoveCacheEntry(filepath string) error {

}

// GetDestQueueSize get current destination queue size
func (w *RedisWriter) GetDestQueueSize() int64 {
	clen, err := w.Client.LLen(w.DestQueueName).Result()
	if err != nil {
		return 0
	}
	return clen
}

// SendFileContent send one file to redis
func (w *RedisWriter) SendFileContent(buffer string) error {
	if !w.Check() {
		return &common.WriterError{Params: "", Prob: "redis connection error"}
	}
	if !w.IsAllow() {
		return &common.WriterError{Params: "", Prob: "redis queue size limit"}
	}
	w.Client.LPush(w.DestQueueName, buffer)
	return nil
}

// Check checks if redis client connection is ok
func (w *RedisWriter) Check() bool {

	pong, err := w.Client.Ping().Result()

	if pong != "PONG" || err != nil {
		return false
	}
	return true
}

// IsAllow check if the dest queue size is out of limit size
func (w *RedisWriter) IsAllow() bool {
	currentSize, err := w.Client.LLen(w.DestQueueName).Result()
	if err != nil {
		return false
	}

	if int(currentSize) >= w.QueueSizeLimit {
		return false
	}

	return true
}
