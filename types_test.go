package zfs

import(
	"testing"
	"encoding/json"
)

func TestProperties(t *testing.T) {
	t.Run("convert to json", func(t *testing.T) {
		var props = make(DatasetProperties)
		props[DatasetPropName] = PropertyValue{
			Value: "abc",
			Source: "-",
		}
		props[DatasetPropWritten] = PropertyValue{
			Value: "234234",
		}
		data, err := json.Marshal(&props)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	})
	t.Run("parse from json", func(t *testing.T) {
		data := []byte(`{"name":{"value":"abc","source":"-"},"written":{"value":"234234","source":""}}`)
		var props = make(DatasetProperties)
		err := json.Unmarshal(data, &props)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(props)
	})
	t.Run("parse from json, emit a \"source\" field", func(t *testing.T) {
		data := []byte(`{"name":{"value":"abc","source":"-"},"written":{"value":"234234"}}`)
		var props = make(DatasetProperties)
		err := json.Unmarshal(data, &props)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(props)
	})
}

func TestDataTypeProp(t *testing.T) {
	t.Run("marshal", func(t *testing.T){
		var p DatasetProp = DatasetPropSnapshotCount
		data, err := json.Marshal(&p)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(string(data))
	})
	t.Run("unmarshal", func(t *testing.T){
		data := []byte(`"snapshot_count"`)
		var p DatasetProp
		err := json.Unmarshal(data, &p)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(p)
	})
}
