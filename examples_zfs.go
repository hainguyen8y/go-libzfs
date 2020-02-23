package zfs

import (
	"fmt"
)

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
	d, err := DatasetCreate(*testPool+"/VOLUME1", DatasetTypeVolume, props)
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
	d, err := DatasetOpen(*testPool+"/DATASET1")
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
