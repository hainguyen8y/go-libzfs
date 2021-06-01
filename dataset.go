package zfs

// #include <stdlib.h>
// #include <libzfs.h>
// #include "common.h"
// #include "zpool.h"
// #include "zfs.h"
import "C"

import (
	"fmt"
	"path"
	"sort"
	"strings"
	"encoding/json"
	"time"
	"unsafe"
)

const (
	msgDatasetIsNil = "Dataset handle not initialized or its closed"
)

// DatasetProperties type is map of dataset or volume properties prop -> value
type DatasetProperties map[DatasetProp]PropertyValue

// DatasetType defines enum of dataset types
type DatasetType int32

const (
	// DatasetTypeFilesystem - file system dataset
	DatasetTypeFilesystem DatasetType = C.ZFS_TYPE_FILESYSTEM
	// DatasetTypeSnapshot - snapshot of dataset
	DatasetTypeSnapshot = C.ZFS_TYPE_SNAPSHOT
	// DatasetTypeVolume - volume (virtual block device) dataset
	DatasetTypeVolume = C.ZFS_TYPE_VOLUME
	// DatasetTypePool - pool dataset
	DatasetTypePool = C.ZFS_TYPE_POOL
	// DatasetTypeBookmark - bookmark dataset
	DatasetTypeBookmark = C.ZFS_TYPE_BOOKMARK
)

var stringToDatasetType = make(map[string]DatasetType)

func init() {
	types := [...]DatasetType{
		DatasetTypeFilesystem,
		DatasetTypeSnapshot,
		DatasetTypeVolume,
		DatasetTypePool,
		DatasetTypeBookmark,
	}

	for _, t := range types {
		stringToDatasetType[t.String()] = t
	}
}

func (t DatasetType) String() string {
	return C.GoString(C.zfs_type_to_name((C.zfs_type_t)(t)))
}

func (t DatasetType) MarshalJSON() ([]byte, error) {
	s := t.String()
	return json.Marshal(s)
}

func (t *DatasetType) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	dstype, ok := stringToDatasetType[j]
	if !ok {
		return fmt.Errorf("type \"%s\" not exists", j)
	}
	*t = dstype
	return err
}

// HoldTag - user holds  tags
type HoldTag struct {
	Name      string
	Timestamp time.Time
}

// Dataset - ZFS dataset object
type Dataset struct {
	list       C.dataset_list_ptr
	Type       DatasetType			`json:"type,omitempty"`
	Properties DatasetProperties	`json:"properties,omitempty"`
	Children   []Dataset			`json:"-"`
}

func NewDataset() (*Dataset) {
	return &Dataset{
		Properties: make(DatasetProperties),
	}
}

func (p Dataset) String() string {
	data, err := json.Marshal(&p)
	if err != nil {
		return ""
	}
	return string(data)
}

func (d *Dataset) openChildren() (err error) {
	d.Children = make([]Dataset, 0, 5)
	list := C.dataset_list_children(d.list)
	for list != nil {
		dataset := Dataset{list: list}
		dataset.Type = DatasetType(C.dataset_type(list))
		dataset.Properties = make(map[DatasetProp]PropertyValue)
		err = dataset.ReloadProperties()
		if err != nil {
			return
		}
		d.Children = append(d.Children, dataset)
		list = C.dataset_next(list)
	}
	for ci := range d.Children {
		if err = d.Children[ci].openChildren(); err != nil {
			return
		}
	}
	return
}

// DatasetOpenAll recursive get handles to all available datasets on system
// (file-systems, volumes or snapshots).
func DatasetOpenAll() (datasets []Dataset, err error) {
	var dataset Dataset
	dataset.list = C.dataset_list_root()
	for dataset.list != nil {
		dataset.Type = DatasetType(C.dataset_type(dataset.list))
		err = dataset.ReloadProperties()
		if err != nil {
			return
		}
		datasets = append(datasets, dataset)
		dataset.list = C.dataset_next(dataset.list)
	}
	for ci := range datasets {
		if err = datasets[ci].openChildren(); err != nil {
			return
		}
	}
	return
}

// DatasetCloseAll close all datasets in slice and all of its recursive
// children datasets
func DatasetCloseAll(datasets []Dataset) {
	for _, d := range datasets {
		d.Close()
	}
}

// DatasetOpen open dataset and all of its recursive children datasets
func DatasetOpen(path string) (d Dataset, err error) {
	if d, err = DatasetOpenSingle(path); err != nil {
		return
	}
	err = d.openChildren()
	return
}

// DatasetOpenSingle open dataset without opening all of its recursive
// children datasets
func DatasetOpenSingle(path string) (d Dataset, err error) {
	csPath := C.CString(path)
	d.list = C.dataset_open(csPath)
	C.free(unsafe.Pointer(csPath))

	if d.list == nil || d.list.zh == nil {
		err = LastError()
		if err == nil {
			err = NewError(ENoent, "dataset not found.")
		}
		err = NewError(err.(*Error).ErrorCode(), fmt.Sprintf("%s - %s", err.Error(), path))
		return
	}
	d.Type = DatasetType(C.dataset_type(d.list))
	d.Properties = make(map[DatasetProp]PropertyValue)
	err = d.ReloadProperties()
	if err != nil {
		return
	}
	return
}

func datasetPropertiesTonvlist(props map[DatasetProp]PropertyValue) (
	cprops C.nvlist_ptr, err error) {
	// convert properties to nvlist C type
	cprops = C.new_property_nvlist()
	if cprops == nil {
		err = NewError(ENomem, "Failed to allocate properties")
		return
	}
	for prop, value := range props {
		csValue := C.CString(value.Value)
		r := C.property_nvlist_add(
			cprops, C.zfs_prop_to_name(C.zfs_prop_t(prop)), csValue)
		C.free(unsafe.Pointer(csValue))
		if r != 0 {
			err = LastError()
			err = NewError(err.(*Error).ErrorCode(), fmt.Sprintf("failed to convert property: %s", err.Error()))
			return
		}
	}
	return
}

// DatasetCreate create a new filesystem or volume on path representing
// pool/dataset or pool/parent/dataset
func DatasetCreate(path string, dtype DatasetType,
	props map[DatasetProp]PropertyValue) (d Dataset, err error) {
	var cprops C.nvlist_ptr
	if cprops, err = datasetPropertiesTonvlist(props); err != nil {
		return
	}
	defer C.nvlist_free(cprops)

	csPath := C.CString(path)
	errcode := C.dataset_create(csPath, C.zfs_type_t(dtype), cprops)
	C.free(unsafe.Pointer(csPath))
	if errcode != 0 {
		err = LastError()
		return
	}
	return DatasetOpen(path)
}

// Close close dataset and all its recursive children datasets (close handle
// and cleanup dataset object/s from memory)
func (d *Dataset) Close() {
	// path, _ := d.Path()
	Global.Mtx.Lock()
	C.dataset_list_close(d.list)
	d.list = nil
	Global.Mtx.Unlock()
	for _, cd := range d.Children {
		cd.Close()
	}
}

// reOpen - close and open dataset. Not thread safe!
func (d *Dataset) reOpen() (err error) {
	d.Close()
	*d, err = DatasetOpen(d.Properties[DatasetPropName].Value)
	return
}

// Destroy destroys the dataset.  The caller must make sure that the filesystem
// isn't mounted, and that there are no active dependents. Set Defer argument
// to true to defer destruction for when dataset is not in use. Call Close() to
// cleanup memory.
func (d *Dataset) Destroy(Defer bool) (err error) {
	if len(d.Children) > 0 {
		path, err := d.Path()
		if err != nil {
			return err
		}
		dsType, err := d.GetProperty(DatasetPropType)
		if err != nil {
			dsType.Value = err.Error() // just put error (why it didn't fetch property type)
		}
		err = NewError(EUndefined, fmt.Sprintf("cannot destroy dataset " + path +
			": " + dsType.Value + " has children"))
		return err
	}
	if d.list != nil {
		if ec := C.dataset_destroy(d.list, booleanT(Defer)); ec != 0 {
			err = LastError()
		}
	} else {
		err = NewError(EUndefined, msgDatasetIsNil)
	}
	return
}

// IsSnapshot - retrun true if datset is snapshot
func (d *Dataset) IsSnapshot() (ok bool) {
	path := d.Properties[DatasetPropName].Value
	ok = (d.Type == DatasetTypeSnapshot || strings.Contains(path, "@"))
	return
}

// DestroyRecursive recursively destroy children of dataset and dataset.
func (d *Dataset) DestroyRecursive() (err error) {
	var path string
	if path, err = d.Path(); err != nil {
		return
	}
	if !strings.Contains(path, "@") { // not snapshot
		if len(d.Children) > 0 {
			for _, c := range d.Children {
				if err = c.DestroyRecursive(); err != nil {
					return
				}
				// close handle to destroyed child dataset
				c.Close()
			}
			// clear closed children array
			d.Children = make([]Dataset, 0)
		}
		err = d.Destroy(false)
	} else {
		var parent Dataset
		tmp := strings.Split(path, "@")
		ppath, snapname := tmp[0], tmp[1]
		if parent, err = DatasetOpen(ppath); err != nil {
			return
		}
		defer parent.Close()
		if len(parent.Children) > 0 {
			for _, c := range parent.Children {
				if path, err = c.Path(); err != nil {
					return
				}
				if strings.Contains(path, "@") {
					continue // skip other snapshots
				}
				if c, err = DatasetOpen(path + "@" + snapname); err != nil {
					continue
				}
				if err = c.DestroyRecursive(); err != nil {
					c.Close()
					return
				}
				c.Close()
			}
		}
		err = d.Destroy(false)
	}
	return
}

// Pool returns pool dataset belongs to
func (d *Dataset) Pool() (p Pool, err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	p.list = C.dataset_get_pool(d.list)
	if p.list != nil && p.list.zph != nil {
		err = p.ReloadProperties()
		return
	}
	err = LastError()
	return
}

// PoolName - return name of the pool
func (d *Dataset) PoolName() string {
	path := d.Properties[DatasetPropName].Value
	i := strings.Index(path, "/")
	if i < 0 {
		return path
	}
	return path[0:i]
}

// ReloadProperties re-read dataset's properties
func (d *Dataset) ReloadProperties() (err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	d.Properties = make(map[DatasetProp]PropertyValue)
	Global.Mtx.Lock()
	defer Global.Mtx.Unlock()

	for prop := DatasetPropType; prop < zfsMaxDatasetProp; prop++ {
		plist := C.read_dataset_property(d.list, C.int(prop))
		if plist == nil {
			continue
		}
		d.Properties[prop] = PropertyValue{Value: C.GoString(&(*plist).value[0]),
			Source: C.GoString(&(*plist).source[0])}
		C.free_properties(plist)
	}
	return
}

// GetProperty reload and return single specified property. This also reloads requested
// property in Properties map.
func (d *Dataset) GetProperty(p DatasetProp) (prop PropertyValue, err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	Global.Mtx.Lock()
	defer Global.Mtx.Unlock()
	plist := C.read_dataset_property(d.list, C.int(p))
	if plist == nil {
		err = LastError()
		return
	}
	defer C.free_properties(plist)
	prop = PropertyValue{Value: C.GoString(&(*plist).value[0]),
		Source: C.GoString(&(*plist).source[0])}
	d.Properties[p] = prop
	return
}

// GetUserProperty - lookup and return user propery
func (d *Dataset) GetUserProperty(p string) (prop PropertyValue, err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	csp := C.CString(p)
	defer C.free(unsafe.Pointer(csp))
	plist := C.read_user_property(d.list, csp)
	if plist == nil {
		err = LastError()
		return
	}
	defer C.free_properties(plist)
	prop = PropertyValue{Value: C.GoString(&(*plist).value[0]),
		Source: C.GoString(&(*plist).source[0])}
	return
}

// SetProperty set ZFS dataset property to value. Not all properties can be set,
// some can be set only at creation time and some are read only.
// Always check if returned error and its description.
func (d *Dataset) SetProperty(p DatasetProp, value string) (err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	csValue := C.CString(value)
	errcode := C.dataset_prop_set(d.list, C.zfs_prop_t(p), csValue)
	C.free(unsafe.Pointer(csValue))
	if errcode != 0 {
		err = LastError()
	}
	// Update Properties member with change made
	if _, err = d.GetProperty(p); err != nil {
		return
	}
	return
}

// SetUserProperty -
func (d *Dataset) SetUserProperty(prop, value string) (err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	csValue := C.CString(value)
	csProp := C.CString(prop)
	errcode := C.dataset_user_prop_set(d.list, csProp, csValue)
	C.free(unsafe.Pointer(csValue))
	C.free(unsafe.Pointer(csProp))
	if errcode != 0 {
		err = LastError()
	}
	return
}

// Clone - clones the dataset.  The target must be of the same type as
// the source.
func (d *Dataset) Clone(target string, props map[DatasetProp]PropertyValue) (rd Dataset, err error) {
	var cprops C.nvlist_ptr
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	if cprops, err = datasetPropertiesTonvlist(props); err != nil {
		return
	}
	defer C.nvlist_free(cprops)
	csTarget := C.CString(target)
	defer C.free(unsafe.Pointer(csTarget))
	if errc := C.dataset_clone(d.list, csTarget, cprops); errc != 0 {
		err = LastError()
		return
	}
	rd, err = DatasetOpen(target)
	return
}

// DatasetSnapshot create dataset snapshot. Set recur to true to snapshot child datasets.
func DatasetSnapshot(path string, recur bool, props map[DatasetProp]PropertyValue) (rd Dataset, err error) {
	var cprops C.nvlist_ptr
	if cprops, err = datasetPropertiesTonvlist(props); err != nil {
		return
	}
	defer C.nvlist_free(cprops)
	csPath := C.CString(path)
	defer C.free(unsafe.Pointer(csPath))
	if errc := C.dataset_snapshot(csPath, booleanT(recur), cprops); errc != 0 {
		err = LastError()
		return
	}
	rd, err = DatasetOpen(path)
	return
}

// Path return zfs dataset path/name
func (d *Dataset) Path() (path string, err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	name := C.dataset_get_name(d.list)
	path = C.GoString(name)
	return
}

// Rollback rollabck's dataset snapshot
func (d *Dataset) Rollback(snap *Dataset, force bool) (err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	if errc := C.dataset_rollback(d.list, snap.list, booleanT(force)); errc != 0 {
		err = LastError()
		return
	}
	d.ReloadProperties()
	return
}

// Promote promotes dataset clone
func (d *Dataset) Promote() (err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	if errc := C.dataset_promote(d.list); errc != 0 {
		err = LastError()
		return
	}
	d.ReloadProperties()
	return
}

// Rename dataset
func (d *Dataset) Rename(newName string, recur,
	forceUnmount bool) (err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	csNewName := C.CString(newName)
	defer C.free(unsafe.Pointer(csNewName))
	if errc := C.dataset_rename(d.list, csNewName,
		booleanT(recur), booleanT(forceUnmount)); errc != 0 {
		err = LastError()
		return
	}
	d.ReloadProperties()
	return
}

// IsMounted checks to see if the mount is active.  If the filesystem is mounted,
// sets in 'where' argument the current mountpoint, and returns true.  Otherwise,
// returns false.
func (d *Dataset) IsMounted() (mounted bool, where string) {
	if d.list == nil {
		return
	}
	Global.Mtx.Lock()
	defer Global.Mtx.Unlock()
	mp := C.dataset_is_mounted(d.list)
	// defer C.free(mp)
	if mounted = (mp != nil); mounted {
		where = C.GoString(mp)
		C.free(unsafe.Pointer(mp))
	}
	return
}

// Mount the given filesystem.
func (d *Dataset) Mount(options string, flags int) (err error) {
	Global.Mtx.Lock()
	defer Global.Mtx.Unlock()
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	csOptions := C.CString(options)
	defer C.free(unsafe.Pointer(csOptions))
	if ec := C.dataset_mount(d.list, csOptions, C.int(flags)); ec != 0 {
		err = LastError()
	}
	return
}

// Unmount the given filesystem.
func (d *Dataset) Unmount(flags int) (err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	if ec := C.dataset_unmount(d.list, C.int(flags)); ec != 0 {
		err = LastError()
	}
	return
}

// UnmountAll unmount this filesystem and any children inheriting the
// mountpoint property.
func (d *Dataset) UnmountAll(flags int) (err error) {
	if d.list == nil {
		err = NewError(EUndefined, msgDatasetIsNil)
		return
	}
	// This is implemented recursive because zfs_unmountall() didn't work
	if len(d.Children) > 0 {
		for _, c := range d.Children {
			if err = c.UnmountAll(flags); err != nil {
				return
			}
		}
	}
	return d.Unmount(flags)
}

// Hold - Adds a single reference, named with the tag argument, to the snapshot.
// Each snapshot has its own tag namespace, and tags must be unique within that space.
func (d *Dataset) Hold(flag string) (err error) {
	var path string
	var pd Dataset
	if path, err = d.Path(); err != nil {
		return
	}
	if !strings.Contains(path, "@") {
		err = NewError(EUndefined, fmt.Sprintf("'%s' is not a snapshot", path))
		return
	}
	pd, err = DatasetOpenSingle(path[:strings.Index(path, "@")])
	if err != nil {
		return
	}
	defer pd.Close()
	csSnapName := C.CString(path[strings.Index(path, "@")+1:])
	defer C.free(unsafe.Pointer(csSnapName))
	csFlag := C.CString(flag)
	defer C.free(unsafe.Pointer(csFlag))
	if 0 != C.zfs_hold(pd.list.zh, csSnapName, csFlag, booleanT(false), -1) {
		err = LastError()
	}
	return
}

// Release - Removes a single reference, named with the tag argument, from the specified snapshot.
// The tag must already exist for each snapshot.  If a hold exists on a snapshot, attempts to destroy
//  that snapshot by using the zfs destroy command return EBUSY.
func (d *Dataset) Release(flag string) (err error) {
	var path string
	var pd Dataset
	if path, err = d.Path(); err != nil {
		return
	}
	if !strings.Contains(path, "@") {
		err = NewError(EUndefined, fmt.Sprintf("'%s' is not a snapshot", path))
		return
	}
	pd, err = DatasetOpenSingle(path[:strings.Index(path, "@")])
	if err != nil {
		return
	}
	defer pd.Close()
	csSnapName := C.CString(path[strings.Index(path, "@")+1:])
	defer C.free(unsafe.Pointer(csSnapName))
	csFlag := C.CString(flag)
	defer C.free(unsafe.Pointer(csFlag))
	if 0 != C.zfs_release(pd.list.zh, csSnapName, csFlag, booleanT(false)) {
		err = LastError()
	}
	return
}

// Holds - Lists all existing user references for the given snapshot
func (d *Dataset) Holds() (tags []HoldTag, err error) {
	var nvl *C.nvlist_t
	var nvp *C.nvpair_t
	var tu64 C.uint64_t
	var path string
	if path, err = d.Path(); err != nil {
		return
	}
	if !strings.Contains(path, "@") {
		err = NewError(EUndefined, fmt.Sprintf("'%s' is not a snapshot", path))
		return
	}
	if 0 != C.zfs_get_holds(d.list.zh, &nvl) {
		err = LastError()
		return
	}
	defer C.nvlist_free(nvl)
	tags = make([]HoldTag, 0, 5)
	for nvp = C.nvlist_next_nvpair(nvl, nvp); nvp != nil; {
		tag := C.nvpair_name(nvp)
		C.nvpair_value_uint64(nvp, &tu64)
		tags = append(tags, HoldTag{
			Name:      C.GoString(tag),
			Timestamp: time.Unix(int64(tu64), 0),
		})

		nvp = C.nvlist_next_nvpair(nvl, nvp)
	}
	return
}

// DatasetPropertyToName convert property to name
// ( returns built in string representation of property name).
// This is optional, you can represent each property with string
// name of choice.
func DatasetPropertyToName(p DatasetProp) (name string) {
	if p == zfsMaxDatasetProp {
		return "numofprops"
	}
	prop := C.zfs_prop_t(p)
	name = C.GoString(C.zfs_prop_to_name(prop))
	return
}

// DestroyPromote - Same as DestroyRecursive() except it will not destroy
// any dependent clones, but promote them first.
// This function will navigate any dependency chain
// of cloned datasets using breadth first search to promote according and let
// you remove dataset regardless of its cloned dependencies.
// Note: that this function wan't work when you want to destroy snapshot this way.
// However it will destroy all snaphsot of destroyed dataset without dependencies,
// otherwise snapshot will move to promoted clone
func (d *Dataset) DestroyPromote() (err error) {
	var snaps []Dataset
	var clones []string
	// We need to save list of child snapshots, to destroy them latter
	// since  they will be moved to promoted clone
	var psnaps []string
	if clones, err = d.Clones(); err != nil {
		return
	}
	if len(clones) > 0 {
		var cds Dataset
		// For this to always work we need to promote youngest clone
		// in terms of most recent origin snapshot or creation time if
		// cloned from same snapshot
		if cds, err = DatasetOpen(clones[0]); err != nil {
			return
		}
		defer cds.Close()
		// since promote will move the snapshots to promoted dataset
		// we need to check and resolve possible name conflicts
		if snaps, err = d.Snapshots(); err != nil {
			return
		}
		for _, s := range snaps {
			spath := s.Properties[DatasetPropName].Value
			sname := spath[strings.Index(spath, "@"):]
			// conflict and resolve
			if ok, _ := cds.FindSnapshotName(sname); ok {
				// snapshot with the same name already exist
				volname := path.Base(spath[:strings.Index(spath, "@")])
				sname = sname + "." + volname
				if err = s.Rename(spath+"."+volname, false, true); err != nil {
					return
				}
			}
			psnaps = append(psnaps, sname)
		}
		if err = cds.Promote(); err != nil {
			return
		}
	}
	// destroy child datasets, since this works recursive
	for _, cd := range d.Children {
		if err = cd.DestroyPromote(); err != nil {
			return
		}
	}
	d.Children = make([]Dataset, 0)
	if err = d.Destroy(false); err != nil {
		return
	}
	// Load with new promoted snapshots
	if len(clones) > 0 && len(psnaps) > 0 {
		var cds Dataset
		if cds, err = DatasetOpen(clones[0]); err != nil {
			return
		}
		defer cds.Close()
		// try to destroy (promoted) snapshots now
		for _, sname := range psnaps {
			if ok, snap := cds.FindSnapshotName(sname); ok {
				snap.Destroy(false)
			}
		}
	}
	return
}

// Snapshots - filter and return all snapshots of dataset
func (d *Dataset) Snapshots() (snaps []Dataset, err error) {
	for _, ch := range d.Children {
		if !ch.IsSnapshot() {
			continue
		}
		snaps = append(snaps, ch)
	}
	return
}

// FindSnapshot - returns true if given path is one of dataset snaphsots
func (d *Dataset) FindSnapshot(path string) (ok bool, snap Dataset) {
	for _, ch := range d.Children {
		if !ch.IsSnapshot() {
			continue
		}
		if ok = (path == ch.Properties[DatasetPropName].Value); ok {
			snap = ch
			break
		}
	}
	return
}

// FindSnapshotName - returns true and snapshot if given snapshot
// name eg. '@snap1' is one of dataset snaphsots
func (d *Dataset) FindSnapshotName(name string) (ok bool, snap Dataset) {
	return d.FindSnapshot(d.Properties[DatasetPropName].Value + name)
}

// Clones - get list of all dataset paths cloned from this
// dataset or this snapshot
// List is sorted descedent by origin snapshot order
func (d *Dataset) Clones() (clones []string, err error) {
	// Clones can only live on same pool
	var root Dataset
	var sortDesc []Dataset
	if root, err = DatasetOpen(d.PoolName()); err != nil {
		return
	}
	defer root.Close()
	dIsSnapshot := d.IsSnapshot()
	// USe breadth first search to find all clones
	queue := make(chan Dataset, 1024)
	defer close(queue) // This will close and cleanup all
	queue <- root      // start from the root element
	for {
		select {
		case ds := <-queue: // pull from queue (breadth first search)
			for _, ch := range ds.Children {
				origin := ch.Properties[DatasetPropOrigin].Value
				if len(origin) > 0 {
					if dIsSnapshot && origin == d.Properties[DatasetPropName].Value {
						// if this dataset is snaphot
						ch.Properties[zfsMaxDatasetProp+1000] = d.Properties[DatasetPropCreateTXG]
						sortDesc = append(sortDesc, ch)
					} else {
						// Check if origin of this dataset is one of snapshots
						ok, snap := d.FindSnapshot(origin)
						if !ok {
							continue
						}
						ch.Properties[zfsMaxDatasetProp+1000] = snap.Properties[DatasetPropCreateTXG]
						sortDesc = append(sortDesc, ch)
					}
				}
				queue <- ch
			}
		default:
			sort.Sort(clonesCreateDesc(sortDesc))
			// This way we get clones ordered from most recent sanpshots first
			for _, c := range sortDesc {
				clones = append(clones, c.Properties[DatasetPropName].Value)
			}
			return
		}
	}
	return
}
