package zfs

import (
	"testing"
	"encoding/json"
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
		var props PoolProperties = pool.Properties
		data, err := json.Marshal(&props)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
		pool.Close()
	})
	t.Run("read features", func(t *testing.T){
		pool, err := PoolOpen(TESTPOOL)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(pool.Features)
		pool.Close()
	})
}
