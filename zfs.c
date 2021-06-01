/* C wrappers around some zfs calls and C in general that should simplify
 * using libzfs from go language, make go code shorter and more readable.
 */

#include <libzfs.h>
#include <memory.h>
#include <string.h>
#include <stdio.h>

#include "common.h"
#include "zpool.h"
#include "zfs.h"


dataset_list_t *create_dataset_list_item() {
	dataset_list_t *zlist = (dataset_list_t *)malloc(sizeof(dataset_list_t));
	if (zlist == NULL) return NULL;

	memset(zlist, 0, sizeof(dataset_list_t));
	return zlist;
}

void dataset_list_close(dataset_list_t *list) {
	if (list != NULL) {
		if (list->zh != NULL) {
			zfs_close(list->zh);
			list->zh = NULL;
		}
		free(list);
	}
	// dataset_list_free(list);
}

void dataset_list_free(dataset_list_t *list) {
	dataset_list_t *next;
	while(list) {
		next = list->pnext;
		free(list);
		list = next;
	}
}

int dataset_list_callb(zfs_handle_t *dataset, void *data) {
	dataset_list_t **lroot = (dataset_list_t**)data;

	if ( (*lroot)->zh == NULL ) {
		(*lroot)->zh = dataset;
	} else {
		dataset_list_t *nroot = create_dataset_list_item();
        if (!nroot) return -1;

		nroot->zh = dataset;
		nroot->pnext = (void*)*lroot;
		*lroot = nroot;
	}
	return 0;
}

dataset_list_ptr dataset_list_root() {
	int err = 0;
	dataset_list_t *zlist = create_dataset_list_item();
    if (zlist == NULL) return NULL;

	err = zfs_iter_root(libzfs_get_handle(), dataset_list_callb, &zlist);
	if ( err != 0  || zlist->zh == NULL) {
		dataset_list_free(zlist);
		return NULL;
	}
	return zlist;
}

dataset_list_ptr dataset_next(dataset_list_t *dataset) {
	return dataset->pnext;
}

zfs_type_t dataset_type(dataset_list_ptr dataset) {
    if (dataset == NULL || dataset->zh == NULL) {
        return 0;
    }
	return zfs_get_type(dataset->zh);
}

dataset_list_ptr dataset_open(const char *path) {
	dataset_list_ptr list = create_dataset_list_item();
	zfs_type_t types = ZFS_TYPE_FILESYSTEM | ZFS_TYPE_SNAPSHOT | ZFS_TYPE_VOLUME | ZFS_TYPE_POOL | ZFS_TYPE_BOOKMARK;
	list->zh = zfs_open(libzfs_get_handle(), path, types);
	if (list->zh == NULL) {
		dataset_list_free(list);
		list = NULL;
	}
	return list;
}

int dataset_create(const char *path, zfs_type_t type, nvlist_ptr props) {
	return zfs_create(libzfs_get_handle(), path, type, props);
}

int dataset_destroy(dataset_list_ptr dataset, boolean_t defer) {
	return zfs_destroy(dataset->zh, defer);
}

dataset_list_t *dataset_list_children(dataset_list_t *dataset) {
	int err = 0;
    if (!dataset || dataset->zh == NULL) return NULL;

	dataset_list_t *zlist = create_dataset_list_item();

    if (!zlist) return NULL;

	err = zfs_iter_children(dataset->zh, dataset_list_callb, &zlist);
	if ( err != 0  || zlist->zh == NULL) {
		dataset_list_free(zlist);
		return NULL;
	}
	(void) zfs_iter_bookmarks(dataset->zh, dataset_list_callb, &zlist);
	return zlist;
}

zpool_list_ptr dataset_get_pool(dataset_list_ptr dataset) {
	zpool_list_ptr pool = create_zpool_list_item();
	if(pool != NULL) {
		pool->zph = zfs_get_pool_handle(dataset->zh);
	}
	return pool;
}

int dataset_prop_set(dataset_list_ptr dataset, zfs_prop_t prop, const char *value) {
	return zfs_prop_set(dataset->zh, zfs_prop_to_name(prop), value);
}

int dataset_user_prop_set(dataset_list_ptr dataset, const char *prop, const char *value) {
	return zfs_prop_set(dataset->zh, prop, value);
}

int dataset_clone(dataset_list_ptr dataset, const char *target, nvlist_ptr props) {
	return zfs_clone(dataset->zh, target, props);
}

int dataset_snapshot(const char *path, boolean_t recur, nvlist_ptr props) {
	return zfs_snapshot(libzfs_get_handle(), path, recur, props);
}

int dataset_rollback(dataset_list_ptr dataset, dataset_list_ptr snapshot, boolean_t force) {
	return zfs_rollback(dataset->zh, snapshot->zh, force);
}

int dataset_promote(dataset_list_ptr dataset) {
	return zfs_promote(dataset->zh);
}

int dataset_rename(dataset_list_ptr dataset, const char* new_name, boolean_t recur, boolean_t force_unm) {
	return zfs_rename(dataset->zh, new_name, recur, force_unm);
}

const char *dataset_is_mounted(dataset_list_ptr dataset){
	char *mp = NULL;
	// zfs_is_mounted returns B_TRUE or B_FALSE
	if (0 != zfs_is_mounted(dataset->zh, &mp)) {
		return mp;
	}
	return NULL;
}

int dataset_mount(dataset_list_ptr dataset, const char *options, int flags) {
	if ( 0 < strlen(options)) {
		return zfs_mount(dataset->zh, options, flags);
	} else {
		return zfs_mount(dataset->zh, NULL, flags);
	}
}

int dataset_unmount(dataset_list_ptr dataset, int flags) {
	return zfs_unmount(dataset->zh, NULL, flags);
}

int dataset_unmountall(dataset_list_ptr dataset, int flags) {
	return zfs_unmountall(dataset->zh, flags);
}

const char *dataset_get_name(dataset_list_ptr ds) {
	return zfs_get_name(ds->zh);
}

//int read_dataset_property(zfs_handle_t *zh, property_list_t *list, int prop) {
property_list_t *read_dataset_property(dataset_list_t *dataset, int prop) {
	int r = 0;
	zprop_source_t source;
	char statbuf[INT_MAX_VALUE];
	property_list_ptr list = NULL;
	list = new_property_list();

	r = zfs_prop_get(dataset->zh, prop,
		list->value, INT_MAX_VALUE, &source, statbuf, INT_MAX_VALUE, 1);
	if (r == 0 && list != NULL) {
		// strcpy(list->name, zpool_prop_to_name(prop));
		zprop_source_tostr(list->source, source);
		list->property = (int)prop;
	} else if (list != NULL) {
		free_properties(list);
		list = NULL;
	}
	return list;
}

// int read_user_property(zfs_handle_t *zh, property_list_t *list, const char *prop) {
property_list_t *read_user_property(dataset_list_t *dataset, const char* prop) {
	nvlist_t *user_props = zfs_get_user_props(dataset->zh);
	nvlist_t *propval;
	zprop_source_t sourcetype;
	char *strval;
	char *sourceval;
	char buf[ZFS_MAXPROPLEN];
	// char source[ZFS_MAX_DATASET_NAME_LEN];
	property_list_ptr list = new_property_list();

	if (zfs_prop_userquota(prop)) {
		sourcetype = ZPROP_SRC_LOCAL;
		(void) strncpy(list->source,
				"local", sizeof (list->source));
		if (zfs_prop_get_userquota(dataset->zh, prop,
			buf, sizeof (buf), B_TRUE) != 0) {
			sourcetype = ZPROP_SRC_NONE;
			strval = "-";
		} else {
			strval = buf;
		}
	}
	else if (zfs_prop_written(prop)) {
		sourcetype = ZPROP_SRC_LOCAL;
		(void) strncpy(list->source,
				"local", sizeof (list->source));
		if (zfs_prop_get_written(dataset->zh, prop,
			buf, sizeof (buf), B_TRUE) != 0) {
			sourcetype = ZPROP_SRC_NONE;
			strval = "-";
		} else {
			strval = buf;
		}
	} else if (nvlist_lookup_nvlist(user_props,
		prop, &propval) != 0) {
		sourcetype = ZPROP_SRC_NONE;
		(void) strncpy(list->source,
				"none", sizeof (list->source));
		strval = "-";
	} else {
		verify(nvlist_lookup_string(propval,
			ZPROP_VALUE, &strval) == 0);
		verify(nvlist_lookup_string(propval,
			ZPROP_SOURCE, &sourceval) == 0);

		if (strcmp(sourceval,
			zfs_get_name(dataset->zh)) == 0) {
			sourcetype = ZPROP_SRC_LOCAL;
			(void) strncpy(list->source,
				"local", sizeof (list->source));
		} else if (strcmp(sourceval,
			ZPROP_SOURCE_VAL_RECVD) == 0) {
			sourcetype = ZPROP_SRC_RECEIVED;
			(void) strncpy(list->source,
				"received", sizeof (list->source));
		} else {
			sourcetype = ZPROP_SRC_INHERITED;
			(void) strncpy(list->source,
				sourceval, sizeof (list->source));
		}
	}
	(void) strncpy(list->value,
				strval, sizeof (list->value));
	return list;
}

char** alloc_cstrings(int size) {
	return malloc(size*sizeof(char*));
}

void strings_setat(char **a, int at, char *v) {
	a[at] = v;
}


sendflags_t *alloc_sendflags() {
	sendflags_t *r = malloc(sizeof(sendflags_t));
	memset(r, 0, sizeof(sendflags_t));
	return r;
}

void sendflags_set_raw(sendflags_t * flags) {
#if LIBZFS_VERSION_MINOR == 7

#else
	flags->raw = B_TRUE;
#endif
}

recvflags_t *alloc_recvflags() {
	recvflags_t *r = malloc(sizeof(recvflags_t));
	memset(r, 0, sizeof(recvflags_t));
	return r;
}

struct zfs_cmd *new_zfs_cmd(){
	struct zfs_cmd *cmd = malloc(sizeof(struct zfs_cmd));
	memset(cmd, 0, sizeof(struct zfs_cmd));
	return cmd;
}

int estimate_send_size(struct zfs_cmd *zc) {
	int rc = zfs_ioctl(libzfs_get_handle(), ZFS_IOC_SEND, zc);
	if (rc != 0) {
		rc = errno;
	}
	return rc;
}
#if LIBZFS_VERSION_MINOR == 7
int gozfs_send_one(zfs_handle_t *zhp, const char *from, int fd, sendflags_t *flags, const char *redactbook) {
    uint32_t lzc_send_flags = 0;

    if (flags->largeblock) {
        lzc_send_flags |= LZC_SEND_FLAG_LARGE_BLOCK;
    }
    if (flags->embed_data) {
        lzc_send_flags |= LZC_SEND_FLAG_EMBED_DATA;
    }
    if (flags->compress) {
        lzc_send_flags |= LZC_SEND_FLAG_COMPRESS;
    }

    return zfs_send_one(zhp, from, fd, lzc_send_flags);
}
#else

int gozfs_send_one(zfs_handle_t *zhp, const char *from, int fd, sendflags_t *flags, const char *redactbook) {
#if LIBZFS_VERSION_PATCH != 1
	return zfs_send_one(zhp, from, fd, flags, redactbook);
#else
	return zfs_send_one(zhp, from, fd, *flags);
#endif
}
#endif
