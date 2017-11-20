//	Walk directory
//	Created: 2017/10/16

package colly

import (
	"strings"
	"os"
	"path/filepath"
	"github.com/pkg/errors"
	"context"
)

type FileItem struct {
	FilePath  string
	FileSize  int64
	FileIndex string
}

type FileWalker struct {
	// Root path
	Root string

	// File Limit Size
	FileLimitSize int64

	// Max walk goroutine size
	MaxWalkerSize int

	// done
	Ctx context.Context
}

func NewWalker(root string, limitSize int64, workers int, ctx context.Context) *FileWalker {
	return &FileWalker{
		Root:          root,
		FileLimitSize: limitSize,
		MaxWalkerSize: workers,
		Ctx:           ctx,
	}
}

func (w *FileWalker) WalkDir(dirName string) (<-chan FileItem, <-chan error) {

	files := make(chan FileItem)
	errc := make(chan error, 1)

	go func() {
		// close out chan
		defer close(files)

		errc <- filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			if info.Size() >= w.FileLimitSize {
				return nil
			}

			select {
			case files <- FileItem{
				FilePath:  path,
				FileIndex: w.TrimRootDirectoryPath(path),
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

// Main walk function
func (w *FileWalker) Walk() (<-chan FileItem, <-chan error) {
	// start from root
	return w.WalkDir(w.Root)
}

// Trim root directory path
func (w *FileWalker) TrimRootDirectoryPath(path string) string {
	return strings.TrimPrefix(path, w.Root)
}
