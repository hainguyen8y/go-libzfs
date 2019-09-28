package zfs

// #include <stdlib.h>
// #include <libzfs.h>
// #include "common.h"
// #include "zpool.h"
// #include "zfs.h"
// #include <memory.h>
// #include <string.h>
import "C"
import (
	"fmt"
)

func AbortResumable(dtname string) error {
	var namepath *C.char
	namepath = C.CString(dtname +"/%recv")
	cerr := C.zfs_dataset_exists(C.libzfs_get_handle(), namepath, C.ZFS_TYPE_FILESYSTEM | C.ZFS_TYPE_VOLUME)
	if cerr != 0 {
		zhp := C.zfs_open(C.libzfs_get_handle(), namepath, C.ZFS_TYPE_FILESYSTEM | C.ZFS_TYPE_VOLUME);
		if zhp != nil {
			rc := C.zfs_destroy(zhp, C.B_FALSE)
			C.zfs_close(zhp)
			if rc != 0 {
				return LastError()
			}
		}
	} else {
		namepath = C.CString(dtname)
		zhp := C.zfs_open(C.libzfs_get_handle(),
			namepath, C.ZFS_TYPE_FILESYSTEM | C.ZFS_TYPE_VOLUME);
		if zhp == nil {
			return LastError()
		}
		if C.zfs_prop_get_int(zhp, C.ZFS_PROP_INCONSISTENT) == 0 ||
			C.zfs_prop_get(zhp, C.ZFS_PROP_RECEIVE_RESUME_TOKEN,
			nil, 0, nil, nil, 0, C.B_TRUE) == -1 {
			err := NewError(int(C.EZFS_BADPROP), fmt.Sprintf("'%s' does not have any resumable receive state to abort", C.GoString(namepath)))
			C.zfs_close(zhp);
			return err
		}
		rc := C.zfs_destroy(zhp, C.B_FALSE);
		C.zfs_close(zhp);
		if rc != 0 {
			return LastError()
		}
	}
	return nil
}

