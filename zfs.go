package zfs

import (
	"os"
)

type Version struct {
	Major 		int
    Minor 		int
    Patch 		int
}

type Properties map[DatasetProp]PropertyValue

type DestroyFlags struct {
	IsChildrenRecursive		bool //-r
	IsDependentRecursive	bool //-R
	IsForcedToUnmount		bool //-f
	IsDryRun				bool //-n
	VerboseInfo				bool //-v
}

type CreateFlags struct {
	IsParentCreated bool //-p
	SparseVolume	bool //-s
}

type RollbackFlags struct {
	IsRecursiveDestroy 		bool //-r
	R						bool //-R
	IsForcedToUnmountClones	bool //-f w/o -R
}

type CloneFlags struct {
	IsParentCreated			bool //-p
}

type ListFlags struct {
	IsRecursive				bool 		`json:"recursive"`	//-p
	Depth					int  		`json:"depth"`		//-d
	Types					[]DatasetType 	`json:"types"`		//-t
	SortProperties			[]DatasetProp	`json:"sort"`		//-s
	SortPropertiesDesc		[]DatasetProp	`json:"sort-desc"`	//-S
	Paths					[]string	`json:"paths"`
}

type MountFlags struct {
	IsOverlay				bool		`json:"is_overlay"`	//-O
	OptionalProperties		[]string	`json:"properties"`	//-o
	IsAll					bool		`json:"is_all"`		//-a
}

type SendFlags struct {
	Verbose    bool `json:"verbose"`    //-v
	Replicate  bool `json:"replicate"` 	//-R
	DoAll      bool	`json:"do_all"`		//-I
	FromOrigin bool	`json:"fromorigin"`
	Dedup      bool	`json:"dedup"`		//-D
	Props      bool	`json:"props"`		//-p
	DryRun     bool `json:"dryrun"`		//-n
	Parsable   bool
	Progress   bool
	LargeBlock bool `json:"large_block"` //-L
	EmbedData  bool	`json:"embed_data"` //-e
	Compress   bool	`json:"compress"`	//-c
	Raw		   bool `json:"raw"`		//--raw
	Backup     bool `json:"backup"`		//-b
	Holds	   bool `json:"holds"`		//-h
}

type RecvFlags struct {
	Verbose     bool	`json:"verbose"`		//-v
	IsPrefix    bool	`json:"isprefix"` 		//-d
	IsTail      bool	`json:"istail"` 		//-e
	DryRun      bool	`json:"dryrun"`			//-n
	Force       bool	`json:"force"`			//-r
	CanmountOff bool	`json:"canmountoff"`
	Resumable   bool	`json:"resumable"`		//-s
	ByteSwap    bool	`json:"byteswap"`
	NoMount     bool	`json:"nomount"`		//-u
	Holds		bool	`json:"holds"`
	SkipHolds	bool	`json:"skipholds"`		//-h
	DoMount		bool	`json:"domount"`
}

type IDataset interface {
	Open(path string) (error)
	Close() (error)
	Path() string
	LibraryVersion() (*Version, error)
	KernelModuleVersion() (*Version, error)
	Create(*CreateFlags, map[DatasetProp]PropertyValue) (IDataset, error)
	Destroy(*DestroyFlags) (error)
	CreateSnapshot(recursive bool, properties map[DatasetProp]PropertyValue) ([]IDataset, error)
	CreateBookmark(nam string) (IDataset, error)
	Rollback(*RollbackFlags) (error)
	Clone(map[DatasetProp]PropertyValue) ([]IDataset, error)
	Rename() (error)
	List(*ListFlags) ([]IDataset, error)
	Properties() (map[DatasetProp]PropertyValue, error)
	Mount(*MountFlags) (error)
	Umount(force, isAll bool) (error)
	SendFrom(from string, outf *os.File, flags SendFlags) (error)
	SendSize(from string, flags *SendFlags) (int64, error)
	ReceiveResumeToken() (string, error)
	Receive(inf *os.File, flags *RecvFlags) (int64, error)
	ReceiveResumeAbort() (error)
}
