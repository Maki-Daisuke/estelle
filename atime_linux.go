//go:build linux

package estelle

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func GetAtime(fi os.FileInfo) time.Time {
	sys, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		panic(fmt.Errorf("cannot cast to Stat_t: %T", fi.Sys()))
	}
	return time.Unix(sys.Atim.Unix())
}
