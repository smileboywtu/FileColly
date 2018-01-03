package colly

import (
	"github.com/go-redis/redis"
	"fmt"
)

type CacheWriter interface {
	GetCacheQueueSize() int64
	CacheFileEntry(files []string) error
	RemoveCacheEntry(files []string) error
	GetCacheEntry() ([]string, error)
	CacheFileContent(buffer string) error
	BatchCacheFileContent(buffers []string) error
	IsAllow() bool
}

type RedisWriter struct {
	RClient    *redis.Client `redis client`
	ExQName    string        `exchange redis queue name`
	CacheQName string        `cache file redis queue name`
	QueueLimit int           `redis cache queue limit`
}
type WriterError struct {
	param string
	prob  string
}

func (e *WriterError) Error() string {
	return fmt.Sprintf("%s - %s", e.param, e.prob)
}

func NewRedisWriter(opts *redis.Options, exchange_qname string, cache_qname string, queue_limit int) (*RedisWriter, error) {
	client := redis.NewClient(opts)

	pong, err := client.Ping().Result()
	if err != nil || pong != "PONG" {
		return nil, &WriterError{param: opts.Addr, prob: "redis connection test fail"}
	}

	return &RedisWriter{
		RClient:    client,
		ExQName:    exchange_qname,
		CacheQName: cache_qname,
		QueueLimit: queue_limit,
	}, nil
}

func (w *RedisWriter) GetCacheQueueSize() int64 {
	clen, err := w.RClient.LLen(w.CacheQName).Result()
	if err != nil {
		return 0
	}
	return clen
}

func (w *RedisWriter) CacheFileEntry(files []string) error {

	if !w.Check() {
		return &WriterError{param: "files ...", prob: "redis connection test fail"}
	}

	args := []interface{}{}
	for _, m := range files {
		args = append(args, m)
	}
	err := w.RClient.SAdd(w.ExQName, args...).Err()
	return err
}

func (w *RedisWriter) GetCacheEntry() ([]string, error) {
	if !w.Check() {
		return nil, &WriterError{param: "", prob: "redis connection test fail"}
	}

	results, err := w.RClient.SMembers(w.ExQName).Result()
	if err != nil {
		return nil, &WriterError{param: w.ExQName, prob: err.Error()}
	}

	return results, nil
}

func (w *RedisWriter) RemoveCacheEntry(files []string) error {
	if !w.Check() {
		return &WriterError{param: "files ...", prob: "redis connection test fail"}
	}

	args := []interface{}{}
	for _, m := range files {
		args = append(args, m)
	}
	err := w.RClient.SRem(w.ExQName, args...).Err()
	return err
}

func (w *RedisWriter) CacheFileContent(buffer string) error {
	if !w.Check() {
		return &WriterError{param: "", prob: "redis connection error"}
	}
	if !w.IsAllow() {
		return &WriterError{param: "", prob: "redis queue size limit"}
	}
	w.RClient.LPush(w.CacheQName, buffer)
	return nil
}

func (w *RedisWriter) BatchCacheFileContent(buffers []string) error {

	if !w.Check() {
		return &WriterError{param: "", prob: "redis connection error"}
	}
	if !w.IsAllow() {
		return &WriterError{param: "", prob: "redis queue size limit"}
	}

	s := make([]interface{}, len(buffers))
	for i, v := range buffers {
		s[i] = v
	}
	w.RClient.LPush(w.CacheQName, s...)

	return nil
}

func (w *RedisWriter) Check() bool {

	pong, err := w.RClient.Ping().Result()

	if pong != "PONG" || err != nil {
		return false
	}
	return true
}

func (w *RedisWriter) IsAllow() bool {
	currentSize, err := w.RClient.LLen(w.CacheQName).Result()
	if err != nil {
		return false
	}

	if int(currentSize) >= w.QueueLimit {
		return false
	}

	return true
}
