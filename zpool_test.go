package zfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

/* ------------------------------------------------------------------------- */
// HELPERS:

var TSTPoolName = "TESTPOOL"
var TSTPoolGUID string

func CreateTmpSparse(prefix string, size int64) (path string, err error) {
	sf, err := ioutil.TempFile("/tmp", prefix)
	if err != nil {
		return
	}
	defer sf.Close()
	if err = sf.Truncate(size); err != nil {
		return
	}
	path = sf.Name()
	return
}

var s1path, s2path, s3path string

// This will create sparse files in tmp directory,
// for purpose of creating test pool.
func CreateTestpoolVdisks() (err error) {
	if s1path, err = CreateTmpSparse("zfs_test_", 0x140000000); err != nil {
		return
	}
	if s2path, err = CreateTmpSparse("zfs_test_", 0x140000000); err != nil {
		// try cleanup
		os.Remove(s1path)
		return
	}
	if s3path, err = CreateTmpSparse("zfs_test_", 0x140000000); err != nil {
		// try cleanup
		os.Remove(s1path)
		os.Remove(s2path)
		return
	}
	return
}

// Cleanup sparse files used for tests
func CleanupVDisks() {
	// try cleanup
	os.Remove(s1path)
	os.Remove(s2path)
	os.Remove(s3path)
}

/* ------------------------------------------------------------------------- */
// TESTS:

// Create 3 sparse file in /tmp directory each 5G size, and use them to create
// mirror TESTPOOL with one spare "disk"
func TestPoolCreate(t *testing.T) {
	println("TEST PoolCreate ... ")
	// first check if pool with same name already exist
	// we don't want conflict
	for {
		p, err := PoolOpen(TSTPoolName)
		if err != nil {
			break
		}
		p.Close()
		TSTPoolName += "0"
	}
	var err error

	if err = CreateTestpoolVdisks(); err != nil {
		t.Error(err)
		return
	}

	disks := [2]string{s1path, s2path}

	var vdev VDevTree
	var vdevs, mdevs, sdevs []VDevTree
	for _, d := range disks {
		mdevs = append(mdevs,
			VDevTree{Type: VDevTypeFile, Path: d})
	}
	sdevs = []VDevTree{
		{Type: VDevTypeFile, Path: s3path}}
	vdevs = []VDevTree{
		VDevTree{Type: VDevTypeMirror, Devices: mdevs},
	}
	vdev.Devices = vdevs
	vdev.Spares = sdevs

	props := make(map[PoolProp]PropertyValue)
	fsprops := make(map[DatasetProp]PropertyValue)
	features := make(map[string]string)
	fsprops[DatasetPropMountpoint] = PropertyValue{Value: "none"}
	features["async_destroy"] = FENABLED
	features["empty_bpobj"] = FENABLED
	features["lz4_compress"] = FENABLED

	pool, err := PoolCreate(TSTPoolName, vdev, features, props, fsprops)
	if err != nil {
		t.Error(err)
		// try cleanup
		os.Remove(s1path)
		os.Remove(s2path)
		os.Remove(s3path)
		return
	}
	defer pool.Close()

	pguid, _ := pool.GetProperty(PoolPropGUID)
	TSTPoolGUID = pguid.Value

	print("PASS\n\n")
}

// Open and list all pools and them state on the system
// Then list properties of last pool in the list
func TestPoolOpenAll(t *testing.T) {
	println("TEST PoolOpenAll() ... ")
	var pname string
	pools, err := PoolOpenAll()
	if err != nil {
		t.Error(err)
		return
	}
	println("\tThere is ", len(pools), " ZFS pools.")
	for _, p := range pools {
		pname, err = p.Name()
		if err != nil {
			t.Error(err)
			p.Close()
			return
		}
		pstate, err := p.State()
		if err != nil {
			t.Error(err)
			p.Close()
			return
		}
		println("\tPool: ", pname, " state: ", pstate)
		p.Close()
	}
	print("PASS\n\n")
}

func TestPoolDestroy(t *testing.T) {
	println("TEST POOL Destroy( ", TSTPoolName, " ) ... ")
	p, err := PoolOpen(TSTPoolName)
	if err != nil {
		t.Error(err)
		return
	}
	defer p.Close()
	if err = p.Destroy(TSTPoolName); err != nil {
		t.Error(err.Error())
		return
	}
	print("PASS\n\n")
}

func TestFailPoolOpen(t *testing.T) {
	println("TEST open of non existing pool ... ")
	pname := "fail to open this pool"
	p, err := PoolOpen(pname)
	if err != nil {
		print("PASS\n\n")
		return
	}
	t.Error("PoolOpen pass when it should fail")
	p.Close()
}

func TestExport(t *testing.T) {
	println("TEST POOL Export( ", TSTPoolName, " ) ... ")
	p, err := PoolOpen(TSTPoolName)
	if err != nil {
		t.Error(err)
		return
	}
	p.Export(false, "Test exporting pool")
	defer p.Close()
	print("PASS\n\n")
}

func TestExportForce(t *testing.T) {
	println("TEST POOL ExportForce( ", TSTPoolName, " ) ... ")
	p, err := PoolOpen(TSTPoolName)
	if err != nil {
		t.Error(err)
		return
	}
	p.ExportForce("Test force exporting pool")
	defer p.Close()
	print("PASS\n\n")
}

func TestImport(t *testing.T) {
	println("TEST POOL Import( ", TSTPoolName, " ) ... ")
	p, err := PoolImport(TSTPoolName, []string{"/tmp"})
	if err != nil {
		t.Error(err)
		return
	}
	defer p.Close()
	print("PASS\n\n")
}

func TestImportByGUID(t *testing.T) {
	println("TEST POOL ImportByGUID( ", TSTPoolGUID, " ) ... ")
	p, err := PoolImportByGUID(TSTPoolGUID, []string{"/tmp"})
	if err != nil {
		t.Error(err)
		return
	}
	defer p.Close()
	print("PASS\n\n")
}

func printVDevTree(vt VDevTree, pref string) {
	first := pref + vt.Name
	fmt.Printf("%-30s | %-10s | %-10s | %s\n", first, vt.Type,
		vt.Stat.State.String(), vt.Path)
	for _, v := range vt.Devices {
		printVDevTree(v, "  "+pref)
	}
	if len(vt.Spares) > 0 {
		fmt.Println("spares:")
		for _, v := range vt.Spares {
			printVDevTree(v, "  "+pref)
		}
	}

	if len(vt.L2Cache) > 0 {
		fmt.Println("l2cache:")
		for _, v := range vt.L2Cache {
			printVDevTree(v, "  "+pref)
		}
	}
}

func TestPoolImportSearch(t *testing.T) {
	println("TEST PoolImportSearch")
	pools, err := PoolImportSearch([]string{"/tmp"})
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, p := range pools {
		println()
		println("---------------------------------------------------------------")
		println("pool: ", p.Name)
		println("guid: ", p.GUID)
		println("state: ", p.State.String())
		fmt.Printf("%-30s | %-10s | %-10s | %s\n", "NAME", "TYPE", "STATE", "PATH")
		println("---------------------------------------------------------------")
		printVDevTree(p.VDevs, "")
	}
	print("PASS\n\n")
}

func TestPoolProp(t *testing.T) {
	println("TEST PoolProp on ", TSTPoolName, " ... ")
	if pool, err := PoolOpen(TSTPoolName); err == nil {
		defer pool.Close()
		// Turn on snapshot listing for pool
		pool.SetProperty(PoolPropListsnaps, "off")
		// Verify change is succesfull
		if pool.Properties[PoolPropListsnaps].Value != "off" {
			t.Error(fmt.Errorf("Update of pool property failed"))
			return
		}

		// Test fetching property
		propHealth, err := pool.GetProperty(PoolPropHealth)
		if err != nil {
			t.Error(err)
			return
		}
		println("Pool property health: ", propHealth.Value)

		propGUID, err := pool.GetProperty(PoolPropGUID)
		if err != nil {
			t.Error(err)
			return
		}
		println("Pool property GUID: ", propGUID.Value)

		// this test pool should not be bootable
		prop, err := pool.GetProperty(PoolPropBootfs)
		if err != nil {
			t.Error(err)
			return
		}
		if prop.Value != "-" {
			t.Errorf("Failed at bootable fs property evaluation")
			return
		}

		// fetch all properties
		if err = pool.ReloadProperties(); err != nil {
			t.Error(err)
			return
		}
	} else {
		t.Error(err)
		return
	}
	print("PASS\n\n")
}

func TestPoolStatusAndState(t *testing.T) {
	println("TEST pool Status/State ( ", TSTPoolName, " ) ... ")
	pool, err := PoolOpen(TSTPoolName)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer pool.Close()

	if _, err = pool.Status(); err != nil {
		t.Error(err.Error())
		return
	}

	var pstate PoolState
	if pstate, err = pool.State(); err != nil {
		t.Error(err.Error())
		return
	}
	println("POOL", TSTPoolName, "state:", PoolStateToName(pstate))

	print("PASS\n\n")
}

func TestPoolVDevTree(t *testing.T) {
	var vdevs VDevTree
	println("TEST pool VDevTree ( ", TSTPoolName, " ) ... ")
	pool, err := PoolOpen(TSTPoolName)
	if err != nil {
		t.Error(err.Error())
		return
	}
	defer pool.Close()
	vdevs, err = pool.VDevTree()
	if err != nil {
		t.Error(err.Error())
		return
	}
	fmt.Printf("%-30s | %-10s | %-10s | %s\n", "NAME", "TYPE", "STATE", "PATH")
	println("---------------------------------------------------------------")
	printVDevTree(vdevs, "")
	print("PASS\n\n")
}

/* ------------------------------------------------------------------------- */
// EXAMPLES:

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
