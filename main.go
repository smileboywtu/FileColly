package main

import (
	"github.com/go-redis/redis"
	"fmt"
	"github.com/smileboywtu/FileCollector/colly"
	"os"
	"time"
	"github.com/go-ini/ini"
	"strconv"
)
import (
	_ "net/http/pprof"
	"net/http"
)

// Get bytes
func getBytes(size string) int64 {

	weight := size[len(size)-1]
	ret, err := strconv.Atoi(size[:len(size)-1])
	if err != nil {
		panic(ret)
	}
	ret64 := int64(ret)
	if weight == 'B' {
		return ret64
	} else if weight == 'K' {
		return ret64 * (1 << 10)
	} else if weight == 'M' {
		return ret64 * (1 << 20)
	} else if weight == 'G' {
		return ret64 * (1 << 30)
	}

	return ret64
}

// Parse user config
func parseConfig(file string) *colly.Collector {

	// load config file
	appconf, err := ini.Load(file)
	if err != nil {
		panic(err)
	}

	rhost := appconf.Section("redis").Key("host").String()
	rport := appconf.Section("redis").Key("port").String()
	rdb, err := appconf.Section("redis").Key("db").Int()
	if err != nil {
		panic(err)
	}

	var opts *redis.Options
	if passwd := appconf.Section("redis").Key("passwd").String(); len(passwd) == 0 {
		opts = &redis.Options{
			Addr:       fmt.Sprintf("%s:%s", rhost, rport),
			DB:         rdb,
			MaxRetries: 3,
		}
	} else {
		opts = &redis.Options{
			Addr:       fmt.Sprintf("%s:%s", rhost, rport),
			DB:         rdb,
			Password:   passwd,
			MaxRetries: 3,
		}
	}

	// init log
	lfile := appconf.Section("log").Key("log_file").String()
	colly.InitLogger(lfile)

	// init collector
	collyDir := appconf.Section("collector").Key("collect_dir").String()
	sendQName := appconf.Section("collector").Key("send_queue_name").String()
	cacheQName := appconf.Section("collector").Key("cache_queue_name").String()
	bufferLimitSize := appconf.Section("collector").Key("max_cache_size").String()
	maxFileSize := appconf.Section("collector").Key("max_file_size").String()
	queueLimitSize, err := appconf.Section("collector").Key("max_cache_file").Int()
	if err != nil {
		panic(err)
	}
	backend, err := colly.NewRedisWriter(opts, cacheQName, sendQName, queueLimitSize)
	if backend == nil || err != nil {
		fmt.Fprintf(os.Stderr, "redis connect error")
		os.Exit(-1)
	}
	BUFFERSIZE := 500
	scanner := colly.NewDirScanner(collyDir, BUFFERSIZE)
	return colly.NewCollector(
		backend,
		scanner,
		2*time.Second,
		1*time.Second,
		getBytes(bufferLimitSize),
		getBytes(maxFileSize))
}

func main() {

	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	c := parseConfig("config.ini")

	go c.Sync()
	go c.Cache()

	for {

		select {
		case <-c.SyncDone:
			go c.Sync()
		case <-c.CacheDone:
			go c.Cache()
		case <-time.After(2 * time.Second):
			break
		}
	}
}
