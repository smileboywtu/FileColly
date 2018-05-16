//	Walk directory
//	Created: 2017/10/16

package colly

import (
	"os"
	"sync"
	"strings"
	"context"
	"path/filepath"
	"github.com/pkg/errors"
)

type FileItem struct {
	FilePath  string
	FileSize  int64
	FileIndex string
}

type FileWalker struct {
	sync.RWMutex
	Directory     string
	MaxWalkerSize int
	filters       []FilterFuncs
	Rule          Rule
	Ctx           context.Context
}

// NewDirectoryWorker create new worker to enumerate files in directory
func NewDirectoryWorker(directory string, workers int, rule Rule, ctx context.Context) *FileWalker {
	return &FileWalker{
		Directory:     directory,
		MaxWalkerSize: workers,
		filters:       make([]FilterFuncs, 0, 5),
		Rule:          rule,
		Ctx:           ctx,
	}
}

// OnFilter add file filter to file workers
func (w *FileWalker) OnFilter(callback FilterFuncs) {
	w.Lock()
	w.filters = append(w.filters, callback)
	w.Unlock()
}

func (w *FileWalker) WalkDir(dirName string) (<-chan FileItem, <-chan error) {

	files := make(chan FileItem)
	errc := make(chan error, 1)

	go func() {
		defer close(files)

		errc <- filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			// filters
			for _, filterFunc := range w.filters {
				if !filterFunc(path, w.Rule) {
					return nil
				}
			}

			select {
			case files <- FileItem{
				FilePath:  path,
				FileIndex: w.TrimDirectoryDirectoryPath(path),
				FileSize:  info.Size(),
			}:
			case <-w.Ctx.Done():
				return errors.New("walk canceled")
			}

			return nil
		})
	}()

	return files, errc
}

func (w *FileWalker) Walk() (<-chan FileItem, <-chan error) {
	return w.WalkDir(w.Directory)
}

func (w *FileWalker) TrimDirectoryDirectoryPath(path string) string {
	return strings.TrimPrefix(path, w.Directory)
}
