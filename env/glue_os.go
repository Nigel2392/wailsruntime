//go:build !js && !wasm
// +build !js,!wasm

package env

import "os"

/*
	const (
		O_RDONLY FileFlags = 1
		O_WRONLY FileFlags = 2
		O_RDWR   FileFlags = 4
		O_APPEND FileFlags = 8
		O_CREATE FileFlags = 16
		O_EXCL   FileFlags = 32
		O_SYNC   FileFlags = 64
		O_TRUNC  FileFlags = 128
	)
*/

var syscallMap = map[FileFlags]int{
	1:   os.O_RDONLY,
	2:   os.O_WRONLY,
	4:   os.O_RDWR,
	8:   os.O_APPEND,
	16:  os.O_CREATE,
	32:  os.O_EXCL,
	64:  os.O_SYNC,
	128: os.O_TRUNC,
}

func convertFlags(flags FileFlags) int {
	var newFlags = 0
	for k, v := range syscallMap {
		if flags&k == k {
			newFlags |= v
		}
	}
	return newFlags
}
