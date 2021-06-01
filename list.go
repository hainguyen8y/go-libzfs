package zfs

// #include <stdlib.h>
// #include <libzfs.h>
// #include "common.h"
// #include "zpool.h"
// #include "zfs.h"
import "C"

import (
)

type ListOptions struct {
	Types DatasetType
	Recursive bool
	Depth int32
	Paths []string
}

func listChildren(d Dataset, opts ListOptions) (datasets []Dataset, err error) {
	var tempDatasets []Dataset
	defer DatasetCloseAll(tempDatasets)
	list := C.dataset_list_children(d.list)
	for list != nil {
		dataset := Dataset{list: list}
		dataset.Type = DatasetType(C.dataset_type(list))
		if (dataset.Type & opts.Types) == 0 {
			if opts.Types == DatasetTypeSnapshot ||  opts.Types == DatasetTypeBookmark {
				tempDatasets = append(tempDatasets, dataset)
			}
		} else {
			dataset.Properties = make(map[DatasetProp]PropertyValue)
			err = dataset.ReloadProperties()
			if err != nil {
				DatasetCloseAll(datasets)
				return
			}
			datasets = append(datasets, dataset)
		}
		list = C.dataset_next(list)
	}

	if !opts.Recursive {
		return
	}
	if opts.Depth > 0 {
		opts.Depth--
		if opts.Depth == 0 {
			opts.Recursive = false
		}
	}
	var childrenDatasets []Dataset
	for _, d := range datasets {
		var dts []Dataset
		dts, err = listChildren(d, opts)
		if err != nil {
			break
		}
		childrenDatasets = append(childrenDatasets, dts...)
	}
	if err == nil {
		for _, d := range tempDatasets {
			var dts []Dataset
			dts, err = listChildren(d, opts)
			if err != nil {
				break
			}
			childrenDatasets = append(childrenDatasets, dts...)
		}
	}
	datasets = append(datasets, childrenDatasets...)
	if err != nil {
		DatasetCloseAll(datasets)
		datasets = nil
	}
	return
}

func listRoot(opts ListOptions) (datasets []Dataset, err error) {
	var tempDatasets []Dataset
	var dataset Dataset
	defer DatasetCloseAll(tempDatasets)
	dataset.list = C.dataset_list_root()
	// Retrieve all datasets
	for dataset.list != nil {
		dataset.Type = DatasetType(C.dataset_type(dataset.list))
		if (dataset.Type & opts.Types) == 0 {
			if opts.Types == DatasetTypeSnapshot ||  opts.Types == DatasetTypeBookmark {
				tempDatasets = append(tempDatasets, dataset)
			} else {
				dataset.list = C.dataset_next(dataset.list)
				continue
			}
		} else {
			err = dataset.ReloadProperties()
			if err != nil {
				DatasetCloseAll(datasets)
				return
			}
			datasets = append(datasets, dataset)
		}
		dataset.list = C.dataset_next(dataset.list)
	}

	if !opts.Recursive {
		return
	}
	if opts.Depth > 0 {
		opts.Depth--
		if opts.Depth == 0 {
			opts.Recursive = false
		}
	}
	var childrenDatasets []Dataset
	for _, d := range datasets {
		var dts []Dataset
		dts, err = listChildren(d, opts)
		if err != nil {
			break
		}
		childrenDatasets = append(childrenDatasets, dts...)
	}
	if err == nil {
		for _, d := range tempDatasets {
			var dts []Dataset
			dts, err = listChildren(d, opts)
			if err != nil {
				break
			}
			childrenDatasets = append(childrenDatasets, dts...)
		}
	}
	datasets = append(datasets, childrenDatasets...)
	if err != nil {
		DatasetCloseAll(datasets)
		datasets = nil
	}
	return
}

func listPath(path string, opts ListOptions) (datasets []Dataset, err error) {
	var tempDatasets []Dataset
	defer DatasetCloseAll(tempDatasets)
	dataset, err := DatasetOpenSingle(path)
	if err != nil {
		return nil, err
	}
	if dataset.Type & opts.Types == 0 {
		if opts.Types == DatasetTypeSnapshot ||  opts.Types == DatasetTypeBookmark {
			tempDatasets = append(tempDatasets, dataset)
		} else {
			dataset.Close()
			return
		}
	} else {
		datasets = append(datasets, dataset)
	}

	if !opts.Recursive {
		return
	}
	if opts.Depth > 0 {
		opts.Depth--
		if opts.Depth == 0 {
			opts.Recursive = false
		}
	}
	var childrenDatasets []Dataset
	for _, d := range datasets {
		var dts []Dataset
		dts, err = listChildren(d, opts)
		if err == nil {
			childrenDatasets = append(childrenDatasets, dts...)
		} else {
			break
		}
	}
	if err == nil {
		for _, d := range tempDatasets {
			var dts []Dataset
			dts, err = listChildren(d, opts)
			if err == nil {
				childrenDatasets = append(childrenDatasets, dts...)
			} else {
				break
			}
		}
	}

	datasets = append(datasets, childrenDatasets...)
	if err != nil {
		DatasetCloseAll(datasets)
		datasets = nil
	}
	return
}

func List(opts ListOptions) (datasets []Dataset, err error) {
	if opts.Paths == nil {
		return listRoot(opts)
	}
	for _, path := range opts.Paths {
		var dts []Dataset
		dts, err = listPath(path, opts)
		if err == nil {
			datasets = append(datasets, dts...)
		} else {
			DatasetCloseAll(datasets)
			break
		}
	}
	return
}
