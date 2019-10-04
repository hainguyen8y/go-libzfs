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
	"os"
	"unsafe"
)

func SendResume(outf *os.File, flags *SendFlags, resumeToken string) error {
	cflags := to_sendflags_t(flags)
	defer C.free(unsafe.Pointer(cflags))

	cresume_token := C.CString(resumeToken)
	defer C.free(unsafe.Pointer(cresume_token))

	rc := C.zfs_send_resume(C.libzfs_get_handle(), cflags, C.int(outf.Fd()), cresume_token)
	if rc != 0 {
		return LastError()
	}
	return nil
}
