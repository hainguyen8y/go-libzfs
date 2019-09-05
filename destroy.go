package zfs

// #include <stdlib.h>
// #include <libzfs.h>
// #include "common.h"
// #include "zpool.h"
// #include "zfs.h"
// int snapshot_to_nvl_cb(zfs_handle_t *zhp, void *arg)
// {
// 	int err = 0;
// 	nvlist_t *pnvl = (nvlist_t*) arg;
//
// 	if (nvlist_add_boolean(pnvl, zfs_get_name(zhp))) {
// 		return -1;
// 	}
// 	return (err);
// }
import "C"

import (
	"errors"
	"strings"
	"fmt"
	"unsafe"
)

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

	cerr := C.zfs_iter_snapspec(zhp, snapspec, (C.zfs_iter_f)(unsafe.Pointer(C.snapshot_to_nvl_cb)), unsafe.Pointer(nvl));
	if cerr != C.ENOENT {
		return fmt.Errorf("iter snapspec: %s", C.GoString(C.libzfs_last_error_str()))
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
