// Test Suit for backend
package colly

import (
	"time"
	"testing"
	"strconv"

	"github.com/go-redis/redis"
)

func TestRedisWriter_CacheFileEntry(t *testing.T) {

	opts := &redis.Options{
		Addr:     "127.0.0.1:6379",
		DB:       0,
		Password: "",
	}
	CacheQueueName := "cache:queue"
	DestQueueName := "cache:dest"

	inst, errs := NewRedisWriter(opts, CacheQueueName, DestQueueName, 500)
	if errs != nil {
		t.Error(errs.Error())
	}

	errs = inst.CacheFileEntry("/tmp/a.txt", strconv.FormatInt(time.Now().Unix(), 10))
	if errs != nil {
		t.Error(errs.Error())
	}

}
