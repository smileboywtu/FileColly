package colly

import (
	"fmt"
	"sync"
	"os"
	"log"
	"io/ioutil"
	"github.com/pkg/errors"
	"context"
)

var logger *log.Logger

type FilterCallback func(filepath string) bool

type EncodeResult struct {
	Path          string
	EncodeContent string
	Err           error
}

type Collector struct {
	sync.RWMutex
	Backend         CacheWriter
	Walker          *FileWalker
	SingleLimitSize int64
	ParallelReaders int
	ParallelSenders int
	ReserveFlag     bool
	filtercallbacks []FilterCallback
	ctx             context.Context
	cancleFunc      context.CancelFunc
}

type CollectorError struct {
	prob string
}

func (e *CollectorError) Error() string {
	return fmt.Sprintf("%s", e.prob)
}

func InitLogger(logFile string) {

	fd, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open log file error")
	}
	logger = log.New(fd, "collector: ", log.Lshortfile)
}

func NewCollector(
	root string,
	limitSize int64,
	backend CacheWriter,
	readerNumber int,
	senderNumber int,
	reserveFlag bool) *Collector {

	ctx, cancle := context.WithCancel(context.Background())
	c := Collector{
		Backend:         backend,
		filtercallbacks: make([]FilterCallback, 0, 8),
		ParallelReaders: readerNumber,
		ParallelSenders: senderNumber,
		ReserveFlag:     reserveFlag,
		ctx:             ctx,
		cancleFunc:      cancle,
	}
	c.Walker = NewWalker(root, limitSize, 50, ctx)

	return &c
}

func (c *Collector) OnFilter(callback FilterCallback) {
	c.Lock()
	c.filtercallbacks = append(c.filtercallbacks, callback)
	c.Unlock()
}

func (c *Collector) ListCacheFiles() []string {
	if result, err := c.Backend.GetCacheEntry(); err != nil {
		return nil
	} else {
		return result
	}
}

func (c *Collector) SendPoll(result chan<- EncodeResult, item EncodeResult) {
	select {
	case result <- item:
	case <-c.ctx.Done():
		return
	}
}

func (c *Collector) EncodingFile(fileItems <-chan FileItem, result chan<- EncodeResult) {

	for item := range fileItems {

		if !c.GetMatch(item.FilePath) {
			c.SendPoll(result, EncodeResult{item.FilePath, "", errors.New("file not match")})
		}

		data, err := ioutil.ReadFile(item.FilePath)
		if err != nil {
			c.SendPoll(result, EncodeResult{item.FilePath, "", err})
		}
		encoder := &FileEncoder{
			FilePath:    item.FileIndex,
			FileContent: make([]byte, len(data)),
		}
		copy(encoder.FileContent, data)
		packBytes, err := encoder.Encode()
		c.SendPoll(result, EncodeResult{item.FilePath, packBytes, err})
	}

}

func (c *Collector) Sync() {

	var wg sync.WaitGroup
	result := make(chan EncodeResult)

	fileItems, errc := c.Walker.Walk()

	wg.Add(c.ParallelReaders)
	for i := 0; i < c.ParallelReaders; i++ {
		go func() {
			c.EncodingFile(fileItems, result)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(result)
	}()

	c.Cache(result)

	if err := <-errc; err != nil {
		fmt.Println(err.Error())
		logger.Println(err.Error())
	}

}

func (c *Collector) Cache(results <-chan EncodeResult) {

	var wg sync.WaitGroup
	wg.Add(c.ParallelSenders)

	for i := 0; i < c.ParallelSenders; i++ {
		go func() {
			for r := range results {
				if false == c.Backend.IsAllow() && true == c.ReserveFlag {
					continue
				}
				if r.Err == nil {
					c.Backend.CacheFileContent(r.EncodeContent)
				}
				os.Remove(r.Path)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}

func (c *Collector) GetMatch(filepath string) bool {
	if len(c.filtercallbacks) > 0 {
		for _, filter := range c.filtercallbacks {
			if !filter(filepath) {
				return false
			}
		}
	}
	return true
}

func (c *Collector) ShutDown() {
	c.cancleFunc()
}
