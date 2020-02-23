package zfs

import (
	"fmt"
)

func ExamplePoolProp() {
	if pool, err := PoolOpen("SSD"); err == nil {
		print("Pool size is: ", pool.Properties[PoolPropSize].Value)
		// Turn on snapshot listing for pool
		pool.SetProperty(PoolPropListsnaps, "on")
		println("Changed property",
			PoolPropertyToName(PoolPropListsnaps), "to value:",
			pool.Properties[PoolPropListsnaps].Value)

		prop, err := pool.GetProperty(PoolPropHealth)
		if err != nil {
			panic(err)
		}
		println("Update and print out pool health:", prop.Value)
	} else {
		print("Error: ", err)
	}
}

// Open and list all pools on system with them properties
func ExamplePoolOpenAll() {
	// Lets open handles to all active pools on system
	pools, err := PoolOpenAll()
	if err != nil {
		println(err)
	}

	// Print each pool name and properties
	for _, p := range pools {
		// Print fancy header
		fmt.Printf("\n -----------------------------------------------------------\n")
		fmt.Printf("   POOL: %49s   \n", p.Properties[PoolPropName].Value)
		fmt.Printf("|-----------------------------------------------------------|\n")
		fmt.Printf("|  PROPERTY      |  VALUE                |  SOURCE          |\n")
		fmt.Printf("|-----------------------------------------------------------|\n")

		// Iterate pool properties and print name, value and source
		for key, prop := range p.Properties {
			pkey := PoolProp(key)
			if pkey == PoolPropName {
				continue // Skip name its already printed above
			}
			fmt.Printf("|%14s  | %20s  | %15s  |\n",
				PoolPropertyToName(pkey),
				prop.Value, prop.Source)
			println("")
		}
		println("")

		// Close pool handle and free memory, since it will not be used anymore
		p.Close()
	}
}

func ExamplePoolCreate() {
	disks := [2]string{"/dev/disk/by-id/ATA-123", "/dev/disk/by-id/ATA-456"}

	var vdev VDevTree
	var vdevs, mdevs, sdevs []VDevTree

	// build mirror devices specs
	for _, d := range disks {
		mdevs = append(mdevs,
			VDevTree{Type: VDevTypeDisk, Path: d})
	}

	// spare device specs
	sdevs = []VDevTree{
		{Type: VDevTypeDisk, Path: "/dev/disk/by-id/ATA-789"}}

	// pool specs
	vdevs = []VDevTree{
		VDevTree{Type: VDevTypeMirror, Devices: mdevs},
	}

	vdev.Devices = vdevs
	vdev.Spares = sdevs

	// pool properties
	props := make(map[PoolProp]PropertyValue)
	// root dataset filesystem properties
	fsprops := make(map[DatasetProp]PropertyValue)
	// pool features
	features := make(map[string]string)

	// Turn off auto mounting by ZFS
	fsprops[DatasetPropMountpoint] = PropertyValue{Value: "none"}

	// Enable some features
	features["async_destroy"] = "enabled"
	features["empty_bpobj"] = "enabled"
	features["lz4_compress"] = "enabled"

	// Based on specs formed above create test pool as 2 disk mirror and
	// one spare disk
	pool, err := PoolCreate("TESTPOOL", vdev, features, props, fsprops)
	if err != nil {
		println("Error: ", err.Error())
		return
	}
	defer pool.Close()
}

func ExamplePool_Destroy() {
	pname := "TESTPOOL"

	// Need handle to pool at first place
	p, err := PoolOpen(pname)
	if err != nil {
		println("Error: ", err.Error())
		return
	}

	// Make sure pool handle is free after we are done here
	defer p.Close()

	if err = p.Destroy("Example of pool destroy (TESTPOOL)"); err != nil {
		println("Error: ", err.Error())
		return
	}
}

func ExamplePoolImport() {
	p, err := PoolImport("TESTPOOL", []string{"/dev/disk/by-id"})
	if err != nil {
		panic(err)
	}
	p.Close()
}

func ExamplePool_Export() {
	p, err := PoolOpen("TESTPOOL")
	if err != nil {
		panic(err)
	}
	defer p.Close()
	if err = p.Export(false, "Example exporting pool"); err != nil {
		panic(err)
	}
}

func ExamplePool_ExportForce() {
	p, err := PoolOpen("TESTPOOL")
	if err != nil {
		panic(err)
	}
	defer p.Close()
	if err = p.ExportForce("Example exporting pool"); err != nil {
		panic(err)
	}
}

func ExamplePool_State() {
	p, err := PoolOpen("TESTPOOL")
	if err != nil {
		panic(err)
	}
	defer p.Close()
	pstate, err := p.State()
	if err != nil {
		panic(err)
	}
	println("POOL TESTPOOL state:", PoolStateToName(pstate))
}

// func TestPool_VDevTree(t *testing.T) {
// 	type fields struct {
// 		poolName string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 		{
// 			name:    "test1",
// 			fields:  fields{"TESTPOOL"},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			pool, _ := PoolOpen(tt.fields.poolName)
// 			defer pool.Close()
// 			gotVdevs, err := pool.VDevTree()
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("Pool.VDevTree() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			jsonData, _ := json.MarshalIndent(gotVdevs, "", "\t")
// 			t.Logf("gotVdevs: %s", string(jsonData))
// 		})
// 	}
// }
