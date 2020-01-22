package zfs

import (
	"testing"
)

func TestListRoot(t *testing.T) {
	dts, err := List(ListOptions{
		Types: DatasetTypeSnapshot,
		Recursive: true,
		Depth: 1,
		Paths: []string{"CustDATA/tank2", "CustDATA"},
	});
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(dts))
	DatasetCloseAll(dts)
}

func TestListNoExistPath(t *testing.T) {
	dts, err := List(ListOptions{
		Types: DatasetTypeSnapshot,
		Recursive: true,
		Depth: 1,
		Paths: []string{"CustDATA/tank2dad"},
	});
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(len(dts))
	DatasetCloseAll(dts)
}
