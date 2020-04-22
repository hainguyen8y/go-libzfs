package zfs

import (
	"testing"
	"encoding/json"
	"strings"
)

func TestDatasetType(t *testing.T) {
	t.Run("parsing the correct DatasetType", func (t *testing.T) {
		var types []DatasetType
		data := []byte(`["filesystem","snapshot", "volume"]`)
		err := json.Unmarshal(data, &types)
		if err != nil {
			t.Fatal(err)
		}
		if types[0] != DatasetTypeFilesystem {
			t.Fatalf("parse %s expect DatasetTypeFilesystem", types[0].String())
		}
		if types[1] != DatasetTypeSnapshot {
			t.Fatalf("parse %s expect DatasetTypeSnapshot", types[1].String())
		}
		if types[2] != DatasetTypeVolume {
			t.Fatalf("parse %s expect DatasetTypeVolume", types[2].String())
		}
		t.Log(types)
	})
	t.Run("parsing the incorrect DatasetType", func (t *testing.T) {
		var types []DatasetType
		data := []byte(`["filesystems","snapshot", "volume"]`)
		err := json.Unmarshal(data, &types)
		if err == nil {
			t.Fatal("should return not exists")
		}
		t.Log(err)
	})
	t.Run("convert to the json", func (t *testing.T) {
		types := []DatasetType{DatasetTypeSnapshot, DatasetTypeVolume}
		correct := "[\"snapshot\",\"volume\"]"
		data, err := json.Marshal(&types)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Compare(correct, string(data)) != 0 {
			t.Fatalf("correct \"%s\" but return \"%s\"", correct, string(data))
		}
	})
}

func Test_DatasetOpen(t *testing.T) {

	testDatasetName := *testPool+"/tank1"

	d, err := DatasetCreate(testDatasetName, DatasetTypeFilesystem, nil)
	if err == nil {
		d.Close()
	}

	t.Run("open dataset not exist", func(t *testing.T) {
		dt, err := DatasetOpenSingle(*testPool+"/tank2/tank4")
		if err == nil {
			defer dt.Close()
			t.Fatal("should not exist")
		}
		if err1, ok := err.(*Error); ok {
			if ErrorCode(err1.ErrorCode()) == ENoent {
				t.Log("correct with", err)
			} else {
				t.Errorf("return code %d, expect %d\n", ErrorCode(err1.ErrorCode()), ENoent)
			}

		} else {
			t.Fatal(err)
		}
	})
	t.Run("open dataset", func(t *testing.T){
		dt, err := DatasetOpenSingle(testDatasetName)
		if err != nil {
			t.Fatal(err)
		}
		defer dt.Close()
		t.Log(dt)
		data, err := json.Marshal(dt)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
		t.Log(dt.Type)
	})
	t.Run("get properties of dataset", func(t *testing.T){
		dt, err := DatasetOpenSingle(testDatasetName)
		if err != nil {
			t.Fatal(err)
		}
		defer dt.Close()
		var props DatasetProperties = dt.Properties
		data, err := json.Marshal(&props)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	})
}

func Test_Bookmarks(t *testing.T) {
	testBookmarkName := *testPool+"#abc"
	dt, err := DatasetOpenSingle(testBookmarkName)
	if err == nil {
		dt.DestroyRecursive()
		dt.Close()
	}

	t.Run("open bookmark no exist", func(t *testing.T) {
		dt, err := DatasetOpenSingle(testBookmarkName)
		if err != nil {
			if err1, ok := err.(*Error); ok && err1.ErrorCode() != int(ENoent) {
				t.Fatal(err1)
			}
			t.Log(err)
			return
		}
		dt.Close()
		t.Fatal(err)
	})
}
