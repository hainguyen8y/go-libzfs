package zfs

import (
	"os"
)

type Version struct {
	Major 		int
    Minor 		int
    Patch 		int
}

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
	Types					[]string 	`json:"types"`		//-t
	SortProperties			[]Property  `json:"sort"`		//-s
	SortPropertiesDesc		[]Property  `json:"sort-desc"`	//-S
}

type MountFlags struct {
	IsOverlay				bool		`json:"is_overlay"`	//-O
	OptionalProperties		[]Property	`json:"properties"`	//-o
	IsAll					bool		`json:"is_all"`		//-a
}

type SendFlags struct {
	Verbose    bool `json:"verbose"`    //-v
	Replicate  bool `json:"replicate"` 	//-R
	DoAll      bool	`json:"do_all"`		//-I
	FromOrigin bool
	Dedup      bool	`json:"dedup"`		//-D
	Props      bool	`json:"props"`		//-p
	DryRun     bool `json:"dryrun"`		//-n
	Parsable   bool
	Progress   bool
	EmbedData  bool	`json:"embed_data"` //-e
	Compress   bool	`json:"compress"`	//-c
	Holds	   bool `json:"holds"`		//-h
	LargeBlock bool `json:"large_block"` //-L
}

type RecvFlags struct {
	Verbose     bool	`json:"verbose"`		//-v
	IsPrefix    bool	`json:"isprefix"` 		//-d
	IsTail      bool	`json:"istail"` 		//-e
	DryRun      bool	`json:"dryrun"`			//-n
	Force       bool	`json:"force"`			//-r
	CanmountOff bool
	Resumable   bool	`json:"resumable"`		//-s
	ByteSwap    bool
	NoMount     bool	`json:"nomount"`		//-u
	SkipHolds	bool	`json:"skipholds"`		//-h
}

type DatasetIf interface {
	Init(path string) (error)
	Deinit() (error)
	LibraryVersion() (*Version, error)
	KernelModuleVersion() (*Version, error)
	Create(*CreateFlags, []Property) (DatasetIf, error)
	Destroy(*DestroyFlags) (error)
	CreateSnapshot(recursive bool, properties []Property) ([]DatasetIf, error)
	Rollback(*RollbackFlags) (error)
	Clone([]Property) ([]DatasetIf, error)
	Rename() (error)
	List(*ListFlags) ([]DatasetIf, error)
	Properties([]Property) ([]Property, error)
	Mount(*MountFlags) (error)
	Umount(force, isAll bool) (error)
	SendFrom(from string, outf *os.File, flags SendFlags) (error)
	SendSize(from string, flags *SendFlags) (int64, error)
	ReceiveResumeToken() (string, error)
	Receive(inf *os.File, flags *RecvFlags) (int64, error)
	ReceiveResumeAbort() (error)
}
