package colly

import (
	"fmt"
	"errors"

	"github.com/go-redis/redis"
)

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

// NewRedisWriter init a new backend for cache and exchange
func NewRedisWriter(opts *redis.Options, destQName string, qLimit int) (*RedisWriter, error) {
	client := redis.NewClient(opts)

	pong, err := client.Ping().Result()
	if err != nil || pong != "PONG" {
		return nil, errors.New(fmt.Sprintf("redis connect error: %s, connect params: %s", err, opts))
	}

	return &RedisWriter{
		Client:         client,
		DestQueueName:  destQName,
		QueueSizeLimit: qLimit,
	}, nil
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
		return errors.New(fmt.Sprintf("send file to destination error"))
	}
	if !w.IsAllow() {
		return errors.New(fmt.Sprintf("destination queue reach the limit size(%d)", w.QueueSizeLimit))
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
