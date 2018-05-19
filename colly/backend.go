package colly

import (
	"fmt"
	"errors"

	"github.com/coreos/bbolt"
	"github.com/go-redis/redis"
	"github.com/smileboywtu/FileColly/common"
)

type CacheWriter interface {
	GetCacheEntry(filepath string) (string, error)
	CacheFileEntry(filepath string, timestamp string) error
	CacheFileLookup(filepath string) bool
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
	DBFile     string
	BucketName []byte
}

type Backend struct {
	Sender DestWriter
	Cacher CacheWriter
}

// NewBackend create new backend
func NewBackend(opts *redis.Options, destQName string, qLimit int) (*Backend, error) {
	sender, errs := NewRedisWriter(opts, destQName, qLimit)
	if errs != nil {
		return nil, errs
	}

	cacher, errs := NewFileCacheWriter("fscache.db")
	if errs != nil {
		return nil, errs
	}

	return &Backend{
		Sender: sender,
		Cacher: cacher,
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

// NewFileCacheWriter create new DB file
func NewFileCacheWriter(dbfile string) (*BoltCacher, error) {
	db, errs := bolt.Open(dbfile, 0600, nil)
	if errs != nil {
		return nil, errors.New("open backend cache db failed")
	}
	defer db.Close()

	// create bucket
	bucketName := "fscacher"
	errs = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return errors.New(fmt.Sprintf("create bucket: %s", err))
		}
		return nil
	})
	if errs != nil {
		return nil, errs
	}

	return &BoltCacher{
		DBFile:     dbfile,
		BucketName: []byte(bucketName),
	}, nil
}

func (w *BoltCacher) CacheFileLookup(filepath string) bool {
	value, _ := w.GetCacheEntry(filepath)
	return len(value) > 0
}

func (w *BoltCacher) CacheFileEntry(filepath string, timestamp string) error {
	db, errs := bolt.Open(w.DBFile, 0600, nil)
	if errs != nil {
		return errors.New("open backend cache db failed")
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(w.BucketName).Put([]byte(filepath), []byte(timestamp))
	})
}

func (w *BoltCacher) GetCacheEntry(filepath string) (string, error) {
	db, errs := bolt.Open(w.DBFile, 0600, &bolt.Options{1})
	if errs != nil {
		return "", errors.New("open backend cache db failed")
	}
	defer db.Close()

	var value string
	db.View(func(tx *bolt.Tx) error {
		if ret := tx.Bucket([]byte(w.BucketName)).Get([]byte(filepath)); ret != nil {
			value = string(ret)
		} else {
			value = ""
		}
		return nil
	})

	if len(value) == 0 {
		return "", errors.New(fmt.Sprintf("key %s does not exist.", filepath))
	}
	return value, nil
}

// RemoveCacheEntry delete cache file in records
func (w *BoltCacher) RemoveCacheEntry(filepath string) error {
	db, errs := bolt.Open(w.DBFile, 0600, nil)
	if errs != nil {
		return errors.New("open backend cache db failed")
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(w.BucketName).Delete([]byte(filepath))
	})
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
