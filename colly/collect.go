package colly

import (
	"fmt"
	"time"
	"sync"
	"os"
	"log"
	"io/ioutil"
)

// Log
var logger *log.Logger

// Filter callback design
type FilterCallback func(filepath string) bool

type Collector struct {
	// BACKEND exchange and send file
	Backend CacheWriter

	// DIRECTORY scanner
	Scanner *DirScanner

	// CHECK if sync done
	SyncDone chan bool

	// CHECK if cache done
	CacheDone chan bool

	// SYNC wait delay
	SyncDelay time.Duration

	//CACHE wait delay
	CacheDelay time.Duration

	// Buffer
	BufferCacheLimit int64

	// Max File Size
	MaxFileSize int64

	// FILTERS
	filtercallbacks []FilterCallback

	lock *sync.Mutex
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
	backend CacheWriter,
	scanner *DirScanner,
	syncDelay time.Duration,
	cacheDelay time.Duration,
	bufferLimit int64,
	maxFileSize int64) *Collector {

	return &Collector{
		Backend:          backend,
		Scanner:          scanner,
		SyncDelay:        syncDelay,
		CacheDelay:       cacheDelay,
		SyncDone:         make(chan bool, 1),
		CacheDone:        make(chan bool, 1),
		BufferCacheLimit: bufferLimit,
		MaxFileSize:      maxFileSize,
		filtercallbacks:  make([]FilterCallback, 0, 8),
	}
}

// List all scanned files
func (c *Collector) ListScannedFiles() []string {
	return c.Scanner.Scandir()
}

//	Add filter callbacks
func (c *Collector) OnFilter(callback FilterCallback) {
	c.lock.Lock()

}

//	List Cached file
func (c *Collector) ListCacheFiles() []string {
	if result, err := c.Backend.GetCacheEntry(); err != nil {
		return nil
	} else {
		return result
	}
}

// Cache file content to redis
func (c *Collector) Sync() error {

	defer func() {
		time.Sleep(c.SyncDelay)
		c.SyncDone <- true
	}()

	// scan files
	c.Scanner.ClearCache()
	buffer := c.Scanner.Scandir()

	// write to backend
	return c.Backend.CacheFileEntry(buffer)

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

// Cache file into redis
func (c *Collector) Cache() error {

	defer func() {
		time.Sleep(c.CacheDelay)
		c.CacheDone <- true
	}()

	files, err := c.Backend.GetCacheEntry()
	if err != nil {
		return &CollectorError{prob: "Get Cache Entry failed"}
	}

	var buffer []string
	var bufferLimit int64 = c.BufferCacheLimit

	var wg = &sync.WaitGroup{}
	var wl = &sync.RWMutex{}
	parallel := make(chan bool, 50)
	for _, file := range files {
		if !c.GetMatch(file) {
			continue
		}
		if fileInfo, err := os.Stat(file); os.IsNotExist(err) {
			logger.Println(err)
			continue
		} else if fileInfo.Size() >= c.MaxFileSize {
			continue
		} else if bufferLimit -= fileInfo.Size(); bufferLimit <= 0 {
			break
		}

		wg.Add(1)
		parallel <- true
		go func(filePath string) {

			defer func() {
				<-parallel
				wg.Done()
			}()

			content, err := ioutil.ReadFile(file)
			if err != nil {
				os.Remove(file)
				return
			}
			index := c.Scanner.TrimRootDirectoryPath(file)
			encoder := &FileEncoder{
				FilePath:    index,
				FileContent: content,
			}
			packBytes, err := encoder.Encode()
			if err != nil {
				return
			}

			wl.Lock()
			buffer = append(buffer, packBytes)
			wl.Unlock()

			// Delete file after cache
			os.Remove(file)

		}(file)
	}
	wg.Wait()
	// bulk cache file
	c.Backend.BatchCacheFileContent(buffer)

	return nil
}
