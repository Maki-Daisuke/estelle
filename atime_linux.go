//go:build linux

package estelle

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func getAtime(de os.DirEntry) time.Time {
	fi, err := de.Info()
	if err != nil {
		panic(fmt.Errorf("cannot get file info from dirent: %w", err))
	}
	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		panic(fmt.Errorf("cannot cast to Stat_t: %T", fi.Sys()))
	}
	return time.Unix(sys.Atim.Unix())
}
