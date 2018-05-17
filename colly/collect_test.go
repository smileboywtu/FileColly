// test performance for colly

//  4C 16G
//  239 second
//  file count: 500000  2092 f/s
//  filesize: 20 Bytes

//	go test -v -bench=. ./colly -run=BenchmarkCollector_Start
//	goos: linux
//	goarch: amd64
//	pkg: github.com/smileboywtu/FileColly/colly
//  BenchmarkCollector_Start-8               	       1	239480616423 ns/op

package colly

import (
	"testing"
)

func BenchmarkCollector_Start(b *testing.B) {
	appOptions := &AppConfigOption{
		RedisHost: "127.0.0.1",
		RedisPort: 6379,
		RedisDB:   0,
		RedisPW:   "",

		LoadCacheDB:                false,
		CacheRedisQueueName:        "cache:queue:tmp",
		DestinationRedisQueueName:  "cache:queue:dest",
		DestinationRedisQueueLimit: 500000,

		ReadWaitTime:     3,
		SenderMaxWorkers: 500,
		ReaderMaxWorkers: 500,

		FileMaxSize: "200M",

		ReserveFile: false,

		FileCacheTimeout: 3600,

		LogFileName: "../hack/sender.log",

		CollectDirectory: "../hack/test_data",
	}

	colly, errs := NewCollector(appOptions)
	if errs != nil {
		b.Error(errs)
	}

	colly.Start()

}
