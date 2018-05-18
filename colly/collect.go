// colly define the flow to file collecting
// and cache policy
package colly

import (
	"os"
	"log"
	"fmt"
	"sync"
	"time"
	"strconv"
	"context"
	"io/ioutil"
	"github.com/pkg/errors"
	"github.com/go-redis/redis"
	"gopkg.in/natefinch/lumberjack.v2"
	"github.com/smileboywtu/FileColly/common"
)

var logger *log.Logger

type Collector struct {
	sync.RWMutex

	// App Configs
	AppConfigs *AppConfigOption

	// Backend Redis Instance
	BackendInst *RedisWriter

	// File Walker Instance
	FileWalkerInst *FileWalker

	// File filters
	Rule    Rule
	filters []FilterFuncs

	// Files Deal numbers
	FileCount int64

	ctx        context.Context
	cancleFunc context.CancelFunc
}

func InitLogger(logFile string) {

	fd, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open log file error")
	}
	logger = log.New(fd, "collector: ", log.Lshortfile)
	logger.SetOutput(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	})
}

// NewCollector init a collector to collect file in directory
func NewCollector(opts *AppConfigOption) (*Collector, error) {

	ctx, cancle := context.WithCancel(context.Background())

	// init logger
	InitLogger(opts.LogFileName)

	// init backend option
	redisOpts := &redis.Options{
		Addr:       fmt.Sprintf("%s:%d", opts.RedisHost, opts.RedisPort),
		DB:         opts.RedisDB,
		Password:   opts.RedisPW,
		MaxRetries: 3,
	}
	backend, errs := NewRedisWriter(redisOpts, opts.CacheRedisQueueName, opts.DestinationRedisQueueName, opts.DestinationRedisQueueLimit)
	if backend == nil || errs != nil {
		return nil, errs
	}

	rule := Rule{
		FileSizeLimit:   common.HumanSize2Bytes(opts.FileMaxSize),
		ReserveFile:     opts.ReserveFile,
		CollectWaitTime: opts.ReadWaitTime,
		AllowEmpty:      false,
	}

	return &Collector{
		AppConfigs:     opts,
		BackendInst:    backend,
		FileWalkerInst: NewDirectoryWorker(opts.CollectDirectory, opts.ReaderMaxWorkers, rule, ctx),
		FileCount:      0,
		Rule:           rule,
		filters:        make([]FilterFuncs, 0, 8),

		ctx:        ctx,
		cancleFunc: cancle,
	}, nil
}

// OnFilter add new filter to collector
func (c *Collector) OnFilter(callback FilterFuncs) {
	c.Lock()
	c.filters = append(c.filters, callback)
	c.Unlock()
}

func (c *Collector) GetFileCount() int64 {
	return c.FileCount
}

func (c *Collector) IncreaseFileCount(n int) {
	c.Lock()
	c.FileCount += int64(n)
	c.Unlock()
}

func (c *Collector) CountClear() {
	c.Lock()
	c.FileCount = 0
	c.Unlock()
}

// ListCacheFiles get current cache file from backend
func (c *Collector) ListCacheFiles() map[string]string {
	if result, err := c.BackendInst.GetCacheEntry(); err != nil {
		return nil
	} else {
		return result
	}
}

// sendPoll send file to
func (c *Collector) sendPoll(result chan<- EncodeResult, item EncodeResult) {
	select {
	case result <- item:
	case <-c.ctx.Done():
		return
	}
}

// encodeFlow encodes file content and send to backend
func (c *Collector) encodeFlow(fileItems <-chan FileItem, result chan<- EncodeResult, cacheBuffer chan<- string) {

	for item := range fileItems {

		if !c.GetMatch(item.FilePath) {
			c.sendPoll(result, EncodeResult{item.FilePath, "", errors.New("file not match")})
			continue
		}

		// send to cache checkup
		go func() {
			cacheBuffer <- item.FilePath
		}()

		// file has been send do not cache again
		if c.AppConfigs.ReserveFile && c.BackendInst.CacheFastLookup(item.FilePath) {
			continue
		}

		data, err := ioutil.ReadFile(item.FilePath)
		if err != nil {
			c.sendPoll(result, EncodeResult{item.FilePath, "", err})
		}
		encoder := &FileContentEncoder{
			FilePath:    item.FileIndex,
			FileContent: make([]byte, len(data)),
		}
		copy(encoder.FileContent, data)
		packBytes, err := encoder.Encode()
		c.sendPoll(result, EncodeResult{item.FilePath, packBytes, err})
	}

}

func (c *Collector) Start() {

	var wg sync.WaitGroup
	buffers := make(chan EncodeResult)
	fileItems, errc := c.FileWalkerInst.Walk()

	c.CountClear()
	c.IncreaseFileCount(int(c.BackendInst.GetDestQueueSize()))

	// load cache entry from db
	if c.AppConfigs.LoadCacheDB {
		errs := c.BackendInst.LoadEntryFromDB()
		if errs != nil {
			logger.Println("load cache to backend error: ", errs.Error())
		}
	}

	// start cache flow
	cacheBuffer := c.cacheFlow()
	defer close(cacheBuffer)

	wg.Add(c.AppConfigs.ReaderMaxWorkers)
	for i := 0; i < c.AppConfigs.ReaderMaxWorkers; i++ {
		go func() {
			c.encodeFlow(fileItems, buffers, cacheBuffer)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(buffers)
	}()

	// wait all buffer deal done
	c.sendFlow(buffers)

	if err := <-errc; err != nil {
		fmt.Println(err.Error())
		logger.Println(err.Error())
	}

	logger.Printf("current time: %s, send file total: %d", time.Now().Format("2006-01-02T15:04:05"), c.FileCount)
	c.BackendInst.DumpEntry2File()
}

// sendFlow cache current file in pipeline and remove file from directory
// if queue is out of limit size or reserve file is true, then do nothing
// about the file
func (c *Collector) sendFlow(buffers <-chan EncodeResult) {

	var wg sync.WaitGroup
	wg.Add(c.AppConfigs.SenderMaxWorkers)

	for i := 0; i < c.AppConfigs.SenderMaxWorkers; i++ {
		go func() {
			for r := range buffers {

				if c.FileCount > int64(c.AppConfigs.DestinationRedisQueueLimit) {
					if c.FileCount-int64(c.AppConfigs.DestinationRedisQueueLimit) > 10 {
						c.CountClear()
						c.IncreaseFileCount(int(c.BackendInst.GetDestQueueSize()))
					} else {
						c.IncreaseFileCount(1)
						logger.Println("destination redis queue if full")
						continue
					}
				}

				if r.Err == nil {
					c.IncreaseFileCount(1)
					c.BackendInst.SendFileContent(r.EncodeContent)
					logger.Println("send file: ", r.Path)
				}
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

// cacheFlow cache file entry in redis
// if file reach the cache expire time then remove it
func (c *Collector) cacheFlow() chan string {

	cacheBuffer := make(chan string, 2)

	var wg sync.WaitGroup
	wg.Add(c.AppConfigs.SenderMaxWorkers)

	for i := 0; i < c.AppConfigs.SenderMaxWorkers; i++ {
		go func() {
			for path := range cacheBuffer {

				if !c.AppConfigs.ReserveFile {
					os.Remove(path)
				}

				if timestamp, errs := c.BackendInst.CacheFileCheck(path); errs != nil {
					c.BackendInst.CacheFileEntry(path, strconv.FormatInt(time.Now().Unix(), 10))
				} else if before, _ := strconv.ParseInt(timestamp, 10, 64); before+int64(c.AppConfigs.FileCacheTimeout) <= time.Now().Unix() {
					os.Remove(path)
					c.BackendInst.RemoveCacheEntry(path)
				}
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
	}()

	return cacheBuffer
}

// GetMatch traverse the filters and check if file should be send
func (c *Collector) GetMatch(filepath string) bool {
	if len(c.filters) > 0 {
		for _, filterFunc := range c.filters {
			if !filterFunc(filepath, c.Rule) {
				return false
			}
		}
	}
	return true
}

// ShutDown close the file collect daemon
func (c *Collector) ShutDown() {
	c.cancleFunc()
}
