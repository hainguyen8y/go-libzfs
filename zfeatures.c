
#include <libzfs.h>
#include <zfeature_common.h>
#include <memory.h>
#include <string.h>
#include <stdio.h>
#include "common.h"
#include "zpool.h"
#include "zfs.h"

const char *zfeatures_get_name(int index) {
    return spa_feature_table[index].fi_uname;
}
