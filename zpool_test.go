package zfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

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

func TestPool(t *testing.T) {
	var TSTPoolGUID string
	var s1path, s2path, s3path string
	// first check if pool with same name already exist
	// we don't want conflict
	TSTPoolName := *testPool
	for {
		p, err := PoolOpen(TSTPoolName)
		if err != nil {
			break
		}
		p.Close()
		TSTPoolName += "0"
	}
	var err error
	// Create 3 sparse file in /tmp directory each 5G size, and use them to create
	// mirror TESTPOOL with one spare "disk"
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

	t.Run("create pool", func(t *testing.T){
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
			t.Fatal(err)
		}
		defer pool.Close()

		pguid, err := pool.GetProperty(PoolPropGUID)
		if err != nil {
			t.Fatal("getting pool GUID property should not error, but return: ", err)
		}
		TSTPoolGUID = pguid.Value
		t.Log(pool)
	})

	t.Run("open pool not exist", func(t *testing.T){
		pname := "fail to open this pool"
		p, err := PoolOpen(pname)
		if err == nil {
			t.Fatal("PoolOpen pass when it should fail")
			p.Close()
		}
		if err1, ok := err.(*Error); ok && err1.ErrorCode() == ENoent {
			t.Log(err1)
		} else {
			t.Error(err1)
		}
	})

	t.Run("export", func(t *testing.T) {
		p, err := PoolOpen(TSTPoolName)
		if err != nil {
			t.Error(err)
			return
		}
		p.Export(false, "Test exporting pool")
		defer p.Close()
	})

	t.Run("import", func(t *testing.T) {
		p, err := PoolImport(TSTPoolName, []string{"/tmp"})
		if err != nil {
			t.Error(err)
			return
		}
		defer p.Close()
	})

	t.Run("export force", func(t *testing.T) {
		p, err := PoolOpen(TSTPoolName)
		if err != nil {
			t.Fatal(err)
		}
		p.ExportForce("Test force exporting pool")
		defer p.Close()
	})

	t.Run("import by GUID", func(t *testing.T) {
		p, err := PoolImportByGUID(TSTPoolGUID, []string{"/tmp"})
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
	})

	t.Run("get status and state", func (t *testing.T) {
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
		t.Log("POOL", TSTPoolName, "state:", PoolStateToName(pstate))
	})

	t.Run("destroy pool", func (t *testing.T){
		p, err := PoolOpen(TSTPoolName)
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		if err = p.Destroy(TSTPoolName); err != nil {
			t.Fatal(err.Error())
		}
		p, err = PoolOpen(TSTPoolName)
		if err == nil {
			t.Fatal("should error")
		}
		t.Log(err)
	})

	os.Remove(s1path)
	os.Remove(s2path)
	os.Remove(s3path)
}

// Open and list all pools and them state on the system
// Then list properties of last pool in the list
func TestPoolOpenAll(t *testing.T) {
	var pname string
	pools, err := PoolOpenAll()
	if err != nil {
		t.Error(err)
		return
	}
	for _, p := range pools {
		pname, err = p.Name()
		if err != nil {
			t.Error(err)
			p.Close()
			return
		}
		t.Log(pname)
		pstate, err := p.State()
		if err != nil {
			t.Error(err)
			p.Close()
			return
		}
		t.Log(pstate)
		p.Close()
	}
}

func printVDevTree(t *testing.T, vt VDevTree, pref string) {
	first := pref + vt.Name
	t.Logf("%-30s | %-10s | %-10s | %s\n", first, vt.Type,
		vt.Stat.State.String(), vt.Path)
	for _, v := range vt.Devices {
		printVDevTree(t, v, "  "+pref)
	}
	if len(vt.Spares) > 0 {
		t.Logf("spares:")
		for _, v := range vt.Spares {
			printVDevTree(t, v, "  "+pref)
		}
	}

	if len(vt.L2Cache) > 0 {
		t.Logf("l2cache:")
		for _, v := range vt.L2Cache {
			printVDevTree(t, v, "  "+pref)
		}
	}
}

func TestPoolImportSearch(t *testing.T) {
	pools, err := PoolImportSearch([]string{"/tmp"})
	if err != nil {
		t.Error(err.Error())
		return
	}
	for _, p := range pools {
		t.Log()
		t.Log("---------------------------------------------------------------")
		t.Log("pool: ", p.Name)
		t.Log("guid: ", p.GUID)
		t.Log("state: ", p.State.String())
		t.Logf("%-30s | %-10s | %-10s | %s\n", "NAME", "TYPE", "STATE", "PATH")
		t.Log("---------------------------------------------------------------")
		printVDevTree(t, p.VDevs, "")
	}
}

func TestPoolProp(t *testing.T) {
	t.Log("TEST PoolProp on ", *testPool, " ... ")
	if pool, err := PoolOpen(*testPool); err == nil {
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
		t.Log("Pool property health: ", propHealth.Value)

		propGUID, err := pool.GetProperty(PoolPropGUID)
		if err != nil {
			t.Error(err)
			return
		}
		t.Log("Pool property GUID: ", propGUID.Value)

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
}

func TestPoolVDevTree(t *testing.T) {
	TSTPoolName := *testPool
	var vdevs VDevTree
	t.Log("TEST pool VDevTree ( ", TSTPoolName, " ) ... ")
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
	t.Logf("%-30s | %-10s | %-10s | %s\n", "NAME", "TYPE", "STATE", "PATH")
	t.Log("---------------------------------------------------------------")
	printVDevTree(t, vdevs, "")
}
