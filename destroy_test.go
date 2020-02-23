package zfs

import (
	"testing"
)

func TestDataset_DestroyPromote(t *testing.T) {
	var c1, c2 Dataset

	props := make(map[DatasetProp]PropertyValue)

	//before testing
	datasetName := *testPool+"/original"
	cloneName1 := *testPool+"/clone1"
	cloneName2 := *testPool+"/clone2"

	d, err := DatasetOpen(datasetName)
	if err == nil {
		err = d.DestroyRecursive()
		if err != nil {
			t.Fatal(err)
		}
		d.Close()
	}

	d, err = DatasetOpen(cloneName1)
	if err == nil {
		err = d.DestroyRecursive()
		if err != nil {
			t.Fatal(err)
		}
		d.Close()
	}

	d, err = DatasetOpen(cloneName2)
	if err == nil {
		err = d.DestroyRecursive()
		if err != nil {
			t.Fatal(err)
		}
		d.Close()
	}

	d, err = DatasetCreate(datasetName,
		DatasetTypeFilesystem, make(map[DatasetProp]PropertyValue))
	if err != nil {
		t.Errorf("DatasetCreate(\"%s\") error: %v", datasetName, err)
		return
	}

	s1, _ := DatasetSnapshot(d.Properties[DatasetPropName].Value+"@snap2", false, props)
	s2, _ := DatasetSnapshot(d.Properties[DatasetPropName].Value+"@snap1", false, props)

	c1, err = s1.Clone(cloneName1, nil)
	if err != nil {
		t.Errorf("d.Clone(\"%s/clone1\", props)) error: %v", *testPool, err)
		d.Close()
		return
	}

	DatasetSnapshot(c1.Properties[DatasetPropName].Value+"@snap1", false, props)

	c2, err = s2.Clone(cloneName2, nil)
	if err != nil {
		t.Errorf("c1.Clone(\"%s/clone1\", props)) error: %v", *testPool, err)
		d.Close()
		c1.Close()
		return
	}
	s2.Close()

	DatasetSnapshot(c2.Properties[DatasetPropName].Value+"@snap0", false, props)
	c1.Close()
	c2.Close()

	// reopen pool
	d.Close()
	if d, err = DatasetOpen(datasetName); err != nil {
		t.Error("DatasetOpen")
		return
	}

	if err = d.DestroyPromote(); err != nil {
		t.Errorf("DestroyPromote error: %v", err)
		d.Close()
		return
	}
	t.Log("Destroy promote completed with success")
	d.Close()
}
