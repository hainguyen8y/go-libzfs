#include <stdio.h>
#include <stdlib.h>
#include <libzfs.h>
#include "common.h"
#include "zpool.h"
#include "zfs.h"
extern void __printf(char *key, char *val);

static int destroy_check_dependent(zfs_handle_t *zhp, void *data) {
    const char *name = zfs_get_name(zhp);
    __printf("snapshot dep", (char*)name);
    zfs_close(zhp);
    return -1;
}

int snapshot_to_nvl_cb(zfs_handle_t *zhp, void *arg)
{
	int err = 0;
	nvlist_t *pnvl = (nvlist_t*) arg;
    err = zfs_iter_dependents(zhp, B_TRUE,
		    destroy_check_dependent, NULL);
    __printf("snap", (char*)zfs_get_name(zhp));
	if (err == 0 && nvlist_add_boolean(pnvl, zfs_get_name(zhp))) {
		zfs_close(zhp);
		return -1234;
	}
    zfs_close(zhp);
	return (err);
}

void print_list(nvlist_t *pnvl) {
	nvpair_t *nvp = NULL;
	while ((nvp = nvlist_next_nvpair(pnvl, nvp)) != NULL) {
		__printf("name", nvpair_name(nvp));
	}
}
