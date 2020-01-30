package zfs

import (
	"testing"
)
func TestDestroySnapshot(t *testing.T) {
	err := DestroySnapshot("CustDATA/tank2/tank1@zfs-auto-snap_daily-2019-06-05-1707")
	if err != nil {
		t.Log(err)
	}
}
