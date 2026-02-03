//go:build !linux

package estelle

import (
	"fmt"
	"os"
	"time"
)

func getAtime(de os.DirEntry) time.Time {
	fi, err := de.Info()
	if err != nil {
		panic(fmt.Errorf("cannot get file info from dirent: %w", err))
	}
	// On non-Linux systems (like Windows), reliable AccessTime access via syscall is complex or unavailable in standard os.FileInfo.
	// Falling back to ModTime as a proxy, which is sufficient for "Approximated LRU" in dev environments.
	return fi.ModTime()
}
