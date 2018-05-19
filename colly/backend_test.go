// Test Suit for backend
package colly

import (
	"time"
	"testing"
	"strconv"

	"github.com/pkg/errors"
	"github.com/go-redis/redis"
	"fmt"
)

var opts = &redis.Options{
	Addr:     "127.0.0.1:6379",
	DB:       0,
	Password: "",
}
var DestQueueName = "cache:dest"

func TestNewRedisWriter(t *testing.T) {
	inst, errs := NewRedisWriter(opts, DestQueueName, 500)
	if errs != nil {
		t.Fatal(errs.Error())
	}

	if inst.Check() {
		t.Log("test create new backend inst done")
	} else {
		t.Error("test create new backend inst error")
	}
}

func TestBoltCacher_CacheFileEntry(t *testing.T) {

	cache, errs := NewFileCacheWriter("fscache.db")
	if errs != nil {
		t.Fatal(errors.New("create cacher fails"))
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	filepath := "/tmp/a.txt"

	errs = cache.CacheFileEntry(filepath, timestamp)
	if errs != nil {
		t.Fatal(errors.New(fmt.Sprintf("put cache file error: %s", errs)))
	}

	before, errs := cache.GetCacheEntry(filepath)
	if errs != nil {
		t.Fatal(errors.New(fmt.Sprintf("get cache file error: %s", errs)))
	}

	if before != timestamp {
		t.Fatal(errors.New("cache file content error"))
	}
}

func TestBoltCacher_RemoveCacheEntry(t *testing.T) {
	cachetimeout := 2
	cache, errs := NewFileCacheWriter("fscache.db")
	if errs != nil {
		t.Fatal(errors.New("create cacher fails"))
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	filepath := "/tmp/a.txt"

	errs = cache.CacheFileEntry(filepath, timestamp)
	if errs != nil {
		t.Fatal(errors.New(fmt.Sprintf("put cache file error: %s", errs)))
	}

	done := make(chan bool, 1)
	time.AfterFunc(time.Duration(cachetimeout+1)*time.Second, func() {
		timestampbefore, errs := cache.GetCacheEntry(filepath)
		if errs != nil {
			t.Fatal(errors.New("cache file lost"))
		}
		if before, _ := strconv.ParseInt(timestampbefore, 10, 64); before+int64(cachetimeout) <= time.Now().Unix() {
			cache.RemoveCacheEntry(filepath)
		}

		if _, errs := cache.GetCacheEntry(filepath); errs == nil {
			t.Fatal(errors.New("cache file not remove after cache timeout"))
		}

		done <- true
	})

	<-done
}

func BenchmarkBoltCacher_CacheFileEntry(b *testing.B) {
	b.StopTimer()
	cache, errs := NewFileCacheWriter("fscache.db")
	if errs != nil {
		b.Fatal(errors.New("create cacher fails"))
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	filepath := "/tmp/a.txt"

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.CacheFileEntry(filepath, timestamp)
	}
}

func BenchmarkBoltCacher_GetCacheEntry(b *testing.B) {
	b.StopTimer()
	cache, errs := NewFileCacheWriter("fscache.db")
	if errs != nil {
		b.Fatal(errors.New("create cacher fails"))
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	filepath := "/tmp/a.txt"
	cache.CacheFileEntry(filepath, timestamp)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		cache.GetCacheEntry(filepath)
	}
}
