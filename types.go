package zfs
// #include <stdlib.h>
// #include <libzfs.h>
// #include "common.h"
// #include "zpool.h"
// #include "zfs.h"
import "C"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

var stringToDatasetPropDic = make(map[string]DatasetProp)
var stringToPoolPropDic = make(map[string]PoolProp)

func init() {
	for i := DatasetPropType; i < DatasetNumProps; i++ {
		stringToDatasetPropDic[i.String()] = i
	}
	for i := PoolPropName; i < PoolNumProps; i++ {
		stringToPoolPropDic[i.String()] = i
	}
}

func (p *DatasetProp) String() string {
	return C.GoString(C.zfs_prop_to_name((C.zfs_prop_t)(*p)))
}

func (p *DatasetProp) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(p.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *DatasetProp) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	prop, ok := stringToDatasetPropDic[j]
	if !ok {
		return fmt.Errorf("prop \"%s\" not exists", j)
	}
	*p = prop
	return err
}

func (p *PoolProp) String() string {
	return C.GoString(C.zpool_prop_to_name((C.zpool_prop_t)(*p)))
}

func (p *PoolProp) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(p.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *PoolProp) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	prop, ok := stringToPoolPropDic[j]
	if !ok {
		return fmt.Errorf("prop \"%s\" not exists", j)
	}
	*p = prop
	return err
}

//{"guid": {"value":"16859519823695578253", "source":"-"}}
func (p *DatasetProperties) MarshalJSON() ([]byte, error) {
	props := make(map[string]PropertyValue)
	maxUint64 := strconv.FormatUint(C.UINT64_MAX, 10)
	for prop, value := range *p {
		name := prop.String()
		if maxUint64 != value.Value && value.Value != "none" {
			if prop == DatasetPropCreation {
				time_int, _ := strconv.ParseInt(value.Value, 10, 64)
				value.Value = time.Unix(time_int, 0).Format("2006-01-02T15:04:05-0700")
			}
			props[name] = value
		}
	}
	data, err := json.Marshal(&props)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *DatasetProperties) UnmarshalJSON(b []byte) error {
	props := make(map[string]PropertyValue)
	err := json.Unmarshal(b, &props)
	if err != nil {
		return err
	}
	for key, value := range props {
		prop, ok := stringToDatasetPropDic[key]
		if !ok {
			return fmt.Errorf("property \"%s\" not exist", key)
		}
		(*p)[prop] = value
	}
	return nil
}

func (p *DatasetProperties) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(data)
}

func (p *PoolProperties) MarshalJSON() ([]byte, error) {
	props := make(map[string]PropertyValue)
	maxUint64 := strconv.FormatUint(C.UINT64_MAX, 10)
	for prop, value := range *p {
		name := prop.String()
		if maxUint64 != value.Value && value.Value != "none" {
			props[name] = value
		}
	}
	data, err := json.Marshal(&props)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (p *PoolProperties) UnmarshalJSON(b []byte) error {
	props := make(map[string]PropertyValue)
	err := json.Unmarshal(b, &props)
	if err != nil {
		return err
	}
	for key, value := range props {
		prop, ok := stringToPoolPropDic[key]
		if !ok {
			return fmt.Errorf("property \"%s\" not exist", key)
		}
		(*p)[prop] = value
	}
	return nil
}

func (p *PoolProperties) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(data)
}
