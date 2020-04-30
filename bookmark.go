package zfs

// #include <stdlib.h>
// #include <string.h>
// #include <libzfs.h>
// #include "common.h"
// #include "zpool.h"
// #include "zfs.h"
import "C"

import (
	"strings"
	"syscall"
)

func (self *Dataset) CreateBookmark(name string) (*Dataset, error) {
	sourcePath, _ := self.Path()
	var bookmarkPath string
	if self.Type != DatasetTypeSnapshot && self.Type != DatasetTypeBookmark {
		return nil, NewError(
			int(EInvalidname),
			"invalid source " + sourcePath + " must be snapshot or bookmark",
		)
	} else if self.Type == DatasetTypeSnapshot {
		tmp := strings.Split(sourcePath, "@")
		bookmarkPath = tmp[0] + "#" + name
	} else {
		tmp := strings.Split(sourcePath, "#")
		bookmarkPath = tmp[0] + "#" + name
	}
	nvl := C.fnvlist_alloc();
	C.fnvlist_add_string(nvl, C.CString(bookmarkPath), C.CString(sourcePath))
	var errlist **C.nvlist_t
	_, err := C.lzc_bookmark(nvl, errlist);
	C.fnvlist_free(nvl);
	if err != nil {
		if errno, ok := err.(syscall.Errno); ok {
			return nil, NewError(
				int(errno),
				err.Error(),
			)
		} else {
			return nil, NewError(
				int(-1),
				err.Error(),
			)
		}
	}
	dm, err := DatasetOpen(bookmarkPath)
	return &dm, err
}
