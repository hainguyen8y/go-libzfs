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
	"errors"
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
		return errors.New("not snapshot")
	}
	dtpath := C.CString(pathname[:at])
	snapspec := C.CString(pathname[at+1:])
	nvl := C.fnvlist_alloc();
	if nvl == nil {
		return errors.New("not allocated memory")
	}
	defer C.nvlist_free(nvl)

	zhp := C.zfs_open(C.libzfsHandle, dtpath, C.ZFS_TYPE_FILESYSTEM | C.ZFS_TYPE_VOLUME)
	if zhp == nil {
		return errors.New(C.GoString(C.libzfs_last_error_str()))
	}
	defer C.zfs_close(zhp)

	zhpdup := C.zfs_handle_dup(zhp)
	defer C.zfs_close(zhpdup)
	cerr := C.zfs_iter_snapspec(zhpdup, snapspec, (C.zfs_iter_f)(unsafe.Pointer(C.snapshot_to_nvl_cb)), unsafe.Pointer(nvl));
	if cerr != 0 && cerr != C.ENOENT {
		return fmt.Errorf("iter snapspec %d: %s", int(cerr), C.GoString(C.libzfs_last_error_str()))
	}

	if C.nvlist_empty(nvl) == C.B_TRUE {
		return errors.New("could not find any snapshots to destroy; check snapshot names.")
	}
	cerr = C.zfs_destroy_snaps_nvl(C.libzfsHandle, nvl, C.B_TRUE);
	if cerr != 0 {
		return fmt.Errorf("destroy snaps: %s", C.GoString(C.libzfs_last_error_str()))
	}

	return nil
}
