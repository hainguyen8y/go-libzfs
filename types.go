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
)

var stringToPropDic = make(map[string]Prop)

func init() {
	for i := DatasetPropType; i < DatasetNumProps; i++ {
		stringToPropDic[i.String()] = i
	}
}

func (p *Prop) String() string {
	return C.GoString(C.zfs_prop_to_name((C.zfs_prop_t)(*p)))
}

func (p *Prop) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(p.String())
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (p *Prop) UnmarshalJSON(b []byte) error {
	var j string
	err := json.Unmarshal(b, &j)
	if err != nil {
		return err
	}
	prop, ok := stringToPropDic[j]
	if !ok {
		return fmt.Errorf("prop \"%s\" not exists", j)
	}
	*p = prop
	return err
}

//{"guid": {"value":"16859519823695578253", "source":"-"}}
func (p *Properties) MarshalJSON() ([]byte, error) {
	props := make(map[string]Property)
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

func (p *Properties) UnmarshalJSON(b []byte) error {
	props := make(map[string]Property)
	err := json.Unmarshal(b, &props)
	if err != nil {
		return err
	}
	for key, value := range props {
		prop, ok := stringToPropDic[key]
		if !ok {
			return fmt.Errorf("property \"%s\" not exist", key)
		}
		(*p)[prop] = value
	}
	return nil
}
