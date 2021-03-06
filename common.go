// Package zfs implements basic manipulation of ZFS pools and data sets.
// Use libzfs C library instead CLI zfs tools, with goal
// to let using and manipulating OpenZFS form with in go project.
//
// TODO: Adding to the pool. (Add the given vdevs to the pool)
// TODO: Scan for pools.
//
//
package zfs

/*
#cgo CFLAGS: -g -D_GNU_SOURCE -DHAVE_IOCTL_IN_SYS_IOCTL_H=1 -D__USE_LARGEFILE64=1
//#cgo LDFLAGS: -lzpool -lnvpair
#cgo pkg-config: libzfs
#cgo LDFLAGS: -l:libzpool.a -l:libzfs.a -l:libzfs_core.a -l:libnvpair.a -l:libuutil.a -ludev -lcrypto -lblkid -luuid -lz -lrt -lm -lpthread

#include <stdlib.h>
#include <libzfs.h>
#include "common.h"
#include "zpool.h"
#include "zfs.h"
*/
import "C"

import (
	"sync"
)

// VDevType type of device in the pool
type VDevType string

func init() {
	C.go_libzfs_init()
	return
}

// Types of Virtual Devices
const (
	VDevTypeRoot      VDevType = "root"      // VDevTypeRoot root device in ZFS pool
	VDevTypeMirror             = "mirror"    // VDevTypeMirror mirror device in ZFS pool
	VDevTypeReplacing          = "replacing" // VDevTypeReplacing replacing
	VDevTypeRaidz              = "raidz"     // VDevTypeRaidz RAIDZ device
	VDevTypeDisk               = "disk"      // VDevTypeDisk device is disk
	VDevTypeFile               = "file"      // VDevTypeFile device is file
	VDevTypeMissing            = "missing"   // VDevTypeMissing missing device
	VDevTypeHole               = "hole"      // VDevTypeHole hole
	VDevTypeSpare              = "spare"     // VDevTypeSpare spare device
	VDevTypeLog                = "log"       // VDevTypeLog ZIL device
	VDevTypeL2cache            = "l2cache"   // VDevTypeL2cache cache device (disk)
)

// Prop type to enumerate all different properties suppoerted by ZFS
type DatasetProp int
type PoolProp    int
// PoolStatus type representing status of the pool
type PoolStatus int

// PoolState type representing pool state
type PoolState uint64

// VDevState - vdev states tye
type VDevState uint64

// VDevAux - vdev aux states
type VDevAux uint64

// Property ZFS pool or dataset property value
type PropertyValue struct {
	Value  string	`json:"value"`
	Source string	`json:"source"`
}

var Global struct {
	Mtx sync.Mutex
}

// Pool status
const (
	/*
	 * The following correspond to faults as defined in the (fault.fs.zfs.*)
	 * event namespace.  Each is associated with a corresponding message ID.
	 */
	PoolStatusCorruptCache      PoolStatus = iota /* corrupt /kernel/drv/zpool.cache */
	PoolStatusMissingDevR                         /* missing device with replicas */
	PoolStatusMissingDevNr                        /* missing device with no replicas */
	PoolStatusCorruptLabelR                       /* bad device label with replicas */
	PoolStatusCorruptLabelNr                      /* bad device label with no replicas */
	PoolStatusBadGUIDSum                          /* sum of device guids didn't match */
	PoolStatusCorruptPool                         /* pool metadata is corrupted */
	PoolStatusCorruptData                         /* data errors in user (meta)data */
	PoolStatusFailingDev                          /* device experiencing errors */
	PoolStatusVersionNewer                        /* newer on-disk version */
	PoolStatusHostidMismatch                      /* last accessed by another system */
	PoolStatusHosidActive                         /* currently active on another system */
	PoolStatusHostidRequired                      /* multihost=on and hostid=0 */
	PoolStatusIoFailureWait                       /* failed I/O, failmode 'wait' */
	PoolStatusIoFailureContinue                   /* failed I/O, failmode 'continue' */
	PoolStatusIOFailureMap                        /* ailed MMP, failmode not 'panic' */
	PoolStatusBadLog                              /* cannot read log chain(s) */
	PoolStatusErrata                              /* informational errata available */

	/*
	 * If the pool has unsupported features but can still be opened in
	 * read-only mode, its status is ZPOOL_STATUS_UNSUP_FEAT_WRITE. If the
	 * pool has unsupported features but cannot be opened at all, its
	 * status is ZPOOL_STATUS_UNSUP_FEAT_READ.
	 */
	PoolStatusUnsupFeatRead  /* unsupported features for read */
	PoolStatusUnsupFeatWrite /* unsupported features for write */

	/*
	 * These faults have no corresponding message ID.  At the time we are
	 * checking the status, the original reason for the FMA fault (I/O or
	 * checksum errors) has been lost.
	 */
	PoolStatusFaultedDevR  /* faulted device with replicas */
	PoolStatusFaultedDevNr /* faulted device with no replicas */

	/*
	 * The following are not faults per se, but still an error possibly
	 * requiring administrative attention.  There is no corresponding
	 * message ID.
	 */
	PoolStatusVersionOlder /* older legacy on-disk version */
	PoolStatusFeatDisabled /* supported features are disabled */
	PoolStatusResilvering  /* device being resilvered */
	PoolStatusOfflineDev   /* device online */
	PoolStatusRemovedDev   /* removed device */

	/*
	 * Finally, the following indicates a healthy pool.
	 */
	PoolStatusOk
)

// Possible ZFS pool states
const (
	PoolStateActive            PoolState = iota /* In active use		*/
	PoolStateExported                           /* Explicitly exported		*/
	PoolStateDestroyed                          /* Explicitly destroyed		*/
	PoolStateSpare                              /* Reserved for hot spare use	*/
	PoolStateL2cache                            /* Level 2 ARC device		*/
	PoolStateUninitialized                      /* Internal spa_t state		*/
	PoolStateUnavail                            /* Internal libzfs state	*/
	PoolStatePotentiallyActive                  /* Internal libzfs state	*/
)

// Pool properties. Enumerates available ZFS pool properties. Use it to access
// pool properties either to read or set soecific property.
const (
	PoolPropCont PoolProp = iota - 2
	PoolPropInval
	PoolPropName
	PoolPropSize
	PoolPropCapacity
	PoolPropAltroot
	PoolPropHealth
	PoolPropGUID
	PoolPropVersion
	PoolPropBootfs
	PoolPropDelegation
	PoolPropAutoreplace
	PoolPropCachefile
	PoolPropFailuremode
	PoolPropListsnaps
	PoolPropAutoexpand
	PoolPropDedupditto
	PoolPropDedupratio
	PoolPropFree
	PoolPropAllocated
	PoolPropReadonly
	PoolPropAshift
	PoolPropComment
	PoolPropExpandsz
	PoolPropFreeing
	PoolPropFragmentaion
	PoolPropLeaked
	PoolPropMaxBlockSize
	PoolPropTName
	PoolPropMaxNodeSize
	PoolPropMultiHost
	PoolPropCheckPoint
	PoolPropLoadGUID
	PoolPropAutoTrim
	PoolNumProps
)

/*
 * Dataset properties are identified by these constants and must be added to
 * the end of this list to ensure that external consumers are not affected
 * by the change. If you make any changes to this list, be sure to update
 * the property table in module/zcommon/zfs_prop.c.
 */
const (
	DatasetPropCont DatasetProp = iota - 2
	DatasetPropBad
	DatasetPropType
	DatasetPropCreation
	DatasetPropUsed
	DatasetPropAvailable
	DatasetPropReferenced
	DatasetPropCompressratio
	DatasetPropMounted
	DatasetPropOrigin
	DatasetPropQuota
	DatasetPropReservation
	DatasetPropVolsize
	DatasetPropVolblocksize
	DatasetPropRecordsize
	DatasetPropMountpoint
	DatasetPropSharenfs
	DatasetPropChecksum
	DatasetPropCompression
	DatasetPropAtime
	DatasetPropDevices
	DatasetPropExec
	DatasetPropSetuid
	DatasetPropReadonly
	DatasetPropZoned
	DatasetPropSnapdir
	DatasetPropPrivate /* not exposed to user, temporary */
	DatasetPropAclinherit
	DatasetPropCreateTXG /* not exposed to the user */
	DatasetPropName      /* not exposed to the user */
	DatasetPropCanmount
	DatasetPropIscsioptions /* not exposed to the user */
	DatasetPropXattr
	DatasetPropNumclones /* not exposed to the user */
	DatasetPropCopies
	DatasetPropVersion
	DatasetPropUtf8only
	DatasetPropNormalize
	DatasetPropCase
	DatasetPropVscan
	DatasetPropNbmand
	DatasetPropSharesmb
	DatasetPropRefquota
	DatasetPropRefreservation
	DatasetPropGUID
	DatasetPropPrimarycache
	DatasetPropSecondarycache
	DatasetPropUsedsnap
	DatasetPropUsedds
	DatasetPropUsedchild
	DatasetPropUsedrefreserv
	DatasetPropUseraccounting /* not exposed to the user */
	DatasetPropStmfShareinfo  /* not exposed to the user */
	DatasetPropDeferDestroy
	DatasetPropUserrefs
	DatasetPropLogbias
	DatasetPropUnique   /* not exposed to the user */
	DatasetPropObjsetid /* not exposed to the user */
	DatasetPropDedup
	DatasetPropMlslabel
	DatasetPropSync
	DatasetPropDnodeSize
	DatasetPropRefratio
	DatasetPropWritten
	DatasetPropClones
	DatasetPropLogicalused
	DatasetPropLogicalreferenced
	DatasetPropInconsistent /* not exposed to the user */
	DatasetPropVolmode
	DatasetPropFilesystemLimit
	DatasetPropSnapshotLimit
	DatasetPropFilesystemCount
	DatasetPropSnapshotCount
	DatasetPropSnapdev
	DatasetPropAcltype
	DatasetPropSelinuxContext
	DatasetPropSelinuxFsContext
	DatasetPropSelinuxDefContext
	DatasetPropSelinuxRootContext
	DatasetPropRelatime
	DatasetPropRedundantMetadata
	DatasetPropOverlay
	DatasetPropPrevSnap
	DatasetPropReceiveResumeToken
	DatasetPropEncryption
	DatasetPropKeyLocation
	DatasetPropKeyFormat
	DatasetPropPBKDF2Salt
	DatasetPropPBKDF2Iters
	DatasetPropEncryptionRoot
	DatasetPropKeyGUID
	DatasetPropKeyStatus
	DatasetPropRemapTXG /* not exposed to the user */
	DatasetNumProps
)

// LastError get last underlying libzfs error description if any
func LastError() (err error) {
	return NewError(ErrorCode(C.libzfs_last_error()), C.GoString(C.libzfs_last_error_str()))
}

// ClearLastError force clear of any last error set by undeliying libzfs
func ClearLastError() (err error) {
	err = LastError()
	C.libzfs_clear_last_error()
	return
}

func LastErrorCode() int {
	return int(C.libzfs_last_error())
}

func booleanT(b bool) (r C.boolean_t) {
	if b {
		return 1
	}
	return 0
}

// vdev states are ordered from least to most healthy.
// A vdev that's VDevStateCantOpen or below is considered unusable.
const (
	VDevStateUnknown  VDevState = iota // Uninitialized vdev
	VDevStateClosed                    // Not currently open
	VDevStateOffline                   // Not allowed to open
	VDevStateRemoved                   // Explicitly removed from system
	VDevStateCantOpen                  // Tried to open, but failed
	VDevStateFaulted                   // External request to fault device
	VDevStateDegraded                  // Replicated vdev with unhealthy kids
	VDevStateHealthy                   // Presumed good
)

// vdev aux states.  When a vdev is in the VDevStateCantOpen state, the aux field
// of the vdev stats structure uses these constants to distinguish why.
const (
	VDevAuxNone         VDevAux = iota // no error
	VDevAuxOpenFailed                  // ldi_open_*() or vn_open() failed
	VDevAuxCorruptData                 // bad label or disk contents
	VDevAuxNoReplicas                  // insufficient number of replicas
	VDevAuxBadGUIDSum                  // vdev guid sum doesn't match
	VDevAuxTooSmall                    // vdev size is too small
	VDevAuxBadLabel                    // the label is OK but invalid
	VDevAuxVersionNewer                // on-disk version is too new
	VDevAuxVersionOlder                // on-disk version is too old
	VDevAuxUnsupFeat                   // unsupported features
	VDevAuxSpared                      // hot spare used in another pool
	VDevAuxErrExceeded                 // too many errors
	VDevAuxIOFailure                   // experienced I/O failure
	VDevAuxBadLog                      // cannot read log chain(s)
	VDevAuxExternal                    // external diagnosis
	VDevAuxSplitPool                   // vdev was split off into another pool
)
