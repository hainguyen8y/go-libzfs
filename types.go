package zfs
// #include <stdlib.h>
// #include <libzfs.h>
// #include "common.h"
// #include "zpool.h"
// #include "zfs.h"
import "C"

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	"errors"
)

var stringToDatasetPropDic = make(map[string]DatasetProp)
var stringToPoolPropDic = make(map[string]PoolProp)
var zfsMaxDatasetProp DatasetProp
var zfsMaxPoolProp PoolProp

func init() {
	if C.ZFS_NUM_PROPS > DatasetNumProps {
		zfsMaxDatasetProp = DatasetNumProps
	} else {
		zfsMaxDatasetProp = DatasetProp(C.ZFS_NUM_PROPS)
	}

	if C.ZPOOL_NUM_PROPS > PoolNumProps {
		zfsMaxPoolProp = PoolNumProps
	} else {
		zfsMaxPoolProp = PoolProp(C.ZPOOL_NUM_PROPS)
	}

	for i := DatasetPropType; i < zfsMaxDatasetProp; i++ {
		stringToDatasetPropDic[i.String()] = i
	}

	for i := PoolPropName; i < zfsMaxPoolProp; i++ {
		stringToPoolPropDic[i.String()] = i
	}
}

func (p DatasetProp) String() string {
	return C.GoString(C.zfs_prop_to_name((C.zfs_prop_t)(p)))
}

func (p DatasetProp) MarshalJSON() ([]byte, error) {
	s := p.String()
	return json.Marshal(s)
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

func (p PoolProp) String() string {
	return C.GoString(C.zpool_prop_to_name((C.zpool_prop_t)(p)))
}

func (p PoolProp) MarshalJSON() ([]byte, error) {
	s := p.String()
	return json.Marshal(s)
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
func (p DatasetProperties) MarshalJSON() ([]byte, error) {
	props := make(map[string]PropertyValue)
	maxUint64 := strconv.FormatUint(C.UINT64_MAX, 10)
	for prop, value := range p {
		name := prop.String()
		if maxUint64 != value.Value && value.Value != "none" {
			if prop == DatasetPropCreation {
				time_int, _ := strconv.ParseInt(value.Value, 10, 64)
				value.Value = time.Unix(time_int, 0).Format("2006-01-02T15:04:05-0700")
			}
			props[name] = value
		}
	}
	return json.Marshal(props)
}

func (p *DatasetProperties) UnmarshalJSON(b []byte) error {
	if p == nil {
		return errors.New("map is nil. use make")
	}
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

func (p DatasetProperties) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(data)
}

func (p PoolProperties) MarshalJSON() ([]byte, error) {
	props := make(map[string]PropertyValue)
	maxUint64 := strconv.FormatUint(C.UINT64_MAX, 10)
	for prop, value := range p {
		name := prop.String()
		if maxUint64 != value.Value && value.Value != "none" {
			props[name] = value
		}
	}
	return json.Marshal(props)
}

func (p *PoolProperties) UnmarshalJSON(b []byte) error {
	if p == nil {
		return errors.New("map is nil. use make")
	}
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

func (p PoolProperties) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return ""
	}
	return string(data)
}
