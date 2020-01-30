package zfs

import (
	"testing"
)

func Test_PoolOpen(t *testing.T) {
	pool, err := PoolOpen(TESTPOOL)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(pool.Features)
	pool.Close()
}

func Test_PoolProperties(t *testing.T) {
	t.Run("read propertis", func(t *testing.T){
		pool, err := PoolOpen(TESTPOOL)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(pool.Properties)
		pool.Close()
	})
}
