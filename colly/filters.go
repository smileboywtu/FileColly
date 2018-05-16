// Define file walker filters and collector filters
package colly

import (
	"os"
	"strings"
	"time"
)

type Rule struct {
	FileSizeLimit   int64
	AllowEmpty      bool
	ReserveFile     bool
	CollectWaitTime int
}

type FilterFuncs func(filepath string, rule Rule) bool

// FileWalkerGenericFilter check basic condition to collect file
func FileWalkerGenericFilter(filepath string, rule Rule) bool {

	fileMeta, errs := os.Stat(filepath)
	// not exists or permission not allow
	if os.IsNotExist(errs) || os.IsPermission(errs) {
		return false
	}

	// file content is empty
	if rule.AllowEmpty && fileMeta.Size()-0 == 0 {
		return true
	}

	// file size is out of limito
	if fileMeta.Size()-rule.FileSizeLimit > 0 {
		return false
	}

	// shadow file
	if strings.HasPrefix(fileMeta.Name(), ".") {
		return false
	}

	// can't be read now
	if fileMeta.ModTime().Unix()+int64(rule.CollectWaitTime) > time.Now().Unix() {
		return false
	}

	return true
}

// CollectorGenericFilter filter file send to backend
func CollectorGenericFilter(filepath string, rule Rule) bool {
	return true
}
