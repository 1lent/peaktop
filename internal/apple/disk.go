package apple

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation

#include <IOKit/IOKitLib.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>
#include <sys/param.h>
#include <sys/mount.h>

static void diskStatfs(const char *path, uint64_t *total, uint64_t *free) {
	struct statfs buf;
	if (statfs(path, &buf) == 0) {
		*total = (uint64_t)buf.f_blocks * (uint64_t)buf.f_bsize;
		*free  = (uint64_t)buf.f_bfree  * (uint64_t)buf.f_bsize;
	} else {
		*total = 0;
		*free  = 0;
	}
}

static CFTypeRef diskCopyProperty(io_registry_entry_t entry, const char *keyStr) {
	CFStringRef cfKey = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!cfKey) return NULL;
	CFTypeRef result = IORegistryEntrySearchCFProperty(
		entry, kIOServicePlane, cfKey, kCFAllocatorDefault, kIORegistryIterateRecursively);
	CFRelease(cfKey);
	return result;
}

static int64_t diskDictGetInt64(CFTypeRef dict, const char *keyStr) {
	CFStringRef key = CFStringCreateWithCString(kCFAllocatorDefault, keyStr, kCFStringEncodingUTF8);
	if (!key) return 0;
	int64_t value = 0;
	if (dict && CFGetTypeID(dict) == CFDictionaryGetTypeID()) {
		CFNumberRef num = (CFNumberRef)CFDictionaryGetValue((CFDictionaryRef)dict, key);
		if (num && CFGetTypeID(num) == CFNumberGetTypeID()) {
			CFNumberGetValue(num, kCFNumberSInt64Type, &value);
		}
	}
	CFRelease(key);
	return value;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const diskIOKitClass = "IOBlockStorageDriver"

func GetDiskCapacity() (totalBytes uint64, freeBytes uint64) {
	path := C.CString("/")
	defer C.free(unsafe.Pointer(path))
	var total, free C.uint64_t
	C.diskStatfs(path, &total, &free)
	return uint64(total), uint64(free)
}

func GetDiskStats() (readBytes uint64, writeBytes uint64, err error) {
	cName := C.CString(diskIOKitClass)
	defer C.free(unsafe.Pointer(cName))
	matcher := C.IOServiceMatching(cName)
	if matcher == 0 {
		return 0, 0, fmt.Errorf("IOServiceMatching %s failed", diskIOKitClass)
	}

	var iter C.io_iterator_t
	result := C.IOServiceGetMatchingServices(C.kIOMainPortDefault, C.CFDictionaryRef(matcher), &iter)
	if result != C.KERN_SUCCESS {
		return 0, 0, fmt.Errorf("IOServiceGetMatchingServices failed: %d", result)
	}
	defer C.IOObjectRelease(C.io_object_t(iter))

	statsKey := C.CString("Statistics")
	defer C.free(unsafe.Pointer(statsKey))
	readKey := C.CString("Bytes (Read)")
	defer C.free(unsafe.Pointer(readKey))
	writeKey := C.CString("Bytes (Write)")
	defer C.free(unsafe.Pointer(writeKey))

	for {
		service := C.IOIteratorNext(iter)
		if service == 0 {
			break
		}

		props := C.diskCopyProperty(C.io_registry_entry_t(service), statsKey)
		C.IOObjectRelease(service)
		if props == 0 {
			continue
		}

		readBytes += uint64(C.diskDictGetInt64(props, readKey))
		writeBytes += uint64(C.diskDictGetInt64(props, writeKey))
		C.CFRelease(props)
	}

	return readBytes, writeBytes, nil
}
