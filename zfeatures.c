
#include <libzfs.h>
#include <zfeature_common.h>
#include <memory.h>
#include <string.h>
#include <stdio.h>
#include "common.h"
#include "zpool.h"
#include "zfs.h"

/*
// WARN: the structure is different betwen 0.7 and 0.8
const char *zfeatures_get_name(int index) {
    return spa_feature_table[index].fi_uname;
}
*/
