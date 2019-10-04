#include <libzfs.h>
#include <memory.h>
#include <string.h>
#include <stdio.h>
#include "common.h"

libzfs_handle_t *g_zfs;

libzfs_handle_t *libzfs_get_handle() {
	return g_zfs;
}

int go_libzfs_init() {
	g_zfs = libzfs_init();
	return 0;
}

int libzfs_last_error() {
	return libzfs_errno(libzfs_get_handle());
}

const char *libzfs_last_error_str() {
	return libzfs_error_description(libzfs_get_handle());
}

int libzfs_clear_last_error() {
	zfs_standard_error(libzfs_get_handle(), EZFS_SUCCESS, "success");
	return 0;
}

property_list_t *new_property_list() {
	property_list_t *r = malloc(sizeof(property_list_t));
	memset(r, 0, sizeof(property_list_t));
	return r;
}

void free_properties(property_list_t *root) {
	if (root != 0) {
		property_list_t *tmp = 0;
		do {
			tmp = root->pnext;
			free(root);
			root = tmp;
		} while(tmp);
	}
}

nvlist_ptr new_property_nvlist() {
	nvlist_ptr props = NULL;
	int r = nvlist_alloc(&props, NV_UNIQUE_NAME, 0);
	if ( r != 0 ) {
		return NULL;
	}
	return props;
}

int property_nvlist_add(nvlist_ptr list, const char *prop, const char *value) {
	return nvlist_add_string(list, prop, value);
}

int redirect_libzfs_stdout(int to) {
	int save, res;
	save = dup(STDOUT_FILENO);
	if (save < 0) {
		return save;
	}
	res = dup2(to, STDOUT_FILENO);
	if (res < 0) {
		return res;
	}
	return save;
}

int restore_libzfs_stdout(int saved) {
	int res;
	fflush(stdout);
	res = dup2(saved, STDOUT_FILENO);
	if (res < 0) {
		return res;
	}
	close(saved);
}

const char *libzfs_strerrno(int errcode) {
	switch (errcode) {
	case EZFS_SUCCESS:
		return "success";
	case EZFS_NOMEM:
		return "out of memory";
	case EZFS_BADPROP:
		return "invalid property value";
	case EZFS_PROPREADONLY:
		return "read-only property";
	case EZFS_PROPTYPE:
		return "property doesn't apply to datasets of this type";
	case EZFS_PROPNONINHERIT:
		return "property cannot be inherited";
	case EZFS_PROPSPACE:
		return "invalid quota or reservation";
	case EZFS_BADTYPE:
		return "operation not applicable to datasets of this type";
	case EZFS_BUSY:
		return "pool or dataset is busy";
	case EZFS_EXISTS:
		return "pool or dataset exists";
	case EZFS_NOENT:
		return "no such pool or dataset";
	case EZFS_BADSTREAM:
		return "invalid backup stream";
	case EZFS_DSREADONLY:
		return "dataset is read-only";
	case EZFS_VOLTOOBIG:
		return "volume size exceeds limit for this system";
	case EZFS_INVALIDNAME:
		return "invalid name";
	case EZFS_BADRESTORE:
		return "unable to restore to destination";
	case EZFS_BADBACKUP:
		return "backup failed";
	case EZFS_BADTARGET:
		return "invalid target vdev";
	case EZFS_NODEVICE:
		return "no such device in pool";
	case EZFS_BADDEV:
		return "invalid device";
	case EZFS_NOREPLICAS:
		return "no valid replicas";
	case EZFS_RESILVERING:
		return "currently resilvering";
	case EZFS_BADVERSION:
		return "unsupported version or feature";
	case EZFS_POOLUNAVAIL:
		return "pool is unavailable";
	case EZFS_DEVOVERFLOW:
		return "too many devices in one vdev";
	case EZFS_BADPATH:
		return "must be an absolute path";
	case EZFS_CROSSTARGET:
		return "operation crosses datasets or pools";
	case EZFS_ZONED:
		return "dataset in use by local zone";
	case EZFS_MOUNTFAILED:
		return "mount failed";
	case EZFS_UMOUNTFAILED:
		return "umount failed";
	case EZFS_UNSHARENFSFAILED:
		return "unshare(1M) failed";
	case EZFS_SHARENFSFAILED:
		return "share(1M) failed";
	case EZFS_UNSHARESMBFAILED:
		return "smb remove share failed";
	case EZFS_SHARESMBFAILED:
		return "smb add share failed";
	case EZFS_PERM:
		return "permission denied";
	case EZFS_NOSPC:
		return "out of space";
	case EZFS_FAULT:
		return "bad address";
	case EZFS_IO:
		return "I/O error";
	case EZFS_INTR:
		return "signal received";
	case EZFS_ISSPARE:
		return "device is reserved as a hot spare";
	case EZFS_INVALCONFIG:
		return "invalid vdev configuration";
	case EZFS_RECURSIVE:
		return "recursive dataset dependency";
	case EZFS_NOHISTORY:
		return "no history available";
	case EZFS_POOLPROPS:
		return "failed to retrieve pool properties";
	case EZFS_POOL_NOTSUP:
		return "operation not supported on this type of pool";
	case EZFS_POOL_INVALARG:
		return "invalid argument for this pool operation";
	case EZFS_NAMETOOLONG:
		return "dataset name is too long";
	case EZFS_OPENFAILED:
		return "open failed";
	case EZFS_NOCAP:
		return "disk capacity information could not be retrieved";
	case EZFS_LABELFAILED:
		return "write of label failed";
	case EZFS_BADWHO:
		return "invalid user/group";
	case EZFS_BADPERM:
		return "invalid permission";
	case EZFS_BADPERMSET:
		return "invalid permission set name";
	case EZFS_NODELEGATION:
		return "delegated administration is disabled on pool";
	case EZFS_BADCACHE:
		return "invalid or missing cache file";
	case EZFS_ISL2CACHE:
		return "device is in use as a cache";
	case EZFS_VDEVNOTSUP:
		return "vdev specification is not supported";
	case EZFS_NOTSUP:
		return "operation not supported on this dataset";
#ifdef EZFS_IOC_NOTSUPPORTED
	case EZFS_IOC_NOTSUPPORTED:
		return "operation not supported by zfs kernel module";
#endif
	case EZFS_ACTIVE_SPARE:
		return "pool has active shared spare device";
	case EZFS_UNPLAYED_LOGS:
		return "log device has unplayed intent logs";
	case EZFS_REFTAG_RELE:
		return "no such tag on this dataset";
	case EZFS_REFTAG_HOLD:
		return "tag already exists on this dataset";
	case EZFS_TAGTOOLONG:
		return "tag too long";
	case EZFS_PIPEFAILED:
		return "pipe create failed";
	case EZFS_THREADCREATEFAILED:
		return "thread create failed";
	case EZFS_POSTSPLIT_ONLINE:
		return "disk was split from this pool into a new one";
	case EZFS_SCRUB_PAUSED:
		return "scrub is paused; use 'zpool scrub' to resume";
	case EZFS_SCRUBBING:
		return "currently scrubbing; use 'zpool scrub -s' to cancel current scrub";
	case EZFS_NO_SCRUB:
		return "there is no active scrub";
	case EZFS_DIFF:
		return "unable to generate diffs";
	case EZFS_DIFFDATA:
		return "invalid diff data";
	case EZFS_POOLREADONLY:
		return "pool is read-only";
#ifdef EZFS_NO_PENDING
	case EZFS_NO_PENDING:
		return "operation is not in progress";
#endif
#ifdef EZFS_CHECKPOINT_EXISTS
	case EZFS_CHECKPOINT_EXISTS:
		return "checkpoint exists";
#endif
#ifdef EZFS_DISCARDING_CHECKPOINT
	case EZFS_DISCARDING_CHECKPOINT:
		return "currently discarding checkpoint";
#endif
#ifdef EZFS_NO_CHECKPOINT
	case EZFS_NO_CHECKPOINT:
		return "checkpoint does not exist";
#endif
#ifdef EZFS_DEVRM_IN_PROGRESS
	case EZFS_DEVRM_IN_PROGRESS:
		return "device removal in progress";
#endif
#ifdef EZFS_VDEV_TOO_BIG
	case EZFS_VDEV_TOO_BIG:
		return "device exceeds supported size";
#endif
#ifdef EZFS_ACTIVE_POOL
	case EZFS_ACTIVE_POOL:
		return "pool is imported on a different host";
#endif
#ifdef EZFS_CRYPTOFAILED
	case EZFS_CRYPTOFAILED:
		return "encryption failure";
#endif
#ifdef EZFS_TOOMANY
	case EZFS_TOOMANY:
		return "argument list too long";
#endif
#ifdef EZFS_INITIALIZING
	case EZFS_INITIALIZING:
		return "currently initializing";
#endif
#ifdef EZFS_NO_INITIALIZE
	case EZFS_NO_INITIALIZE:
		return "there is no active initialization";
#endif
#ifdef EZFS_WRONG_PARENT
	case EZFS_WRONG_PARENT:
		return "invalid parent dataset";
#endif
#ifdef EZFS_TRIMMING
	case EZFS_TRIMMING:
		return "currently trimming";
#endif
#ifdef EZFS_NO_TRIM
	case EZFS_NO_TRIM:
		return "there is no active trim";
#endif
#ifdef EZFS_TRIM_NOTSUP
	case EZFS_TRIM_NOTSUP:
		return "trim operations are not supported by this device";
#endif
#ifdef EZFS_NO_RESILVER_DEFER
	case EZFS_NO_RESILVER_DEFER:
		return "this action requires the resilver_defer feature";
#endif
#ifdef EZFS_EXPORT_IN_PROGRESS
	case EZFS_EXPORT_IN_PROGRESS:
		return ("pool export in progress");
#endif
	case EZFS_UNKNOWN:
		return "unknown error";
	default:
		return "no error";
	}
}
