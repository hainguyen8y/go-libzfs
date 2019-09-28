/* C wrappers around some zfs calls and C in general that should simplify
 * using libzfs from go language, make go code shorter and more readable.
 */

#ifndef __COMMON_H__
#define __COMMON_H__

#ifndef LIBZFS_VERSION_MAJOR
	#define LIBZFS_VERSION_MAJOR 0
	//#warning "This automatically detects the libzfs version for compiling"
	#ifdef IMPORT_ORDER_PREFERRED_1
		#define LIBZFS_VERSION_MINOR	7
	#else
		#define LIBZFS_VERSION_MINOR	8
	#endif
#endif //LIBZFS_VERSION_MAJOR

#ifndef loff_t
	#define loff_t off_t
#endif
#define INT_MAX_NAME 256
#define INT_MAX_VALUE 1024
#define	ZAP_OLDMAXVALUELEN 1024
#define	ZFS_MAX_DATASET_NAME_LEN 256

typedef struct property_list {
	char value[INT_MAX_VALUE];
	char source[ZFS_MAX_DATASET_NAME_LEN];
	int property;
	void *pnext;
} property_list_t;

typedef struct libzfs_handle* libzfs_handle_ptr;
typedef struct nvlist* nvlist_ptr;
typedef struct property_list *property_list_ptr;
typedef struct nvpair* nvpair_ptr;
typedef struct vdev_stat* vdev_stat_ptr;
typedef char* char_ptr;

int go_libzfs_init();
libzfs_handle_t *libzfs_get_handle();

int libzfs_last_error();
const char *libzfs_last_error_str();
int libzfs_clear_last_error();
const char *libzfs_strerrno(int errcode);

property_list_t *new_property_list();
void free_properties(property_list_t *root);

nvlist_ptr new_property_nvlist();
int property_nvlist_add(nvlist_ptr ptr, const char* prop, const char *value);

int redirect_libzfs_stdout(int to);
int restore_libzfs_stdout(int saved);

#endif
