// test performance for colly

//  4C 16G
//  239 second
//  file count: 10000  3000 f/s
//  filesize: 1024 Bytes

//	goos: linux
//	goarch: amd64
//	pkg: github.com/smileboywtu/FileColly/colly
//	BenchmarkCollector_Start100-8            	       1	2504965409 ns/op
//	BenchmarkCollector_Start300-8            	       1	3111993207 ns/op
//	BenchmarkCollector_Start500-8            	       1	3622320302 ns/op
//	BenchmarkCollector_Start1000-8           	       1	3786183160 ns/op


package colly

import (
	"testing"
	"os"
)

func baseCollector_Start(workers int, directory string, b *testing.B) {
	appOptions := &AppConfigOption{
		RedisHost: "127.0.0.1",
		RedisPort: 6379,
		RedisDB:   0,
		RedisPW:   "",

		CacheRedisQueueName:        "cache:queue:tmp",
		DestinationRedisQueueName:  "cache:queue:dest",
		DestinationRedisQueueLimit: 500000,

		ReadWaitTime:     3,
		SenderMaxWorkers: workers,
		ReaderMaxWorkers: workers,

		FileMaxSize: "200M",

		ReserveFile: false,

		FileCacheTimeout: 3600,

		LogFileName: "../hack/sender.log",

		CollectDirectory: directory,
	}

	colly, errs := NewCollector(appOptions)
	if errs != nil {
		b.Error(errs)
	}

	colly.Start()

	b.StopTimer()
	os.Remove("dumpdb.txt")
	b.StartTimer()

}

func BenchmarkCollector_Start100(b *testing.B) {
	baseCollector_Start(100, "../hack/100", b)
}

func BenchmarkCollector_Start300(b *testing.B) {
	baseCollector_Start(300, "../hack/300", b)
}

func BenchmarkCollector_Start500(b *testing.B) {
	baseCollector_Start(500, "../hack/500", b)
}

func BenchmarkCollector_Start1000(b *testing.B) {
	baseCollector_Start(1000, "../hack/1000", b)
}
