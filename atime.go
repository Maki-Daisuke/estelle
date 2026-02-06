//go:build !linux

package estelle

import (
	"os"
	"time"
)

func GetAtime(fi os.FileInfo) time.Time {
	// On non-Linux systems (like Windows), reliable AccessTime access via syscall is complex or unavailable in standard os.FileInfo.
	// Falling back to ModTime as a proxy, which is sufficient for "Approximated LRU" in dev environments.
	return fi.ModTime()
}
