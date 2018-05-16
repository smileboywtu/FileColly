// colly define the flow to file collecting
// and cache policy
package colly

import (
	"fmt"
	"sync"
	"os"
	"log"
	"io/ioutil"
	"context"
	"github.com/pkg/errors"
	"github.com/go-redis/redis"
	"github.com/smileboywtu/FileCollector/common"
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

	ctx        context.Context
	cancleFunc context.CancelFunc
}

func InitLogger(logFile string) {

	fd, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open log file error")
	}
	logger = log.New(fd, "collector: ", log.Lshortfile)
}

// NewCollector init a collector to collect file in directory
func NewCollector(opts *AppConfigOption) (*Collector, error) {

	ctx, cancle := context.WithCancel(context.Background())

	// init backend option
	redisOpts := &redis.Options{
		Addr:       fmt.Sprintf("%s:%s", opts.RedisHost, opts.RedisPort),
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

// ListCacheFiles get current cache file from backend
func (c *Collector) ListCacheFiles() []string {
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
func (c *Collector) encodeFlow(fileItems <-chan FileItem, result chan<- EncodeResult) {

	for item := range fileItems {

		if !c.GetMatch(item.FilePath) {
			c.sendPoll(result, EncodeResult{item.FilePath, "", errors.New("file not match")})
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

func (c *Collector) SendFlow() {

	var wg sync.WaitGroup
	result := make(chan EncodeResult)

	fileItems, errc := c.FileWalkerInst.Walk()

	wg.Add(c.AppConfigs.ReaderMaxWorkers)
	for i := 0; i < c.AppConfigs.ReaderMaxWorkers; i++ {
		go func() {
			c.encodeFlow(fileItems, result)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(result)
	}()

	c.cacheFlow(result)

	if err := <-errc; err != nil {
		fmt.Println(err.Error())
		logger.Println(err.Error())
	}

}

// cacheFlow cache current file in pipeline and remove file from directory
// if queue is out of limit size or reserve file is true, then do nothing
// about the file
func (c *Collector) cacheFlow(results <-chan EncodeResult) {

	var wg sync.WaitGroup
	wg.Add(c.AppConfigs.SenderMaxWorkers)
	for i := 0; i < c.AppConfigs.SenderMaxWorkers; i++ {
		go func() {
			for r := range results {
				if false == c.BackendInst.IsAllow() && true == c.AppConfigs.ReserveFile {
					continue
				}
				if r.Err == nil {
					c.BackendInst.SendFileContent(r.EncodeContent)
				}
				os.Remove(r.Path)
			}
			wg.Done()
		}()
	}

	wg.Wait()
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
