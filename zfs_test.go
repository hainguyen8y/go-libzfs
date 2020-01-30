package zfs

import (
	"fmt"
	"testing"
)

/* ------------------------------------------------------------------------- */
// HELPERS:
var TSTDatasetPath = TSTPoolName + "/DATASET"
var TSTVolumePath = TSTDatasetPath + "/VOLUME"
var TSTDatasetPathSnap = TSTDatasetPath + "@test"

func printDatasets(ds []Dataset) error {
	for _, d := range ds {

		path, err := d.Path()
		if err != nil {
			return err
		}
		p, err := d.GetProperty(DatasetPropType)
		if err != nil {
			return err
		}
		fmt.Printf(" %30s | %10s\n", path, p.Value)
		if len(d.Children) > 0 {
			printDatasets(d.Children)
		}
	}
	return nil
}

/* ------------------------------------------------------------------------- */
// TESTS:

func TestDatasetCreate(t *testing.T) {
	// reinit names used in case TESTPOOL was in conflict
	TSTDatasetPath = TSTPoolName + "/DATASET"
	TSTVolumePath = TSTDatasetPath + "/VOLUME"
	TSTDatasetPathSnap = TSTDatasetPath + "@test"

	println("TEST DatasetCreate(", TSTDatasetPath, ") (filesystem) ... ")
	props := make(map[DatasetProp]PropertyValue)
	d, err := DatasetCreate(TSTDatasetPath, DatasetTypeFilesystem, props)
	if err != nil {
		t.Error(err)
		return
	}
	d.Close()
	print("PASS\n\n")

	strSize := "536870912" // 512M

	println("TEST DatasetCreate(", TSTVolumePath, ") (volume) ... ")
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
	print("PASS\n\n")
}

func TestDatasetOpen(t *testing.T) {
	println("TEST DatasetOpen(", TSTDatasetPath, ") ... ")
	d, err := DatasetOpen(TSTDatasetPath)
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()
	print("PASS\n\n")

	println("TEST Set/GetUserProperty(prop, value string) ... ")
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
	println("go-libzfs:test", " = ",
		p.Value)
	print("PASS\n\n")
}

func TestDatasetSetProperty(t *testing.T) {
	println("TEST Dataset SetProp(", TSTDatasetPath, ") ... ")
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
		println(prop.Value)
		if prop.Value != "on" {
			t.Error(fmt.Errorf("Update of dataset property failed"))
			return
		}
	}
	print("PASS\n\n")
	return
}

func TestDatasetOpenAll(t *testing.T) {
	println("TEST DatasetOpenAll()/DatasetCloseAll() ... ")
	ds, err := DatasetOpenAll()
	if err != nil {
		t.Error(err)
		return
	}
	if err = printDatasets(ds); err != nil {
		DatasetCloseAll(ds)
		t.Error(err)
		return
	}
	DatasetCloseAll(ds)
	print("PASS\n\n")
}

func TestDatasetSnapshot(t *testing.T) {
	println("TEST DatasetSnapshot(", TSTDatasetPath, ", true, ...) ... ")
	props := make(map[DatasetProp]PropertyValue)
	d, err := DatasetSnapshot(TSTDatasetPathSnap, true, props)
	if err != nil {
		t.Error(err)
		return
	}
	defer d.Close()
	print("PASS\n\n")
}

func TestDatasetHoldRelease(t *testing.T) {
	println("TEST Hold/Release(", TSTDatasetPathSnap, ", true, ...) ... ")
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
		println("tag:", tag.Name, "timestamp:", tag.Timestamp.String())
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
		println("* tag:", tag.Name, "timestamp:", tag.Timestamp.String())
	}
	print("PASS\n\n")
}

func TestDatasetDestroy(t *testing.T) {
	println("TEST DATASET Destroy( ", TSTDatasetPath, " ) ... ")
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
	print("PASS\n\n")
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
/* ------------------------------------------------------------------------- */
// EXAMPLES:

// Example of creating ZFS volume
func ExampleDatasetCreate() {
	// Create map to represent ZFS dataset properties. This is equivalent to
	// list of properties you can get from ZFS CLI tool, and some more
	// internally used by lib
	props := make(map[DatasetProp]PropertyValue)

	// I choose to create (block) volume 1GiB in size. Size is just ZFS dataset
	// property and this is done as map of strings. So, You have to either
	// specify size as base 10 number in string, or use strconv package or
	// similar to convert in to string (base 10) from numeric type.
	strSize := "1073741824"

	props[DatasetPropVolsize] = PropertyValue{Value: strSize}
	// In addition I explicitly choose some more properties to be set.
	props[DatasetPropVolblocksize] = PropertyValue{Value: "4096"}
	props[DatasetPropReservation] = PropertyValue{Value: strSize}

	// Lets create desired volume
	d, err := DatasetCreate("TESTPOOL/VOLUME1", DatasetTypeVolume, props)
	if err != nil {
		println(err.Error())
		return
	}
	// Dataset have to be closed for memory cleanup
	defer d.Close()

	println("Created zfs volume TESTPOOL/VOLUME1")
}

func ExampleDatasetOpen() {
	// Open dataset and read its available space
	d, err := DatasetOpen("TESTPOOL/DATASET1")
	if err != nil {
		panic(err.Error())
	}
	defer d.Close()
	var p PropertyValue
	if p, err = d.GetProperty(DatasetPropAvailable); err != nil {
		panic(err.Error())
	}
	println(DatasetPropertyToName(DatasetPropAvailable), " = ",
		p.Value)
}

func ExampleDatasetOpenAll() {
	datasets, err := DatasetOpenAll()
	if err != nil {
		panic(err.Error())
	}
	defer DatasetCloseAll(datasets)

	// Print out path and type of root datasets
	for _, d := range datasets {
		path, err := d.Path()
		if err != nil {
			panic(err.Error())
		}
		p, err := d.GetProperty(DatasetPropType)
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("%30s | %10s\n", path, p.Value)
	}

}
