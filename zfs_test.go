package zfs

import (
	"fmt"
	"flag"
	"testing"
)
//go test -v -run TestDatasetCreate -args --pool=data
var hostAddress = flag.String("host", "127.0.0.1:10000", "the host running zfs service")
var testPool = flag.String("pool", "CustDATA", "the pool of host")
/* ------------------------------------------------------------------------- */
// HELPERS:
var TSTDatasetPath = *testPool + "/DATASET"
var TSTVolumePath = TSTDatasetPath + "/VOLUME"
var TSTDatasetPathSnap = TSTDatasetPath + "@test"

func printDatasets(t *testing.T, ds []Dataset) error {
	for _, d := range ds {

		path, err := d.Path()
		if err != nil {
			return err
		}
		p, err := d.GetProperty(DatasetPropType)
		if err != nil {
			return err
		}
		t.Logf(" %30s | %10s\n", path, p.Value)
		if len(d.Children) > 0 {
			printDatasets(t, d.Children)
		}
	}
	return nil
}

/* ------------------------------------------------------------------------- */
// TESTS:

func TestDatasetCreate(t *testing.T) {
	// reinit names used in case TESTPOOL was in conflict
	TSTDatasetPath = *testPool + "/DATASET"
	TSTVolumePath = TSTDatasetPath + "/VOLUME"
	TSTDatasetPathSnap = TSTDatasetPath + "@test"

	t.Log("TEST DatasetCreate(", TSTDatasetPath, ") (filesystem) ... ")
	props := make(map[DatasetProp]PropertyValue)
	d, err := DatasetCreate(TSTDatasetPath, DatasetTypeFilesystem, props)
	if err != nil {
		t.Error(err)
		return
	}
	d.Close()

	strSize := "536870912" // 512M

	t.Log("TEST DatasetCreate(", TSTVolumePath, ") (volume) ... ")
	props[DatasetPropVolsize] = PropertyValue{Value: strSize}
	// In addition I explicitly choose some more properties to be set.
	props[DatasetPropVolblocksize] = PropertyValue{Value: "4096"}
	props[DatasetPropReservation] = PropertyValue{Value: strSize}
	d, err = DatasetCreate(TSTVolumePath, DatasetTypeVolume, props)
	if err != nil {
		t.Error(err)
		return
	}
	d.Close()
}

func TestDatasetOpen(t *testing.T) {
	t.Log("TEST DatasetOpen(", TSTDatasetPath, ") ... ")
	d, err := DatasetOpen(TSTDatasetPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()

	t.Log("TEST Set/GetUserProperty(prop, value string) ... ")
	var p PropertyValue
	// Test set/get user property
	if err = d.SetUserProperty("go-libzfs:test", "yes"); err != nil {
		t.Error(err)
		return
	}
	if p, err = d.GetUserProperty("go-libzfs:test"); err != nil {
		t.Error(err)
		return
	}
	t.Log("go-libzfs:test", " = ",
		p.Value)
}

func TestDatasetSetProperty(t *testing.T) {
	t.Log("TEST Dataset SetProp(", TSTDatasetPath, ") ... ")
	d, err := DatasetOpen(TSTDatasetPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()
	if err = d.SetProperty(DatasetPropOverlay, "on"); err != nil {
		t.Error(err)
		return
	}
	if prop, err := d.GetProperty(DatasetPropOverlay); err != nil {
		t.Error(err)
		return
	} else {
		t.Log(prop.Value)
		if prop.Value != "on" {
			t.Error(fmt.Errorf("Update of dataset property failed"))
			return
		}
	}
	return
}

func TestDatasetOpenAll(t *testing.T) {
	t.Log("TEST DatasetOpenAll()/DatasetCloseAll() ... ")
	ds, err := DatasetOpenAll()
	if err != nil {
		t.Error(err)
		return
	}
	if err = printDatasets(t, ds); err != nil {
		DatasetCloseAll(ds)
		t.Error(err)
		return
	}
	DatasetCloseAll(ds)
}

func TestDatasetSnapshot(t *testing.T) {
	t.Log("TEST DatasetSnapshot(", TSTDatasetPath, ", true, ...) ... ")
	props := make(map[DatasetProp]PropertyValue)
	d, err := DatasetSnapshot(TSTDatasetPathSnap, true, props)
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()
}

func TestDatasetHoldRelease(t *testing.T) {
	t.Log("TEST Hold/Release(", TSTDatasetPathSnap, ", true, ...) ... ")
	d, err := DatasetOpen(TSTDatasetPathSnap)
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()
	err = d.Hold("keep")
	if err != nil {
		t.Error(err)
		return
	}

	var tags []HoldTag
	tags, err = d.Holds()
	if err != nil {
		t.Error(err)
		return
	}
	for _, tag := range tags {
		t.Log("tag:", tag.Name, "timestamp:", tag.Timestamp.String())
	}

	err = d.Release("keep")
	if err != nil {
		t.Error(err)
		return
	}

	tags, err = d.Holds()
	if err != nil {
		t.Error(err)
		return
	}
	for _, tag := range tags {
		t.Log("* tag:", tag.Name, "timestamp:", tag.Timestamp.String())
	}
}

func TestDatasetDestroy(t *testing.T) {
	t.Log("TEST DATASET Destroy( ", TSTDatasetPath, " ) ... ")
	d, err := DatasetOpen(TSTDatasetPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()
	if err = d.DestroyRecursive(); err != nil {
		t.Error(err)
		return
	}
}

/*
func TestGetWrittenSnapshot(t *testing.T) {
	d, err := DatasetOpen("CustDATA")
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()
	userprop_name := "written" + "@zfs-auto-snap_hourly-2019-06-07-0403"
	t.Log(userprop_name)
	userprop, err := d.GetUserProperty(userprop_name)
	if err != nil {
		t.Log(err)
	} else {
		t.Logf("%s->%s:%s\n", userprop_name, userprop.Source, userprop.Value)
	}
	print("PASS\n\n")
}
*/
