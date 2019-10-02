package zfs

// #include <stdio.h>
// #include <stdlib.h>
// #include <libzfs.h>
// #include "common.h"
// #include "zpool.h"
// #include "zfs.h"
// void __printf(char *key, char *val);
// int snapshot_to_nvl_cb(zfs_handle_t *zhp, void *arg);
// void print_list(nvlist_t *pnvl);
import "C"

import (
	"strings"
	"fmt"
	"unsafe"
)

//export __printf
func __printf(k, v *C.char) {
	fmt.Printf("%s=%s\n", C.GoString(k), C.GoString(v))
}

func DestroySnapshot(pathname string) (err error) {
	at := strings.Index(pathname, "@")
	if at == -1 {
		return NewError(int(C.EZFS_BADTYPE), C.GoString(C.libzfs_strerrno(C.EZFS_BADTYPE)))
	}
	dtpath := C.CString(pathname[:at])
	defer C.free(unsafe.Pointer(dtpath))
	snapspec := C.CString(pathname[at+1:])
	defer C.free(unsafe.Pointer(snapspec))
	nvl := C.fnvlist_alloc();
	if nvl == nil {
		return NewError(int(C.EZFS_NOMEM), C.GoString(C.libzfs_strerrno(C.EZFS_NOMEM)))
	}
	defer C.nvlist_free(nvl)

	zhp := C.zfs_open(C.libzfs_get_handle(), dtpath, C.ZFS_TYPE_FILESYSTEM | C.ZFS_TYPE_VOLUME)
	if zhp == nil {
		return LastError()
	}
	defer C.zfs_close(zhp)

	zhpdup := C.zfs_handle_dup(zhp)
	defer C.zfs_close(zhpdup)
	cerr := C.zfs_iter_snapspec(zhpdup, snapspec, (C.zfs_iter_f)(unsafe.Pointer(C.snapshot_to_nvl_cb)), unsafe.Pointer(nvl));
	if cerr != 0 && cerr != C.ENOENT {
		return LastError()
	}

	if C.nvlist_empty(nvl) == C.B_TRUE {
		return NewError(int(C.ENOENT), "could not find any snapshots to destroy; check snapshot names.")
	}
	cerr = C.zfs_destroy_snaps_nvl(C.libzfs_get_handle(), nvl, C.B_TRUE);
	if cerr != 0 {
		return LastError()
	}

	return nil
}
