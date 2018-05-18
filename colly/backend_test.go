// Test Suit for backend
package colly

import (
	"time"
	"testing"
	"strconv"

	"github.com/pkg/errors"
	"github.com/go-redis/redis"
)

var opts = &redis.Options{
	Addr:     "127.0.0.1:6379",
	DB:       0,
	Password: "",
}
var CacheQueueName = "cache:queue"
var DestQueueName = "cache:dest"

func TestNewRedisWriter(t *testing.T) {
	inst, errs := NewRedisWriter(opts, CacheQueueName, DestQueueName, 500)
	if errs != nil {
		t.Fatal(errs.Error())
	}

	if inst.Check() {
		t.Log("test create new backend inst done")
	} else {
		t.Error("test create new backend inst error")
	}
}

func TestRedisWriter_CacheFileEntry(t *testing.T) {

	inst, errs := NewRedisWriter(opts, CacheQueueName, DestQueueName, 500)
	if errs != nil {
		t.Fatal(errs.Error())
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	filepath := "/tmp/a.txt"
	before, errs := inst.CacheFileCheck(filepath)
	if errs == nil && before == "" {
		t.Fatal(errors.New("redis library do not work as expected"))
	}

	errs = inst.CacheFileEntry(filepath, timestamp)
	if errs != nil {
		t.Fatal(errs.Error())
	}

	before, errs = inst.CacheFileCheck(filepath)
	if errs != nil {
		t.Fatal(errs)
	}

	if before != timestamp {
		t.Fatal(errors.New("timestamp not equal"))
	}
}

//func TestRedisWriter_DumpEntry2File(t *testing.T) {
//	inst, errs := NewRedisWriter(opts, CacheQueueName, DestQueueName, 500)
//	if errs != nil {
//		t.Fatal(errs.Error())
//	}
//
//	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
//	filepath := "/tmp/a.txt"
//
//	errs = inst.CacheFileEntry(filepath, timestamp)
//	if errs != nil {
//		t.Fatal(errs)
//	}
//
//	errs = inst.DumpEntry2File()
//	if errs != nil {
//		t.Fatal(errs)
//	}
//
//	// remove cache
//	inst.RemoveCacheEntry(filepath)
//	inst.LoadEntryFromDB()
//
//	os.Remove("dumpdb.txt")
//	before, errs := inst.CacheFileCheck(filepath)
//	if before != timestamp {
//		t.Error(errors.New("dumps and loads DB from local file fails"))
//	}
//
//}

func TestRedisWriter_RemoveCacheEntry(t *testing.T) {
	cachetimeout := 2

	inst, errs := NewRedisWriter(opts, CacheQueueName, DestQueueName, 500)
	if errs != nil {
		t.Fatal(errs.Error())
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	filepath := "/tmp/a.txt"

	inst.CacheFileEntry(filepath, timestamp)

	done := make(chan bool, 1)
	time.AfterFunc(time.Duration(cachetimeout+1)*time.Second, func() {
		timestampbefore, errs := inst.CacheFileCheck(filepath)
		if errs != nil {
			t.Fatal(errors.New("cache file lost"))
		}
		if before, _ := strconv.ParseInt(timestampbefore, 10, 64); before+int64(cachetimeout) <= time.Now().Unix() {
			inst.RemoveCacheEntry(filepath)
		}

		if _, errs := inst.CacheFileCheck(filepath); errs == nil {
			t.Fatal(errors.New("cache file not remove after cache timeout"))
		}

		done <- true
	})

	<-done
}
