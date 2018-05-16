package colly

import (
	"github.com/go-redis/redis"
	"github.com/smileboywtu/FileCollector/common"
)

type CacheWriter interface {
	GetDestQueueSize() int64
	GetCacheEntry() ([]string, error)

	CacheFileEntry(files []string) error
	SendFileContent(buffer string) error
	BatchSendFileContent(buffers []string) error

	RemoveCacheEntry(files []string) error
	IsAllow() bool
}

type RedisWriter struct {
	RClient        *redis.Client
	CacheQueueName string
	DestQueueName  string
	QueueSizeLimit int
}

// NewRedisWriter init a new backend for cache and exchange
func NewRedisWriter(opts *redis.Options, cacheQName string, destQName string, qLimit int) (*RedisWriter, error) {
	client := redis.NewClient(opts)

	pong, err := client.Ping().Result()
	if err != nil || pong != "PONG" {
		return nil, &common.WriterError{Params: opts.Addr, Prob: "redis connection test fail"}
	}

	return &RedisWriter{
		RClient:        client,
		CacheQueueName: cacheQName,
		DestQueueName:  destQName,
		QueueSizeLimit: qLimit,
	}, nil
}

// GetDestQueueSize get current destination queue size
func (w *RedisWriter) GetDestQueueSize() int64 {
	clen, err := w.RClient.LLen(w.DestQueueName).Result()
	if err != nil {
		return 0
	}
	return clen
}

func (w *RedisWriter) CacheFileEntry(files []string) error {

	if !w.Check() {
		return &common.WriterError{Params: "files ...", Prob: "redis connection test fail"}
	}

	args := []interface{}{}
	for _, m := range files {
		args = append(args, m)
	}
	err := w.RClient.SAdd(w.CacheQueueName, args...).Err()
	return err
}

func (w *RedisWriter) GetCacheEntry() ([]string, error) {
	if !w.Check() {
		return nil, &common.WriterError{Params: "", Prob: "redis connection test fail"}
	}

	results, err := w.RClient.SMembers(w.CacheQueueName).Result()
	if err != nil {
		return nil, &common.WriterError{Params: w.CacheQueueName, Prob: err.Error()}
	}

	return results, nil
}

// RemoveCacheEntry delete cache file in records
func (w *RedisWriter) RemoveCacheEntry(files []string) error {
	if !w.Check() {
		return &common.WriterError{Params: "files ...", Prob: "redis connection test fail"}
	}

	args := []interface{}{}
	for _, m := range files {
		args = append(args, m)
	}
	err := w.RClient.SRem(w.CacheQueueName, args...).Err()
	return err
}

// SendFileContent send one file to redis
func (w *RedisWriter) SendFileContent(buffer string) error {
	if !w.Check() {
		return &common.WriterError{Params: "", Prob: "redis connection error"}
	}
	if !w.IsAllow() {
		return &common.WriterError{Params: "", Prob: "redis queue size limit"}
	}
	w.RClient.LPush(w.DestQueueName, buffer)
	return nil
}

// BatchSendFileContent sends multiple file content to redis client
// in one time
func (w *RedisWriter) BatchSendFileContent(buffers []string) error {

	if !w.Check() {
		return &common.WriterError{Params: "", Prob: "redis connection error"}
	}
	if !w.IsAllow() {
		return &common.WriterError{Params: "", Prob: "redis queue size limit"}
	}

	s := make([]interface{}, len(buffers))
	for i, v := range buffers {
		s[i] = v
	}
	w.RClient.LPush(w.DestQueueName, s...)

	return nil
}

// Check checks if redis client connection is ok
func (w *RedisWriter) Check() bool {

	pong, err := w.RClient.Ping().Result()

	if pong != "PONG" || err != nil {
		return false
	}
	return true
}

// IsAllow check if the dest queue size is out of limit size
func (w *RedisWriter) IsAllow() bool {
	currentSize, err := w.RClient.LLen(w.DestQueueName).Result()
	if err != nil {
		return false
	}

	if int(currentSize) >= w.QueueSizeLimit {
		return false
	}

	return true
}
