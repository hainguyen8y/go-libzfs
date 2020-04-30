package zfs

import (
	"testing"
)

func TestBookmark(t *testing.T) {
	testDatasetName := *testPool+"/tank1"
	testSnapshotName := testDatasetName + "@test"
	snap, err := DatasetSnapshot(testSnapshotName, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer snap.Close()

	t.Run("bookmark a dataset that is filesystem", func(t *testing.T) {
		ds, err := DatasetOpen(testDatasetName)
		if err != nil {
			t.Error(err)
			return
		}
		defer ds.Close()
		bm, err := ds.CreateBookmark("#last")
		if err == nil {
			defer bm.Close()
			t.Fatal("have to return an error")
		}
		t.Log(err)
	})
	t.Run("bookmark a snapshot exists", func(t *testing.T){
		ds, err := DatasetOpen(testSnapshotName)
		if err != nil {
			t.Error(err)
			return
		}
		defer ds.Close()
		bm, err := ds.CreateBookmark("last")
		if err != nil {
			t.Fatal("error: ", err)
		}
		defer bm.Close()
		t.Log(bm)
	})
	t.Run("delete a bookmark", func(t *testing.T){
		ds, err := DatasetOpen(testDatasetName+"#last")
		if err != nil {
			t.Error(err)
			return
		}
		defer ds.Close()
		err = ds.Destroy(false)
		if err != nil {
			t.Fatal(err)
		}
	})

	snap.Destroy(false)
}
