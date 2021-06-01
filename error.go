package zfs

type ErrorCode int

const (
	EUndefined 	ErrorCode = -1
	ESuccess              = 0          /* no error -- success */
)

const (
	ENomem      ErrorCode = iota + 2000 /* out of memory */
	EBadprop                           /* invalid property value */
	EPropreadonly                      /* cannot set readonly property */
	EProptype                          /* property does not apply to dataset type */
	EPropnoninherit                    /* property is not inheritable */
	EPropspace                         /* bad quota or reservation */
	EBadtype                           /* dataset is not of appropriate type */
	EBusy                              /* pool or dataset is busy */
	EExists                            /* pool or dataset already exists */
	ENoent                             /* no such pool or dataset */
	EBadstream                         /* bad backup stream */
	EDsreadonly                        /* dataset is readonly */
	EVoltoobig                         /* volume is too large for 32-bit system */
	EInvalidname                       /* invalid dataset name */
	EBadrestore                        /* unable to restore to destination */
	EBadbackup                         /* backup failed */
	EBadtarget                         /* bad attach/detach/replace target */
	ENodevice                          /* no such device in pool */
	EBaddev                            /* invalid device to add */
	ENoreplicas                        /* no valid replicas */
	EResilvering                       /* currently resilvering */
	EBadversion                        /* unsupported version */
	EPoolunavail                       /* pool is currently unavailable */
	EDevoverflow                       /* too many devices in one vdev */
	EBadpath                           /* must be an absolute path */
	ECrosstarget                       /* rename or clone across pool or dataset */
	EZoned                             /* used improperly in local zone */
	EMountfailed                       /* failed to mount dataset */
	EUmountfailed                      /* failed to unmount dataset */
	EUnsharenfsfailed                  /* unshare(1M) failed */
	ESharenfsfailed                    /* share(1M) failed */
	EPerm                              /* permission denied */
	ENospc                             /* out of space */
	EFault                             /* bad address */
	EIo                                /* I/O error */
	EIntr                              /* signal received */
	EIsspare                           /* device is a hot spare */
	EInvalconfig                       /* invalid vdev configuration */
	ERecursive                         /* recursive dependency */
	ENohistory                         /* no history object */
	EPoolprops                         /* couldn't retrieve pool props */
	EPoolNotsup                        /* ops not supported for this type of pool */
	EPoolInvalarg                      /* invalid argument for this pool operation */
	ENametoolong                       /* dataset name is too long */
	EOpenfailed                        /* open of device failed */
	ENocap                             /* couldn't get capacity */
	ELabelfailed                       /* write of label failed */
	EBadwho                            /* invalid permission who */
	EBadperm                           /* invalid permission */
	EBadpermset                        /* invalid permission set name */
	ENodelegation                      /* delegated administration is disabled */
	EUnsharesmbfailed                  /* failed to unshare over smb */
	ESharesmbfailed                    /* failed to share over smb */
	EBadcache                          /* bad cache file */
	EIsl2CACHE                         /* device is for the level 2 ARC */
	EVdevnotsup                        /* unsupported vdev type */
	ENotsup                            /* ops not supported on this dataset */
	EActiveSpare                       /* pool has active shared spare devices */
	EUnplayedLogs                      /* log device has unplayed logs */
	EReftagRele                        /* snapshot release: tag not found */
	EReftagHold                        /* snapshot hold: tag already exists */
	ETagtoolong                        /* snapshot hold/rele: tag too long */
	EPipefailed                        /* pipe create failed */
	EThreadcreatefailed                /* thread create failed */
	EPostsplitOnline                   /* onlining a disk after splitting it */
	EScrubbing                         /* currently scrubbing */
	ENoScrub                           /* no active scrub */
	EDiff                              /* general failure of zfs diff */
	EDiffdata                          /* bad zfs diff data */
	EPoolreadonly                      /* pool is in read-only mode */
	EUnknown
)

type Error struct {
	code 	ErrorCode
	message	string
}

func (self *Error) Error() string {
	return self.message
}

func (self *Error) ErrorCode() ErrorCode {
	return self.code
}

func NewError(errcode ErrorCode, msg string) error {
	return &Error{
		code: errcode,
		message: msg,
	}
}

