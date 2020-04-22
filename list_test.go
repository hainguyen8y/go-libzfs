package zfs

import (
	"testing"
)

func SetupTest() error {
	dt, err := DatasetOpenSingle(*testPool)
	if err != nil {
		return err
	}
	dt.Close()
	return nil
}

func TearDownTest() error {
	return nil
}

func TestListRoot(t *testing.T) {
	err := SetupTest()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("list snapshots of pool without recursive", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeSnapshot,
			Recursive: false,
		});
		if err != nil {
			t.Error(err)
			return
		} else {
			defer DatasetCloseAll(dts)
		}
		if len(dts) > 0 {
			t.Error("should no snapshot")
		}
	})
	t.Run("list snapshots of pool recursively", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeSnapshot,
			Recursive: true,
		});
		if err != nil {
			t.Error(err)
			return
		} else {
			defer DatasetCloseAll(dts)
		}
		if len(dts) == 0 {
			t.Error("should have snapshot")
		}
	})
	t.Run("list snapshots of pool recursively, depth = 1", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeSnapshot,
			Recursive: true,
			Depth: 1,
		});
		if err != nil {
			t.Error(err)
			return
		} else {
			defer DatasetCloseAll(dts)
		}
		if len(dts) == 0 {
			t.Error("should have snapshot")
		}
	})
	t.Run("list filesystem of pool recursively, depth = 1", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeFilesystem,
			Recursive: true,
			Depth: 1,
		});
		if err != nil {
			t.Error(err)
			return
		} else {
			defer DatasetCloseAll(dts)
		}
		if len(dts) == 0 {
			t.Error("should have snapshot")
		}
	})
	TearDownTest()
}

func TestListWithPath(t *testing.T) {
	err := SetupTest()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("list all snapshots of a dataset recursively, depth = 1, not exist", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeSnapshot,
			Recursive: true,
			Depth: 1,
			Paths: []string{*testPool+"/tank1-abc"},
		});
		if err != nil {
			t.Log(err.(*Error).ErrorCode(), err.(*Error).Error())
		} else {
			defer DatasetCloseAll(dts)
			t.Error("should have an error")
		}
	})
	t.Run("list all snapshots of a dataset with recursive, depth = 1, exists", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeSnapshot,
			Recursive: true,
			Depth: 1,
			Paths: []string{*testPool+"/tank1"},
		});
		if err != nil {
			t.Error(err.(*Error).ErrorCode(), err.(*Error).Error())
		} else {
			defer DatasetCloseAll(dts)
		}
		t.Log(len(dts))
	})
	t.Run("list all snapshots recursively", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeSnapshot,
			Recursive: true,
			Paths: []string{*testPool},
		});
		if err != nil {
			t.Error(err.(*Error).ErrorCode(), err.(*Error).Error())
		} else {
			defer DatasetCloseAll(dts)
		}
		t.Log(len(dts))
	})
	t.Run("list all filesystems recursively", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeFilesystem,
			Recursive: true,
			Paths: []string{*testPool},
		});
		if err != nil {
			t.Error(err.(*Error).ErrorCode(), err.(*Error).Error())
		} else {
			defer DatasetCloseAll(dts)
		}
		t.Log(len(dts))
	})
	TearDownTest()
}

func TestListBookmarks(t *testing.T) {
	err := SetupTest()
	if err != nil {
		t.Fatal(err)
	}
	t.Run("list all bookmark recursively", func (t *testing.T) {
		dts, err := List(ListOptions{
			Types: DatasetTypeBookmark,
			Recursive: true,
			Paths: []string{*testPool},
		});
		if err != nil {
			t.Error(err.(*Error).ErrorCode(), err.(*Error).Error())
		} else {
			defer DatasetCloseAll(dts)
		}
		t.Log(dts)
	})
	TearDownTest()
}
