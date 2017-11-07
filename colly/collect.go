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

	// pool worker
	pool *Pool

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
	bufferLimit int64) *Collector {

	return &Collector{
		Backend:          backend,
		Scanner:          scanner,
		SyncDelay:        syncDelay,
		CacheDelay:       cacheDelay,
		SyncDone:         make(chan bool, 1),
		CacheDone:        make(chan bool, 1),
		BufferCacheLimit: bufferLimit,
		pool:             NewWorkPool(10),
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

type FileItem struct {
	Name string
	*Collector
}

func (f *FileItem) Runner() {
	content, err := ioutil.ReadFile(f.Name)
	if err != nil {
		os.Remove(f.Name)
		return
	}
	index := f.Scanner.TrimRootDirectoryPath(f.Name)
	encoder := &FileEncoder{
		FilePath:    index,
		FileContent: content,
	}
	packBytes, err := encoder.Encode()
	if err != nil {
		return
	}
	f.Backend.CacheFileContent(packBytes)
	log.Println("send file: ", f.Name)

	// Delete file after cache
	os.Remove(f.Name)
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

	var buffer = make([]string, 0, 100)
	var bufferLimit int64 = c.BufferCacheLimit
	var wg = &sync.WaitGroup{}
	var toRemove []string

	for _, file := range files {
		if !c.Backend.IsAllow() {
			break
		}

		toRemove = append(toRemove, file)
		if !c.GetMatch(file) {
			continue
		}
		if fileInfo, err := os.Stat(file); os.IsNotExist(err) {
			logger.Println(err)
			continue
		} else if bufferLimit -= fileInfo.Size(); bufferLimit <= 0 {
			break
		}

		wg.Add(1)
		go func() {
			c.pool.Run(&FileItem{
				Name:      file,
				Collector: c,
			})
			wg.Done()
		}()
	}

	// delete from redis
	c.Backend.RemoveCacheEntry(toRemove)
	wg.Wait()

	// bulk cache file
	c.Backend.BatchCacheFileContent(buffer)
	return nil
}

// Shutdown pool
func (c *Collector) ShutDown() {
	c.pool.ShutDown()
}
