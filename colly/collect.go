package colly

import (
	"fmt"
	"sync"
	"os"
	"log"
	"io/ioutil"
	"github.com/pkg/errors"
)

// Log
var logger *log.Logger

// Filter callback design
type FilterCallback func(filepath string) bool

type EncodeResult struct {
	Path          string
	EncodeContent string
	Err           error
}

type Collector struct {
	sync.RWMutex

	// Root
	// BACKEND exchange and send file
	Backend CacheWriter

	// DIRECTORY scanner
	Walker *FileWalker

	// File limit size
	SingleLimitSize int64

	// parallel file reader
	ParallelReaders int

	// parallel file sender
	ParallelSenders int

	// FILTERS
	filtercallbacks []FilterCallback

	// Cancellation
	done chan struct{}
}

// Error handler
type CollectorError struct {
	prob string
}

func (e *CollectorError) Error() string {
	return fmt.Sprintf("%s", e.prob)
}

// Init collector log
func InitLogger(logFile string) {

	fd, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open log file error")
	}
	logger = log.New(fd, "collector: ", log.Lshortfile)
}

// Collector
func NewCollector(
	root string,
	limitSize int64,
	backend CacheWriter,
	readerNumber int,
	senderNumber int) *Collector {
	c := Collector{
		Backend:         backend,
		filtercallbacks: make([]FilterCallback, 0, 8),
		ParallelReaders: readerNumber,
		ParallelSenders: senderNumber,
		done:            make(chan struct{}),
	}
	c.Walker = NewWalker(root, limitSize, 50, c.done)
	return &c
}

//	Add filter callbacks
func (c *Collector) OnFilter(callback FilterCallback) {
	c.Lock()
	c.filtercallbacks = append(c.filtercallbacks, callback)
	c.Unlock()
}

//	List Cached file
func (c *Collector) ListCacheFiles() []string {
	if result, err := c.Backend.GetCacheEntry(); err != nil {
		return nil
	} else {
		return result
	}
}

// Send wait
func (c *Collector) SendPoll(result chan<- EncodeResult, item EncodeResult) {
	select {
	case result <- item:
	case <-c.done:
		return
	}
}

// Read file and encode
func (c *Collector) EncodingFile(fileItems <-chan FileItem, result chan<- EncodeResult) {

	for item := range fileItems {

		// check if need to deal
		if !c.GetMatch(item.FilePath) {
			c.SendPoll(result, EncodeResult{item.FilePath, "", errors.New("file not match")})
		}

		data, err := ioutil.ReadFile(item.FilePath)
		if err != nil {
			c.SendPoll(result, EncodeResult{item.FilePath, "", err})
		}

		encoder := &FileEncoder{
			FilePath:    item.FileIndex,
			FileContent: data,
		}
		packBytes, err := encoder.Encode()
		c.SendPoll(result, EncodeResult{item.FilePath, packBytes, err})
	}

}

// Collector sync to redis
func (c *Collector) Sync() {

	var wg sync.WaitGroup
	result := make(chan EncodeResult)

	// walk file
	fileItems, errc := c.Walker.Walk()

	// add wait group
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

// Cache pool
func (c *Collector) Cache(results <-chan EncodeResult) {

	var wg sync.WaitGroup
	wg.Add(c.ParallelSenders)

	for i := 0; i < c.ParallelSenders; i++ {
		go func() {
			for r := range results {
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

// Use filter callback to filter files
func (c *Collector) GetMatch(filepath string) bool {
	if len(c.filtercallbacks) > 0 {
		for _, filter := range c.filtercallbacks {
			// do not done for this file
			if !filter(filepath) {
				return false
			}
		}
	}
	return true
}

// cancellation
func (c *Collector) ShutDown() {
	close(c.done)
}
