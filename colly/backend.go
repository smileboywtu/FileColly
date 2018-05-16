package colly

import (
	"os"
	"fmt"
	"bufio"
	"strings"
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
	RClient         *redis.Client
	CacheQueueName  string
	DestQueueName   string
	BackendDumpFile string `dump cache entries into file`
	QueueSizeLimit  int
}

// NewRedisWriter init a new backend for cache and exchange
func NewRedisWriter(opts *redis.Options, cacheQName string, destQName string, qLimit int) (*RedisWriter, error) {
	client := redis.NewClient(opts)

	pong, err := client.Ping().Result()
	if err != nil || pong != "PONG" {
		return nil, &common.WriterError{Params: opts.Addr, Prob: "redis connection test fail"}
	}

	return &RedisWriter{
		RClient:         client,
		CacheQueueName:  cacheQName,
		DestQueueName:   destQName,
		BackendDumpFile: "dumpdb.txt",
		QueueSizeLimit:  qLimit,
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

// CacheFileCheck checks if file path has cache in memory map
// if just return cache timestamp if not just return empty
func (w *RedisWriter) CacheFileCheck(filepath string) (string, error) {
	return w.RClient.HGet(w.CacheQueueName, filepath).Result()
}

func (w *RedisWriter) CacheFileEntry(filepath string, timestamp string) error {

	if !w.Check() {
		return &common.WriterError{Params: "files ...", Prob: "redis connection test fail"}
	}

	err := w.RClient.HSet(w.CacheQueueName, filepath, timestamp).Err()
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
func (w *RedisWriter) RemoveCacheEntry(filepath string) error {
	if !w.Check() {
		return &common.WriterError{Params: "files ...", Prob: "redis connection test fail"}
	}

	err := w.RClient.HDel(w.CacheQueueName, filepath).Err()
	return err
}

// DumpEntry2File dumps cache entries in hashmap to a file
func (w *RedisWriter) DumpEntry2File() error {
	results, errs := w.GetCacheEntry()
	if errs != nil {
		return errs
	}

	fd, errs := os.OpenFile(w.BackendDumpFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if errs != nil {
		return errs
	}
	defer fd.Close()

	for i := 0; i < len(results); i = i + 2 {
		fd.WriteString(fmt.Sprintf("%s %s\n", results[i], results[i+1]))
	}

	return nil
}

// LoadEntryFromDB load local cache db to backend
func (w *RedisWriter) LoadEntryFromDB() error {

	fd, errs := os.Open(w.BackendDumpFile)
	if errs != nil {
		return errs
	}
	defer fd.Close()

	reader := bufio.NewReader(fd)

	for {
		line, errs := reader.ReadString('\n')
		if errs != nil {
			break
		}

		line = strings.Trim(line, "\n")
		metas := strings.Split(line, " ")

		w.CacheFileEntry(metas[0], metas[1])
	}

	return nil
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
