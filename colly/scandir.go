/**

	scan directory and save all file entry into redis or other backend
	currently only support redis

	Created: 2017/10/16

 */
package colly

import (
	"io/ioutil"
	"path"
	"strings"
	"os"
)

// Traverse all the files inside a directory
type DirScanner struct {
	DirPath    string   `directory to scan`
	folders    []string `internal exchange list holder`
	BufferSize int      `channel buffer size`
	LimitSize  int64
}

// Create new Scanner
func NewDirScanner(path string, bufferSize int, limitSize int64) *DirScanner {

	scaninst := &DirScanner{
		DirPath:    path,
		folders:    make([]string, 0, 50),
		BufferSize: bufferSize,
		LimitSize:  limitSize,
	}

	// add path to folder
	scaninst.folders = append(scaninst.folders, path)
	return scaninst
}

// Clear folders, then just add default
func (s *DirScanner) ClearCache() {
	s.folders = s.folders[:0]
	s.folders = append(s.folders, s.DirPath)
}

// Scan files inside directory use wide but not recursive mode
func (s *DirScanner) Scandir() []string {

	var buffer []string

	for len(s.folders) > 0 {

		_dir := s.folders[len(s.folders)-1]
		s.folders = s.folders[:len(s.folders)-1]

		entries, err := ioutil.ReadDir(_dir)
		if err != nil {
			panic(err)
		}
		for _, entry := range entries {
			if entry.IsDir() {
				s.folders = append(s.folders, path.Join(_dir, entry.Name()))
			} else {
				if len(buffer) >= s.BufferSize {
					return buffer
				} else {
					var file = path.Join(_dir, entry.Name())
					info, err := os.Stat(file)
					if err != nil {
						continue
					}
					if info.Size() > s.LimitSize {
						continue
					}
					buffer = append(buffer, path.Join(_dir, entry.Name()))
				}
			}
		}
	}

	return buffer
}

// Trim root directory path
func (s *DirScanner) TrimRootDirectoryPath(path string) string {
	return strings.TrimPrefix(path, s.DirPath)
}
